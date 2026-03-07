package bgiStatus

import (
	"auto-bgi/models"
)

// 查询所有规定记录
func ArchiveRecordList() (map[string]string, error) {
	var archiveRecords []models.ArchiveRecords

	tx := models.DB.Model(&models.ArchiveRecords{}).Find(&archiveRecords)
	if tx.Error != nil {
		return nil, tx.Error
	}
	dataMap := make(map[string]string)
	for _, archiveRecord := range archiveRecords {
		dataMap[archiveRecord.Title] = archiveRecord.ExecuteTime
	}

	return dataMap, nil

}

// 根据名字查询单条记录
func ArchiveRecordDetail(title string) (models.ArchiveRecords, error) {

	var archiveRecords models.ArchiveRecords
	query := models.DB.Model(&models.ArchiveRecords{}).Where("title = ?", title).First(&archiveRecords)
	if query.Error != nil {
		return archiveRecords, query.Error
	}
	return archiveRecords, nil
}
