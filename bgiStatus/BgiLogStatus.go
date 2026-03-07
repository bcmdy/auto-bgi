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
	"time"
)

type bgiLogStatus struct {
	Group                               string // 配置组名称
	Timestamp                           string //时间
	MapTrackingLine                     string // 地图追踪路线
	ScriptName                          string // 脚本名称
	ConfigurationGroupExecutionProgress string // 配置组执行进度
	JSProgress                          string // js进度
	Running                             bool   // 是否运行中
	GroupProgress                       string //一条龙进度
}

var BgiLogStatusInfo bgiLogStatus
var Projects []ScriptGroup.Project

var Group ScriptGroup.ScriptGroupConfig

var BgiGroupEnd = map[string]string{
	"自动秘境结束":               "自动秘境",
	"领取尘歌壶奖励:\"退出到主页\"":    "领取尘歌壶奖励",
	"合成\"浓缩树脂\"":           "合成树脂",
	"→ \"前往冒险家协会领取奖励\" 结束": "领取每日奖励",
	"邮件：\"全部领取\"":          "领取邮件",
}

var RunDetail BGIGroupRunDetail

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

		//一条龙初始化
		if strings.Contains(line, "参数指定的一条龙配置：") {
			name := strings.ReplaceAll(line, "参数指定的一条龙配置：", "")
			//去除空格
			name = strings.TrimSpace(name)
			InitialOneLongProgress(name)
		}

		//找到配置组
		if strings.Contains(line, "配置组") && strings.Contains(line, "开始执行") {
			re := regexp.MustCompile(`"(.*?)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				BgiLogStatusInfo.Group = matches[1]

				RunDetail = BgiLogTime(lastLine)
				data := "【时间：" + RunDetail.StartTime.Format("2006-01-02 15:04:05") + "】\n" +
					"【时长：" + RunDetail.LastRunDuration + "】\n" +
					"【已经运行：" + RunDetail.AlreadyRunTime + "】\n" +
					"【预计：" + RunDetail.ExpectedEndTime.Format("2006-01-02 15:04:05") + "】"
				BgiLogStatusInfo.Timestamp = data

				BgiLogStatusInfo.MapTrackingLine = ""
				BgiLogStatusInfo.ScriptName = ""
				BgiLogStatusInfo.JSProgress = ""
			} else {
				BgiLogStatusInfo.Group = "未找到配置组"
			}
		}

		if BgiGroupEnd[line] != "" {
			if _, ok := OneLongProgress.Details[BgiGroupEnd[line]]; ok {
				OneLongProgress.Details[BgiGroupEnd[line]] = true
			}
		}

		if strings.Contains(line, "配置组") && strings.Contains(line, "执行结束") {

			re := regexp.MustCompile(`"(.*?)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				BgiLogStatusInfo.Group = matches[1] + "==(已经结束)"

				BgiLogStatusInfo.MapTrackingLine = "已经结束"
				BgiLogStatusInfo.ScriptName = "已经结束"
				BgiLogStatusInfo.JSProgress = "已经结束"

				//更新进度
				if _, ok := OneLongProgress.Details[matches[1]]; ok {

					//判断是否是正常结束的
					endTime, err := BgiLastLine(lastLine)
					if err != nil {
						autoLog.Sugar.Errorf("获取结束时间失败: %v", err)
						continue
					}
					//判断是否是正常结束的
					if endTime.Before(RunDetail.ExpectedEndTime) {
						autoLog.Sugar.Errorf("配置组 %s 异常结束，结束时间 %s 早于预期结束时间 %s", matches[1], endTime.Format("2006-01-02 15:04:05"), RunDetail.ExpectedEndTime.Format("2006-01-02 15:04:05"))
						continue
					} else {
						autoLog.Sugar.Infof("配置组 %s 正常结束，结束时间 %s 晚于预期结束时间 %s", matches[1], endTime.Format("2006-01-02 15:04:05"), RunDetail.ExpectedEndTime.Format("2006-01-02 15:04:05"))
						OneLongProgress.Details[matches[1]] = true
					}

				}

			} else {
				BgiLogStatusInfo.Group = "未找到配置组"
			}
		}

		//查找配置组任务执行
		if strings.Contains(line, "配置组任务执行: ") {
			gp := strings.ReplaceAll(line, "配置组任务执行: ", "")
			BgiLogStatusInfo.GroupProgress = gp
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

// bgi时间计算
func BgiLastLine(lastLine string) (time.Time, error) {

	// 匹配格式 [HH:MM:SS.mmm]
	re := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})]`)
	timeMatch := re.FindStringSubmatch(lastLine)

	// 解析文件名中的日期
	fileDate := GetFileNameDate(config.Cfg.BgiLog)

	// 解析起始时间，例如 09:06:24.391
	start, err := time.Parse("2006-01-02 15:04:05", fileDate+" "+timeMatch[1])
	if err != nil {
		autoLog.Sugar.Errorf("解析时间失败: %v", err)
		return time.Time{}, err
	}
	// 返回解析后的时间
	return start, nil
}

// 配置组运行情况分析
func BgiLogTime(lastLine string) BGIGroupRunDetail {
	// 匹配格式 [HH:MM:SS.mmm]
	re := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})]`)
	timeMatch := re.FindStringSubmatch(lastLine)
	if len(timeMatch) > 1 {
		time, err := CalculateTime(config.Cfg.BgiLog, BgiLogStatusInfo.Group, timeMatch[1])
		if err != nil {
			return BGIGroupRunDetail{}
		} else {
			return time

		}
	}
	return BGIGroupRunDetail{}
}
