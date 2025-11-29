package models

import "strings"

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
// 返回值: interface{} - 成功时返回nil，失败时返回错误信息
func DeleteAllArchiveRecords() interface{} {
	// 使用Unscoped()进行软删除，彻底删除所有记录
	// 删除ArchiveRecords表中的所有记录
	db := DB.Unscoped().Delete(&ArchiveRecords{})
	// 检查操作是否出错
	if db.Error != nil {
		// 如果出错，返回错误信息
		return db.Error
	}
	// 操作成功，返回nil
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
