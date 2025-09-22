package OneLong

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/task"
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// changeTaskEnabledList 修改指定一条龙的配置文件中的 TaskEnabledList 字段
func (o *OneLong) changeTaskEnabledList(longName string) error {

	now := time.Now()
	weekdayNum := int(now.Weekday())

	autoLog.Sugar.Infof("今天是: 星期%d", weekdayNum)

	//自定义配置路径
	filename := config.Cfg.BetterGIAddress + "\\User\\OneDragon\\" + longName + ".json"

	// 1. 读取 JSON 文件
	data, err := os.ReadFile(filename)
	if err != nil {
		autoLog.Sugar.Errorf("一条龙读取文件失败%s: %v", longName, err)
		return err
	}

	//2. 解析为 orderedData
	jsonData := orderedmap.New()
	if err := json.Unmarshal(data, &jsonData); err != nil {

		autoLog.Sugar.Errorf("解析 JSON 失败: %v", err)
		return err
	}
	_, b2 := jsonData.Get("SelectedPeriodList")
	if !b2 {
		autoLog.Sugar.Errorf("SelectedPeriodList 字段不存在")
	} else {
		autoLog.Sugar.Infof("SelectedPeriodList 字段存在")
		task.ReadChaBaoBgiConfig(filename)
		return nil
	}

	TaskEnabled, b := jsonData.Get("TaskEnabledList")
	if !b {
		autoLog.Sugar.Errorf("TaskEnabledList 字段不存在")
		return fmt.Errorf("TaskEnabledList 字段不存在")
	}

	aa := TaskEnabled.(orderedmap.OrderedMap)
	re := regexp.MustCompile(`\d+`) // 匹配一个或多个连续数字
	var builder strings.Builder

	var oneLongGroup []string

	builder.WriteString("今日执行一条龙：" + longName + "\n")
	builder.WriteString("今日执行配置组：")
	builder.WriteString("\n")

	var oneLongLog strings.Builder

	for _, s := range aa.Keys() {

		autoLog.Sugar.Infof("配置组:%s", s)
		numbers := re.FindAllString(s, -1)
		if numbers == nil {
			get, _ := aa.Get(s)

			if get == true {
				builder.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
				builder.WriteString("\n")

				oneLongLog.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
				oneLongLog.WriteString("\n")

				oneLongGroup = append(oneLongGroup, s)

				continue
			}
			continue
		}
		autoLog.Sugar.Infof("匹配的数字:%v", numbers)
		if contains(numbers, weekdayNum) {
			autoLog.Sugar.Infof("配置组:[" + s + "]已到执行时间")
			aa.Set(s, true)
			//builder.WriteString(fmt.Sprintf("%s：%v", s, true))
			builder.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
			builder.WriteString("\n")

			oneLongLog.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
			oneLongLog.WriteString("\n")

			oneLongGroup = append(oneLongGroup, s)
			continue
		} else {
			autoLog.Sugar.Infof("配置组:[" + s + "]还未到执行时间")
			aa.Set(s, false)
			continue
		}
	}

	updatedData, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 编码失败")
	}

	// 6. 写回文件
	if err := os.WriteFile(filename, updatedData, 0644); err != nil {

		autoLog.Sugar.Errorf("写入文件失败: %v", err)
		return fmt.Errorf("自定义配置写入文件失败")
	}

	//发送通知
	Notice.SentText(builder.String())

	//计算一条龙时间
	go func() {
		bgiStatus.GetTodayOneLongTime(oneLongGroup)
	}()

	return nil

}

// 使用循环遍历检查数字是否包含在数组中
func contains(slice []string, num int) bool {
	for _, v := range slice {
		vInt, _ := strconv.Atoi(v)
		if vInt == num {
			return true
		}
	}
	return false
}
