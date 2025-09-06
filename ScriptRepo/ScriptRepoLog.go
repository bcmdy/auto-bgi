package ScriptRepo

import (
	"auto-bgi/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

type ScriptRepoLog struct {
	Time    string    `json:"time"`
	Indexes []Indexes `json:"indexes"`
}

type Indexes struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Version     string    `json:"version"`
	Authors     []Authors `json:"authors"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	LastUpdated string    `json:"lastUpdated"`
	Children    []Indexes `json:"children"`
}

type Authors struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type Repos struct {
	TypeName string
	Repo     []Repo
}

type Repo struct {
	FilePath    string
	Authors     string
	LastUpdated string
	Tags        string
	Version     string
	Description string
}

func Read() []Repos {
	filePath := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo.json")

	// 读取文件
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return []Repos{}
	}

	// 解析 JSON
	var repoLog ScriptRepoLog
	if err := json.Unmarshal(content, &repoLog); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return []Repos{}
	}

	repoLog.Indexes = filterIndexesByLast3Days(repoLog.Indexes)

	// 转换为 []Repos
	repos := ConvertToRepos(repoLog.Indexes)

	// 打印结果
	//for _, r := range repos {
	//	fmt.Println("TypeName:", r.TypeName)
	//	for _, rep := range r.Repo {
	//		fmt.Printf("  FilePath: %s,  Authors: %s, LastUpdated: %s, Version: %s, Tags: %s\n",
	//			rep.FilePath, rep.Authors, rep.LastUpdated, rep.Version, rep.Tags)
	//	}
	//}
	return repos
}

// 递归过滤 Indexes，只保留最近三天更新的节点
func filterIndexesByLast3Days(indexes []Indexes) []Indexes {
	var result []Indexes
	now := time.Now()
	threeDaysAgo := now.AddDate(0, 0, -3)

	for _, idx := range indexes {
		// 递归过滤子节点
		idx.Children = filterIndexesByLast3Days(idx.Children)

		// 判断 LastUpdated 是否在最近三天
		include := false
		if idx.LastUpdated != "" {
			t, err := time.Parse("2006-01-02 15:04:05", idx.LastUpdated)
			if err == nil && t.After(threeDaysAgo) && t.Before(now.Add(time.Second)) {
				include = true
			}
		}

		// 如果自己或子节点满足条件，保留
		if include || len(idx.Children) > 0 {
			result = append(result, idx)
		}
	}
	return result
}

// 将 Indexes 树转换为 []Repos
func ConvertToRepos(indexes []Indexes) []Repos {
	var result []Repos
	for _, idx := range indexes {
		repoList := []Repo{}

		flattenIndexFiles(idx, "", &repoList)

		if len(repoList) > 0 {
			result = append(result, Repos{
				TypeName: idx.Name,
				Repo:     repoList,
			})
		}
	}
	return result
}

// 判断一个字符串是否是时间
func IsValidTime(timeStr string) bool {

	_, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return false
	}

	return true
}

// 只保留文件类型的递归
func flattenIndexFiles(idx Indexes, parentPath string, repoList *[]Repo) {
	currentPath := idx.Name
	if parentPath != "" {
		currentPath = parentPath + "/" + idx.Name
	}

	if IsValidTime(idx.LastUpdated) {
		*repoList = append(*repoList, Repo{
			FilePath:    currentPath,
			Authors:     joinAuthors(idx.Authors),
			LastUpdated: idx.LastUpdated,
			Version:     idx.Version,
			Tags:        strings.Join(idx.Tags, ", "),
			Description: idx.Description,
		})
	}

	for _, child := range idx.Children {
		flattenIndexFiles(child, currentPath, repoList)
	}
}

// 拼接作者
func joinAuthors(authors []Authors) string {
	names := []string{}
	for _, a := range authors {
		name := strings.TrimSpace(a.Name)
		if name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ", ")
}
