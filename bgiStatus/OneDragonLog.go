package bgiStatus

type OneLongLogProgress struct {
	OneLongName string
	// 存储具体的状态值
	Details map[string]bool
	// 存储 Key 的顺序，以便按顺序遍历
	Order []string
}

var OneLongProgress OneLongLogProgress

func InitialOneLongProgress(OneLongName string) {
	dragonConfig := readOneDragonConfig(OneLongName + ".json")

	OneLongProgress.OneLongName = dragonConfig.Name
	OneLongProgress.Details = make(map[string]bool)
	OneLongProgress.Order = []string{} // 初始化顺序切片

	for _, enabled := range dragonConfig.TaskEnabledList {
		if enabled.Enabled {
			// 1. 记录顺序
			OneLongProgress.Order = append(OneLongProgress.Order, enabled.Name)
			// 2. 初始化值
			OneLongProgress.Details[enabled.Name] = false
		}
	}
}
