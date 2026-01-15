package models

import (
	"auto-bgi/autoLog"
	"strings"
)

// DeleteArchiveRecord 用于删除档案记录
// 参数:
//
//	id - 要删除的档案记录的ID
//
// 返回值:
//
//	error - 操作过程中遇到的错误，如果删除成功则返回nil
func DeleteArchiveRecord(id int) error {

	// 使用config.DB执行删除操作，删除条件为传入的id
	db := DB.Delete(&ArchiveRecords{}, id)
	// 检查删除操作是否出错
	if db.Error != nil {
		return db.Error
	}
	// 如果删除成功，返回nil
	return nil
}

// DeleteAllArchiveRecords 删除所有归档记录的函数
func DeleteAllArchiveRecords() error {
	db := DB.Unscoped().
		Where("1 = 1").
		Delete(&ArchiveRecords{})

	if db.Error != nil {
		autoLog.Sugar.Error(db.Error)
		return db.Error
	}
	return nil
}

/**
 * 根据标题删除档案记录
 * @param title 要删除的档案记录的标题
 * @return interface{} 操作结果，成功返回nil，失败返回错误信息
 */
func DeleteArchiveRecordByTitle(title string) interface{} {
	// 使用Unscoped()进行软删除，彻底删除所有记录
	// 删除ArchiveRecords表中的所有记录
	db := DB.Unscoped().Where("title = ?", title).Delete(&ArchiveRecords{})
	// 检查操作是否出错
	if db.Error != nil {
		// 如果出错，返回错误信息
		return db.Error
	}
	// 操作成功，返回nil
	return nil

}

func InsertArchiveRecord(title string, s string) interface{} {

	// 创建一个新的ArchiveRecords对象，并设置其title和content字段
	archiveRecord := ArchiveRecords{Title: title, ExecuteTime: s}
	// 使用config.DB执行插入操作
	db := DB.Create(&archiveRecord)
	// 检查插入操作是否出错
	if db.Error != nil {
		return db.Error
	}
	// 插入成功，返回nil
	return nil
}

func GetArchiveRecordByTitle(name string) ArchiveRecords {
	if strings.Contains(name, "==(已经结束)") {
		name = strings.Replace(name, "==(已经结束)", "", 1)
	}

	var archiveRecords ArchiveRecords
	db := DB.Where("title = ?", name).First(&archiveRecords)
	if db.Error != nil {
		return archiveRecords
	}
	return archiveRecords

}

func ListArchiveRecords() []ArchiveRecords {
	var archiveRecords []ArchiveRecords
	db := DB.Find(&archiveRecords)
	if db.Error != nil {
		return archiveRecords
	}
	return archiveRecords
}
