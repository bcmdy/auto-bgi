package BackpackStatistics

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/models"
	"strings"
)

// 读取关注的背包统计材料
func List() (backpackStatistics []models.BackpackStatistics) {

	tx := models.DB.Model(&models.BackpackStatistics{}).Find(&backpackStatistics)
	if tx.Error != nil {
		autoLog.Sugar.Errorf("读取背包统计材料失败: %v", tx.Error)
		return nil
	}
	if len(backpackStatistics) == 0 {
		data := config.Cfg.BagStatistics
		split := strings.Split(data, ",")
		for _, v := range split {
			backpackStatistics = append(backpackStatistics, models.BackpackStatistics{
				Material: v,
			})
		}
		//批量新增
		tx := models.DB.Create(&backpackStatistics)
		if tx.Error != nil {
			autoLog.Sugar.Errorf("批量新增背包统计材料失败: %v", tx.Error)
			return nil
		}
	}
	return backpackStatistics
}

// 新增背包统计材料
func Add(material string) error {
	//判断是否已存在,存在则不新增
	backpackStatistics := models.BackpackStatistics{}
	tx := models.DB.Model(&models.BackpackStatistics{}).Where("material = ?", material).First(&backpackStatistics)
	if tx.Error == nil {
		return nil
	}

	tx = models.DB.Create(&models.BackpackStatistics{
		Material: material,
	})
	if tx.Error != nil {
		autoLog.Sugar.Errorf("新增背包统计材料失败: %v", tx.Error)
		return tx.Error
	}
	autoLog.Sugar.Infof("材料关注成功")
	return nil
}

// 删除背包统计材料
func Delete(material string) error {
	tx := models.DB.Model(&models.BackpackStatistics{}).Where("material = ?", material).Delete(&models.BackpackStatistics{})
	if tx.Error != nil {
		autoLog.Sugar.Errorf("取消材料关注失败: %v", tx.Error)
		return tx.Error
	}
	autoLog.Sugar.Infof("材料取消关注成功")
	return nil
}
