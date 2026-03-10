package bgiStatus

import "auto-bgi/autoLog"

type OneLongLogProgress struct {
	OneLongName string
	// 存储具体的状态值
	Details map[string]bool
	// 存储 Key 的顺序，以便按顺序遍历
	Order []string
}

var OneLongProgress = OneLongLogProgress{
	OneLongName: "",
	Details:     make(map[string]bool),
	Order:       []string{},
}

func InitialOneLongProgress(OneLongName string) {
	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			autoLog.Sugar.Errorf("初始化一条龙进度失败: %v", err)
		}
	}()

	if OneLongProgress.OneLongName == OneLongName {
		return
	}
	OneLongProgress = OneLongLogProgress{
		OneLongName: "",
		Details:     make(map[string]bool),
		Order:       []string{},
	}

	dragonConfig := readOneDragonConfig(OneLongName + ".json")

	OneLongProgress.OneLongName = OneLongName

	for _, enabled := range dragonConfig.TaskEnabledList {
		if enabled.Enabled {
			// 1. 记录顺序
			OneLongProgress.Order = append(OneLongProgress.Order, enabled.Name)
			// 2. 初始化值
			OneLongProgress.Details[enabled.Name] = false
		}
	}

	autoLog.Sugar.Infof("初始化一条龙进度：%v", OneLongProgress)
}

// 获取没有跑完的配置组
func NoRunningOneLongProgress() []string {
	var unfinishedOneLong []string
	// 按照插入时的顺序进行输出
	for _, name := range OneLongProgress.Order {
		status := OneLongProgress.Details[name]
		if !status {
			unfinishedOneLong = append(unfinishedOneLong, name)
		}
	}
	return unfinishedOneLong
}
