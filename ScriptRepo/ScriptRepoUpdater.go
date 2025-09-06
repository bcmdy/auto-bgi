package ScriptRepo

import (
	"auto-bgi/autoLog"
	myConfig "auto-bgi/config"
	"encoding/json"
	"errors"
	"fmt"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func UpdateCenterRepoByGit(repoUrl string) (string, bool, error) {
	if repoUrl == "" {
		return "", false, errors.New("仓库URL不能为空")
	}

	reposPath := myConfig.Cfg.BetterGIAddress + "\\Repos"
	repoPath := filepath.Join(reposPath, "bettergi-scripts-list-git")
	updated := false

	var oldRepoJsonContent []byte

	// 包装整个逻辑
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// 仓库不存在，执行克隆
		log.Printf("浅克隆仓库: %s 到 %s", repoUrl, repoPath)
		_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:           repoUrl,
			Depth:         1,
			SingleBranch:  true,
			ReferenceName: plumbing.NewBranchReferenceName("main"),
			Progress:      os.Stdout,
		})
		if err != nil {
			return "", false, fmt.Errorf("克隆仓库失败: %w", err)
		}
		updated = true
	} else {
		// 仓库已存在，检查 repo.json 备份
		oldRepoJsonPath := ""
		filepath.WalkDir(reposPath, func(path string, d fs.DirEntry, err error) error {
			if err == nil && d.Name() == "repo.json" {
				oldRepoJsonPath = path
				return fs.SkipAll
			}
			return nil
		})
		if oldRepoJsonPath != "" {
			oldRepoJsonContent, _ = ioutil.ReadFile(oldRepoJsonPath)
		}

		// 打开仓库
		r, err := git.PlainOpen(repoPath)
		if err != nil {
			return "", false, fmt.Errorf("打开仓库失败: %w", err)
		}

		// 检查远程 URL
		remote, err := r.Remote("origin")
		if err != nil {
			return "", false, fmt.Errorf("获取远程失败: %w", err)
		}

		if remote.Config().URLs[0] != repoUrl {
			log.Printf("更新远程URL: 从 %s 到 %s", remote.Config().URLs[0], repoUrl)
			remoteCfg := &config.RemoteConfig{
				Name: "origin",
				URLs: []string{repoUrl},
			}
			r.DeleteRemote("origin")
			r.CreateRemote(remoteCfg)
		}

		// 拉取最新代码
		worktree, err := r.Worktree()
		if err != nil {
			return "", false, fmt.Errorf("获取工作区失败: %w", err)
		}

		headRef, _ := r.Head()
		oldCommit := headRef.Hash().String()

		err = worktree.Pull(&git.PullOptions{
			RemoteName:    "origin",
			SingleBranch:  true,
			ReferenceName: plumbing.NewBranchReferenceName("main"),
			Progress:      os.Stdout,
		})

		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", false, fmt.Errorf("拉取更新失败: %w", err)
		}

		newHeadRef, _ := r.Head()
		updated = oldCommit != newHeadRef.Hash().String()
	}

	// 如果有更新，处理 repo.json 差异
	if updated {
		newRepoJsonPath := ""
		filepath.WalkDir(reposPath, func(path string, d fs.DirEntry, err error) error {
			if err == nil && d.Name() == "repo.json" {
				newRepoJsonPath = path
				return fs.SkipAll
			}
			return nil
		})

		if newRepoJsonPath != "" {
			newRepoJsonContent, _ := ioutil.ReadFile(newRepoJsonPath)

			// 检查 repo_update.json
			parentDir := filepath.Dir(repoPath)
			repoUpdateJsonPath := filepath.Join(parentDir, "repo_update.json")
			var updatedContent string

			if _, err := os.Stat(repoUpdateJsonPath); err == nil {
				repoUpdateContent, _ := ioutil.ReadFile(repoUpdateJsonPath)
				updatedContent = addUpdateMarkersToNewRepo(string(repoUpdateContent), string(newRepoJsonContent))
			} else {
				updatedContent = addUpdateMarkersToNewRepo(string(oldRepoJsonContent), string(newRepoJsonContent))
			}

			updatedRepoJsonPath := filepath.Join(parentDir, "repo_updated.json")
			_ = ioutil.WriteFile(updatedRepoJsonPath, []byte(updatedContent), 0644)
			autoLog.Sugar.Infof("repo.json更新内容:%s", updatedContent)
		}
	}

	return repoPath, updated, nil
}

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

	// 在老版节点中找同名节点
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

	// 检查本节点是否更新
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

	// 递归处理 children
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

					// 如果子节点是叶子节点并且父节点未标记更新，则标记
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

// parseLastUpdated 解析时间戳，失败返回零值
func parseLastUpdated(v interface{}) time.Time {
	if s, ok := v.(string); ok && s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			return t
		}
		// 如果不是标准格式，可以尝试其他格式
		// 比如 2024-06-01 12:34:56
		t2, err2 := time.Parse("2006-01-02 15:04:05", s)
		if err2 == nil {
			return t2
		}
	}
	return time.Time{}
}
