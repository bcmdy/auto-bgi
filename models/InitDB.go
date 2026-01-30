package models

import (
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite" // 显式别名，避免冲突
	"gorm.io/gorm"          // 显式别名，避免本地包名冲突
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {

	//newLogger := logger.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags), // 输出到控制台
	//	logger.Config{
	//		SlowThreshold: time.Second, // 慢 SQL 阈值
	//		LogLevel:      logger.Info, // 输出 Info 级别（会打印所有 SQL）
	//		Colorful:      true,        // 彩色输出
	//	},
	//)

	var err error
	DB, err = gorm.Open(sqlite.Open("./archive.db"), &gorm.Config{
		//Logger: newLogger, // !!! 把 logger 加进去
		//Logger: logger.Default.LogMode(logger.Error),
		Logger: logger.Default.LogMode(logger.Silent), // 不打印任何日志
	})
	////逻辑
	//DB.Migrator().DropIndex(&MoraleRecord{}, "idx_MoraleRecord_time")

	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}
	// 自动迁移模式
	err2 := DB.AutoMigrate(&ArchiveRecords{}, &AutoBgiConfig{}, &TaskCron{}, &BackpackStatistics{}, &MoraleRecord{})
	if err2 != nil {
		return fmt.Errorf("failed to migrate database: %v", err2)
	}
	return nil
}
