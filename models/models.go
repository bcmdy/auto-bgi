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
	AutoBgiKey   string `gorm:"auto_bgi_key"`
	AutoBgiValue string `gorm:"auto_bgi_value"`
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
	Mark    string `json:"mark"`   //备注
}

func (TaskCron) TableName() string {
	return "TaskCron"
}

// 背包统计
type BackpackStatistics struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Material string `json:"material"`
}

func (BackpackStatistics) TableName() string {
	return "BackpackStatistics"
}

// 摩拉记录
type MoraleRecord struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	UID    string `json:"uid"`
	Name   string `json:"name"`
	Time   string `json:"Time"`
	Num    int    `json:"morale"` // 摩拉
	Action string `json:"action"` // 收入/支出
}

func (MoraleRecord) TableName() string {
	return "MoraleRecord_two"
}
