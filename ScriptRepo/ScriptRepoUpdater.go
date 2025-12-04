package ScriptRepo

import (
	myConfig "auto-bgi/config"
	"encoding/json"
	"errors"
	"fmt"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	otiai10Copy "github.com/otiai10/copy"
	"golang.org/x/term"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var proxyOptions = transport.ProxyOptions{}

func init() {
	if myConfig.Cfg.Notice.TGNotice.Proxy != "" {
		proxyOptions.URL = myConfig.Cfg.Notice.TGNotice.Proxy
	}

}

// UpdateCenterRepoByGit 克隆或同步仓库到本地路径，并返回本地路径与是否更新过的标志。
// 注意：依赖外部变量 proxyOptions, myConfig等（与原实现一致）。
func UpdateCenterRepoByGit(repoUrl string) (string, bool, error) {
	if repoUrl == "" {
		return "", false, errors.New("仓库 URL 不能为空")
	}

	reposPath := filepath.Join(myConfig.Cfg.BetterGIAddress, "Repos")
	repoPath := filepath.Join(reposPath, "bettergi-scripts-list-git")
	updated := false

	// 确保父目录存在
	if err := os.MkdirAll(reposPath, 0o755); err != nil {
		return "", false, fmt.Errorf("创建目录失败: %w", err)
	}

	// 决定是否把 Stdout 作为 Progress 写入目标（仅当 Stdout 是交互式终端时）
	var progress io.Writer = nil
	// 如果你更愿意在非交互时丢弃而不是 nil，可以设为 io.Discard
	if term.IsTerminal(int(os.Stdout.Fd())) {
		// 只有在真实终端才传 os.Stdout，防止在 Windows 服务或 CI 中出现 "The handle is invalid" 错误
		progress = os.Stdout
	} else {
		// 非交互时不输出进度，nil 更加保守（不会写入）
		progress = nil
	}

	// 仓库不存在 -> 完整克隆
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Printf("完整克隆仓库: %s 到 %s", repoUrl, repoPath)
		_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:          repoUrl,
			SingleBranch: false, // 拉取所有分支，避免缺分支
			Progress:     progress,
			ProxyOptions: proxyOptions,
		})
		if err != nil {
			return "", false, fmt.Errorf("克隆仓库失败: %w", err)
		}
		return repoPath, true, nil
	} else if err != nil {
		// 其它 Stat 错误
		return "", false, fmt.Errorf("检查仓库路径失败: %w", err)
	}

	// 打开已有仓库
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", false, fmt.Errorf("打开仓库失败: %w", err)
	}

	// 确保远程 origin 存在并且 URL 正确
	remote, err := r.Remote("origin")
	if err != nil {
		if err == git.ErrRemoteNotFound {
			_, err = r.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{repoUrl},
			})
			if err != nil {
				return "", false, fmt.Errorf("创建远程 origin 失败: %w", err)
			}
			remote, _ = r.Remote("origin")
		} else {
			return "", false, fmt.Errorf("获取远程失败: %w", err)
		}
	}

	remoteURLs := remote.Config().URLs
	if len(remoteURLs) == 0 || remoteURLs[0] != repoUrl {
		log.Printf("更新远程 URL: 从 %v 到 %s", remoteURLs, repoUrl)
		if err := r.DeleteRemote("origin"); err != nil {
			log.Printf("删除远程 origin 失败（继续尝试创建新 remote）: %v", err)
		}
		if _, err := r.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{repoUrl},
		}); err != nil {
			return "", false, fmt.Errorf("创建/更新远程 origin 失败: %w", err)
		}
	}

	// Fetch 更新远程引用
	fetchErr := r.Fetch(&git.FetchOptions{
		RemoteName:   "origin",
		Progress:     progress,
		Force:        true,
		Prune:        true,
		ProxyOptions: proxyOptions,
	})
	if fetchErr != nil && fetchErr != git.NoErrAlreadyUpToDate {
		return "", false, fmt.Errorf("fetch 失败: %w", fetchErr)
	}

	// 优先尝试 origin/main，然后 origin/release
	tryBranches := []string{"main", "release"}
	var remoteRef *plumbing.Reference
	var chosenRemoteRefName string
	for _, b := range tryBranches {
		remoteBranch := plumbing.NewRemoteReferenceName("origin", b)
		ref, err := r.Reference(remoteBranch, true)
		if err == nil {
			remoteRef = ref
			chosenRemoteRefName = remoteBranch.String()
			break
		}
	}
	if remoteRef == nil {
		return "", false, errors.New("获取远程 main/release 分支引用失败（远程不存在这些分支）")
	}

	// 比较当前 HEAD 与 remoteRef：相同则无需 reset，从而 updated=false
	headRef, err := r.Head()
	if err != nil {
		log.Printf("获取本地 HEAD 失败（将继续 reset）: %v", err)
	} else {
		if headRef.Hash() == remoteRef.Hash() {
			log.Printf("本地已与 %s 同步，无需更新。hash=%s", chosenRemoteRefName, remoteRef.Hash().String())
			return repoPath, false, nil
		}
	}

	// 执行硬 reset 到远程引用
	worktree, err := r.Worktree()
	if err != nil {
		return "", false, fmt.Errorf("获取工作树失败: %w", err)
	}
	if err := worktree.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: remoteRef.Hash(),
	}); err != nil {
		return "", false, fmt.Errorf("强制 reset 到 %s(%s) 失败: %w", chosenRemoteRefName, remoteRef.Hash().String(), err)
	}

	updated = true
	log.Printf("同步到远程 %s: %s", chosenRemoteRefName, remoteRef.Hash().String())

	return repoPath, updated, nil
}

