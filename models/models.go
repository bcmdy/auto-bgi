package models

import "time"

type ArchiveRecords struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `json:"title"`
	ExecuteTime string    `json:"execute_time"`
	CreatedAt   time.Time `json:"created_at"`
}

// 归档记录
func (ArchiveRecords) TableName() string {
	return "archive_records"
}

type AutoBgiConfig struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	AutoBgiKey   string `json:"autobgi_key"`
	AutoBgiValue string `json:"autobgi_value"`
}

// 配置
func (AutoBgiConfig) TableName() string {
	return "autoBgi_config"
}

// 定时任务
type TaskCron struct {
	ID      int    `json:"id"`                       // 对外展示的 ID：cron.EntryID
	EntryID int    `gorm:"entry_id" json:"entry_id"` // 数据库主键 id，只内部使用
	Name    string `json:"name"`
	Spec    string `json:"spec"`
	Next    string `json:"next"`
	Data    string `json:"data"`   // 任务参数
	Status  int    `json:"status"` // 1=运行中, 0=暂停（持久化在 DB）
}

func (TaskCron) TableName() string {
	return "TaskCron"

}
