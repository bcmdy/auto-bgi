package Starward

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var StarwardDB *gorm.DB

func StarwardDBInit() error {

	var err error
	StarwardDB, err = gorm.Open(sqlite.Open("Z:\\res\\ruienni\\Starward\\StarwardDatabase.db"), &gorm.Config{
		//Logger: newLogger, // !!! 把 logger 加进去
		//Logger: logger.Default.LogMode(logger.Error),
		Logger: logger.Default.LogMode(logger.Silent), // 不打印任何日志
	})
	////逻辑
	//DB.Migrator().DropIndex(&MoraleRecord{}, "idx_MoraleRecord_time")

	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	return nil
}