// addUpdateMarkersToNewRepo 比较旧 JSON 和新 JSON，标记 hasUpdate
func addUpdateMarkersToNewRepo(oldContent, newContent string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("标记repo.json更新失败（panic），返回原内容: %v", r)
		}
	}()

	var oldJson, newJson map[string]interface{}

	if err := json.Unmarshal([]byte(oldContent), &oldJson); err != nil {
		oldJson = make(map[string]interface{})
	}
	if err := json.Unmarshal([]byte(newContent), &newJson); err != nil {
		return newContent
	}

	oldIndexes, okOld := oldJson["indexes"].([]interface{})
	newIndexes, okNew := newJson["indexes"].([]interface{})

	if okOld && okNew {
		for _, ni := range newIndexes {
			if newIndexObj, ok := ni.(map[string]interface{}); ok {
				markNodeUpdates(newIndexObj, oldIndexes)
			}
		}
	}

	result, err := json.MarshalIndent(newJson, "", "  ")
	if err != nil {
		return newContent
	}
	return string(result)
}

// markNodeUpdates 递归标记更新
func markNodeUpdates(newNode map[string]interface{}, oldNodes []interface{}) bool {
	newName, _ := newNode["name"].(string)
	if newName == "" {
		return false
	}

	var oldNode map[string]interface{}
	for _, n := range oldNodes {
		if obj, ok := n.(map[string]interface{}); ok {
			if name, _ := obj["name"].(string); name == newName {
				oldNode = obj
				break
			}
		}
	}

	hasDirectUpdate := false
	hasChildUpdate := false

	if oldNode != nil {
		oldTime := parseLastUpdated(oldNode["lastUpdated"])
		newTime := parseLastUpdated(newNode["lastUpdated"])
		if newTime.After(oldTime) {
			newNode["hasUpdate"] = true
			hasDirectUpdate = true
		}
	} else {
		newNode["hasUpdate"] = true
		hasDirectUpdate = true
	}

	if children, ok := newNode["children"].([]interface{}); ok && len(children) > 0 {
		var oldChildren []interface{}
		if oldNode != nil {
			if oc, ok := oldNode["children"].([]interface{}); ok {
				oldChildren = oc
			}
		}

		for _, child := range children {
			if childObj, ok := child.(map[string]interface{}); ok {
				childHasUpdate := markNodeUpdates(childObj, oldChildren)
				if childHasUpdate {
					hasChildUpdate = true
					if _, exists := childObj["children"]; !exists && !hasDirectUpdate && childObj["hasUpdate"] != nil {
						newNode["hasUpdate"] = true
						hasDirectUpdate = true
					}
				}
			}
		}
	}

	return hasDirectUpdate || hasChildUpdate
}

// parseLastUpdated 解析时间
func parseLastUpdated(v interface{}) time.Time {
	if s, ok := v.(string); ok && s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			return t
		}
		t2, err2 := time.Parse("2006-01-02 15:04:05", s)
		if err2 == nil {
			return t2
		}
	}
	return time.Time{}
}

// Subscribe 订阅脚本
func SubscribeScript(ScriptName string) (string, error) {
	ReposScriptPath := filepath.Join(myConfig.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "js", ScriptName)

	//复制user
	UserScriptPath := filepath.Join(myConfig.Cfg.BetterGIAddress, "User", "JsScript", ScriptName)

	err := otiai10Copy.Copy(ReposScriptPath, UserScriptPath)
	if err != nil {
		return "订阅失败", err
	}
	return "订阅成功", nil

}
