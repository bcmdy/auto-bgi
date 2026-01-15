package models

// 根据AutoBgiKey更新AutoBgiValue，没有就新增
func UpdateAutoBgiValue(autoBgiKey string, autoBgiValue string) error {
	var autoBgiConfig AutoBgiConfig
	err := DB.Where("auto_bgi_key = ?", autoBgiKey).First(&autoBgiConfig).Error
	if err != nil {
		autoBgiConfig = AutoBgiConfig{
			AutoBgiKey:   autoBgiKey,
			AutoBgiValue: autoBgiValue,
		}
		err = DB.Create(&autoBgiConfig).Error
		if err != nil {
			return err
		}
	} else {
		autoBgiConfig.AutoBgiValue = autoBgiValue
		err = DB.Save(&autoBgiConfig).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// 根据AutoBgiKey获取AutoBgiValue，没有就新增
func GetAutoBgiValue(autoBgiKey string) (string, error) {
	var autoBgiConfig AutoBgiConfig
	err := DB.Where("auto_bgi_key = ?", autoBgiKey).First(&autoBgiConfig).Error
	if err != nil {
		return "", err
	}
	return autoBgiConfig.AutoBgiValue, nil
}
