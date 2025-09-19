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
	"log"
	"os"
	"path/filepath"
	"time"
)

var proxyOptions = transport.ProxyOptions{}

func init() {
	if myConfig.Cfg.Notice.Type == "TG" {
		proxyOptions.URL = myConfig.Cfg.Notice.TGNotice.Proxy
	}

}

// UpdateCenterRepoByGit 强制同步 main 分支并标记更新
func UpdateCenterRepoByGit(repoUrl string) (string, bool, error) {
	if repoUrl == "" {
		return "", false, errors.New("仓库URL不能为空")
	}

	reposPath := filepath.Join(myConfig.Cfg.BetterGIAddress, "Repos")
	repoPath := filepath.Join(reposPath, "bettergi-scripts-list-git")
	updated := false

	// 仓库不存在 -> 完整克隆
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Printf("完整克隆仓库: %s 到 %s", repoUrl, repoPath)
		_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:          repoUrl,
			SingleBranch: true,
			Progress:     os.Stdout,
			ProxyOptions: proxyOptions,
		})
		if err != nil {
			return "", false, fmt.Errorf("克隆仓库失败: %w", err)
		}
		updated = true
	} else {

		// 打开仓库
		r, err := git.PlainOpen(repoPath)
		if err != nil {
			return "", false, fmt.Errorf("打开仓库失败: %w", err)
		}

		// 确保远程 URL 正确
		remote, err := r.Remote("origin")
		if err != nil {
			return "", false, fmt.Errorf("获取远程失败: %w", err)
		}
		if remote.Config().URLs[0] != repoUrl {
			log.Printf("更新远程URL: 从 %s 到 %s", remote.Config().URLs[0], repoUrl)
			_ = r.DeleteRemote("origin")
			_, _ = r.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{repoUrl},
			})
		}

		// 强制 fetch 更新远程引用
		err = r.Fetch(&git.FetchOptions{
			RemoteName:   "origin",
			Progress:     os.Stdout,
			Force:        true,
			ProxyOptions: proxyOptions,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", false, fmt.Errorf("fetch 失败: %w", err)
		}

		// 获取远程 main 分支
		remoteBranch := plumbing.NewRemoteReferenceName("origin", "main")
		remoteRef, err := r.Reference(remoteBranch, true)
		if err != nil {
			return "", false, fmt.Errorf("获取远程 main 分支引用失败: %w", err)
		}

		// 强制 reset --hard 到远程 main
		worktree, _ := r.Worktree()
		err = worktree.Reset(&git.ResetOptions{
			Mode:   git.HardReset,
			Commit: remoteRef.Hash(),
		})
		if err != nil {
			return "", false, fmt.Errorf("强制 reset 到 main 失败: %w", err)
		}

		updated = true
		log.Printf("同步到远程 main: %s", remoteRef.Hash().String())
	}

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
