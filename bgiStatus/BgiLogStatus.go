package bgiStatus

import (
	"auto-bgi/ScriptGroup"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type bgiLogStatus struct {
	Group                               string // 配置组名称
	Timestamp                           string //时间
	MapTrackingLine                     string // 地图追踪路线
	ScriptName                          string // 脚本名称
	ConfigurationGroupExecutionProgress string // 配置组执行进度
	JSProgress                          string // js进度
	Running                             bool   // 是否运行中
}

var BgiLogStatusInfo bgiLogStatus
var Projects []ScriptGroup.Project

var Group ScriptGroup.ScriptGroupConfig

// 倒序读取日志，找到第一个配置组和地图追踪路线就返回
func InitBgiLogStatus() {
	logName := config.Cfg.BgiLog
	if logName == "" {
		autoLog.Sugar.Infof("bgi日志名称为空")
		return
	}
	//filePath := filepath.Join(config.Cfg.BetterGIAddress, "log", logName)

	file, err := os.Open(logName)
	if err != nil {
		autoLog.Sugar.Errorf("打开bgi日志文件失败: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var lastLine string

	for scanner.Scan() {
		line := scanner.Text()

		//找到配置组
		if strings.Contains(line, "配置组") && strings.Contains(line, "开始执行") {
			re := regexp.MustCompile(`"(.*?)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				BgiLogStatusInfo.Group = matches[1]

				//提取开始时间
				BgiLogStatusInfo.Timestamp = BgiLogTime(lastLine)

				BgiLogStatusInfo.MapTrackingLine = ""
				BgiLogStatusInfo.ScriptName = ""
				BgiLogStatusInfo.JSProgress = ""
			} else {
				BgiLogStatusInfo.Group = "未找到配置组"
			}
		}

		//当前运行脚本
		if strings.Contains(line, "开始执行JS脚本") {
			re := regexp.MustCompile(`"(.*?)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				BgiLogStatusInfo.ScriptName = matches[1]
			} else {
				BgiLogStatusInfo.ScriptName = "未找到脚本名称"
			}
		}

		//当前运行路线
		if strings.Contains(line, "开始执行地图追踪任务") {
			re := regexp.MustCompile(`"(.*?)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				BgiLogStatusInfo.MapTrackingLine = matches[1]
			} else {
				BgiLogStatusInfo.MapTrackingLine = "未找到地图追踪路线"
			}
		}

		//js进度
		if strings.Contains(line, "当前进度：") || strings.Contains(line, "当前次数：") {
			BgiLogStatusInfo.JSProgress = line
		}

		lastLine = line
	}

	readConfig := Group.ReadConfig(BgiLogStatusInfo.Group)
	Projects = readConfig.Projects
	index := GetProjectIndex(BgiLogStatusInfo.ScriptName)
	index2 := GetProjectIndex(BgiLogStatusInfo.MapTrackingLine)
	if index2 != 0 {
		index = index2
	}
	BgiLogStatusInfo.ConfigurationGroupExecutionProgress = fmt.Sprintf("%d/%d", index, len(Projects))

}

// 根据路径名称获取排序
func GetProjectIndex(name string) int {
	for _, project := range Projects {
		if project.Name == name {
			if project.Type == "Pathing" {
				BgiLogStatusInfo.ScriptName = ""
			}
			return project.Index
		}
	}
	return 0
}

// bgi时间提取
func BgiLogTime(lastLine string) string {
	// 匹配格式 [HH:MM:SS.mmm]
	re := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})]`)
	timeMatch := re.FindStringSubmatch(lastLine)
	if len(timeMatch) > 1 {
		time, err := CalculateTime(config.Cfg.BgiLog, BgiLogStatusInfo.Group, timeMatch[1])
		if err != nil {
			return "没有归档记录"
		} else {
			return time
		}
	}
	return "没有归档记录"
}
