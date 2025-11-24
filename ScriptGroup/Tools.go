package ScriptGroup

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

/**
 * 列出所有脚本组
 * @return []string 脚本组名称列表
 * @return error 错误信息，如果发生错误则返回非nil值
 */
func listGroups() ([]string, error) {
	// 指定要读取的文件夹路径
	//自定义配置路径
	folderPath := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup" // 构建脚本组文件夹的完整路径

	var groupNames []string // 用于存储找到的脚本组名称

	// 遍历文件夹
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // 如果遍历过程中出现错误，直接返回错误
		}

		// 检查是否是 JSON 文件
		if filepath.Ext(d.Name()) == ".json" { // 检查文件扩展名是否为.json
			// 打印 JSON 文件名（相对于文件夹的路径）
			relativePath, err := filepath.Rel(folderPath, path) // 获取相对于文件夹路径的相对路径
			if err != nil {
				return err // 如果获取相对路径失败，返回错误
			}

			name := strings.Replace(relativePath, ".json", "", -1) // 移除.json扩展名，得到组名

			groupNames = append(groupNames, name) // 将组名添加到结果列表中

		}

		return nil // 继续遍历
	})

	if err != nil {
		return nil, err // 如果遍历过程中出现错误，返回空列表和错误
	}

	return groupNames, nil // 返回找到的所有组名列表和nil错误
}

// StartGroups 启动配置组
func startGroups(names []string) error {
	control.CloseSoftware()
	time.Sleep(1 * time.Second)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")

	// 检查文件是否存在
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return err
	}

	args := append([]string{"--startGroups"}, names...) // 每个组名单独参数

	cmd := exec.Command(betterGIPath, args...)

	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	if err != nil {
		autoLog.Sugar.Errorf("启动配置组失败: %v", err)
		return err
	}
	autoLog.Sugar.Infof("启动配置组成功: %v", names)
	return nil
}

// CalculateExecuteOrder 计算当天执行序号
// now: 当前时间
// splitHour: 分界时间 (0 或 4)
// cycle: 周期天数 (>=1)
// startOrder: 起始执行序号 (>=1)
func calculateExecuteOrder(now time.Time, splitHour, cycle, startOrder int) int {
	if cycle <= 0 {
		return -1 // 非法输入
	}
	if startOrder <= 0 {
		startOrder = 1
	}

	// 处理分界时间
	day := now
	if now.Hour() < splitHour {
		day = day.AddDate(0, 0, -1) // 小于分界时间算前一天
	}

	// 基准日（可改成配置的起始日，这里用 1970-01-01）
	base := time.Date(1970, 1, 1, 0, 0, 0, 0, now.Location())

	// 计算天数差
	days := int(day.Sub(base).Hours() / 24)

	// 当前是周期内的第几天
	order := (days % cycle) + startOrder

	// 如果超过周期，则回到 1
	if order > cycle {
		order -= cycle
	}

	return order
}

// IsTodayExecute 判断今天是否执行
func IsTodayExecute(now time.Time, splitHour, cycle, startOrder, executeOrder int) string {
	todayOrder := calculateExecuteOrder(now, splitHour, cycle, startOrder)
	if todayOrder == executeOrder {
		return "✅ 今天要执行"
	}

	return "❌ 今天不执行"
}
