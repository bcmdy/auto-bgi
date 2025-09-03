package ScriptGroup

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/task"
	"auto-bgi/tools"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/otiai10/copy"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 读取配置组配置
func (s *ScriptGroupConfig) ReadConfig(name string) ScriptGroupConfig {
	filename := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup\\" + name + ".json"
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return ScriptGroupConfig{}
	}

	var scriptGroupConfig ScriptGroupConfig
	if err := json.Unmarshal(file, &scriptGroupConfig); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return ScriptGroupConfig{}
	}

	return scriptGroupConfig
}

// 去除指定地图追踪 json
func (s *ScriptGroupConfig) RemoveProjectsByTypeAndFolder(name string, folderName string) ScriptGroupConfig {
	// 读取配置
	configData := s.ReadConfig(name)

	// 过滤掉 Type="Pathing" && FolderName=folderName 的项目
	newProjects := make([]Project, 0, len(configData.Projects))
	for _, project := range configData.Projects {
		if !(project.Type == "Pathing" && project.FolderName == folderName) {
			newProjects = append(newProjects, project)
		}
	}
	configData.Projects = newProjects

	return configData

}

// 查询仓库指定地图追踪所有json文件
func (s *ScriptGroupConfig) listPathingJsons(repoDir string, folderName string) ([]Project, error) {

	subFolderPath, err := tools.FindJSONFiles(repoDir + "\\" + folderName)
	if err != nil {
		autoLog.Sugar.Errorf("查找子文件夹失败: %v", err)
		return nil, err
	}
	var projects []Project

	for _, filePath := range subFolderPath {
		project := Project{}

		name := filepath.Base(filePath)
		project.Name = name

		filePathAndFolderName := strings.Replace(filePath, repoDir+"\\", "", 1)
		folderName := strings.Replace(filePathAndFolderName, "\\"+name, "", 1)
		project.FolderName = folderName

		projects = append(projects, project)

	}
	return projects, nil
}

// 地图追踪更新
func (s *ScriptGroupConfig) UpdatePathing(updatePath config.UpdatePathing) (string, error) {
	repoDir := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "pathing")

	var err error
	//清空地图追踪文件夹
	err = os.RemoveAll(config.Cfg.BetterGIAddress + "\\User\\AutoPathing\\" + updatePath.FolderName)
	if err != nil {
		autoLog.Sugar.Errorf("%s清空地图追踪文件夹失败: %v", updatePath.Name, err)
	}

	//复制地图追踪文件夹
	err = copy.Copy(repoDir+"\\"+updatePath.FolderName, config.Cfg.BetterGIAddress+"\\User\\AutoPathing\\"+updatePath.FolderName)
	if err != nil {
		autoLog.Sugar.Errorf("%s复制地图追踪文件夹失败: %v", updatePath.Name, err)

	}
	//查询仓库指定地图追踪所有json文件
	listPathings, err := s.listPathingJsons(repoDir, updatePath.FolderName)
	if err != nil {
		autoLog.Sugar.Errorf("%s查询仓库指定地图追踪所有json文件失败: %v", updatePath.Name, err)

	}
	scriptGroupConfig := s.RemoveProjectsByTypeAndFolder(updatePath.Name, updatePath.FolderName)
	index := 0

	//空配置组判断
	if len(scriptGroupConfig.Projects) > 0 {
		for i := range scriptGroupConfig.Projects {
			scriptGroupConfig.Projects[i].Index = i + 1
		}
		project := scriptGroupConfig.Projects[len(scriptGroupConfig.Projects)-1]
		index = project.Index + 1
	}

	var projects []Project

	for _, pathing := range listPathings {
		pathing.Index = index
		index++
		pathing.Type = "Pathing"
		pathing.Status = "Enabled"
		pathing.Schedule = "Daily"
		pathing.RunNum = 1
		pathing.AllowJsNotification = true
		pathing.JsScriptSettingsObject = nil
		projects = append(projects, pathing)
	}

	scriptGroupConfig.Projects = append(scriptGroupConfig.Projects, projects...)

	//写入配置
	filename := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup\\" + updatePath.Name + ".json"
	data, err := json.MarshalIndent(scriptGroupConfig, "", "    ")
	if err != nil {
		autoLog.Sugar.Errorf("JSON 编码失败: %v", err)

	}
	if err = os.WriteFile(filename, data, 0644); err != nil {
		autoLog.Sugar.Errorf("%s写入文件失败: %v", updatePath.Name, err)
	}
	if err != nil {
		return "更新地图追踪失败", err
	} else {
		autoLog.Sugar.Infof("%s更新成功", updatePath.Name)
	}
	s.ListPathingUpdatePaths()

	autoLog.Sugar.Infof("通知-%s地图追踪更新成功", updatePath.Name)

	return "更新地图追踪成功", nil

}

func (s *ScriptGroupConfig) CleanAllPathing(c *gin.Context) {
	var updatePath config.UpdatePathing
	if err := c.ShouldBindJSON(&updatePath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
		return
	}
	scriptGroupConfig := s.RemoveProjectsByTypeAndFolder(updatePath.Name, updatePath.FolderName)
	for i := range scriptGroupConfig.Projects {
		scriptGroupConfig.Projects[i].Index = i + 1
	}
	//写入配置
	filename := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup\\" + updatePath.Name + ".json"
	data, err := json.MarshalIndent(scriptGroupConfig, "", "    ")
	if err != nil {
		autoLog.Sugar.Errorf("JSON 编码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if err = os.WriteFile(filename, data, 0644); err != nil {
		autoLog.Sugar.Errorf("%s写入文件失败: %v", updatePath.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	s.ListPathingUpdatePaths()
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": "清理成功"})

}

func (s *ScriptGroupConfig) ListPathingUpdatePaths() error {
	groups, err := task.ListGroups()
	if err != nil {
		autoLog.Sugar.Errorf("读取配置组失败: %v", err)
		return err
	}

	var UpdatePath []config.UpdatePathing

	for _, group := range groups {
		var UpdatePathing config.UpdatePathing
		UpdatePathing.Name = group
		scriptGroupConfig := s.ReadConfig(group)
		projects := scriptGroupConfig.Projects
		projectMap := make(map[string]string)
		for i, project := range projects {
			if project.Type == "Pathing" {
				projectMap[project.FolderName] = strconv.Itoa(i + 1)
			}
		}
		for FolderName, _ := range projectMap {
			UpdatePathing.FolderName = FolderName
			UpdatePath = append(UpdatePath, UpdatePathing)
		}

	}

	err = s.SavePathing(UpdatePath)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScriptGroupConfig) UpdatePaths(context *gin.Context) {
	err := s.ListPathingUpdatePaths()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success", "data": "ok"})
}
