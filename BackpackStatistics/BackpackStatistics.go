package BackpackStatistics

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/models"
	"encoding/json"
	"os"
	"path/filepath"
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
		if len(split) == 0 {
			return nil
		}
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
		config.Cfg.BagStatistics = ""
		// 序列化为JSON字符串，格式化输出
		aa, err := json.MarshalIndent(config.Cfg, "", "  ")
		if err != nil {
			return
		}

		// 写入 main.json，路径可以自定义，这里示例写当前运行目录
		filePath := filepath.Join(".", "main.json")
		err = os.WriteFile(filePath, aa, 0644)
		if err != nil {

			return
		}
		//重新加载配置文件
		_ = config.ReloadConfig()
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

// 清空所有背包统计
func ClearAll() error {
	// 使用 Where("1 = 1") 来绕过全局删除保护
	tx := models.DB.Where("1 = 1").Delete(&models.BackpackStatistics{})
	if tx.Error != nil {
		autoLog.Sugar.Errorf("清空背包统计失败: %v", tx.Error)
		return tx.Error
	}
	return nil
}
