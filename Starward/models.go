package Starward

// GenshinGachaItem 抽卡物品
type GenshinGachaItem struct {
	Uid       int    `json:"uid" gorm:"primaryKey;column:uid;not null"`
	Id        int    `json:"id" gorm:"primaryKey;column:id;not null"`
	Name      string `json:"name" gorm:"column:name;not null"`
	Time      string `json:"time" gorm:"column:time;not null"`
	ItemId    int    `json:"item_id" gorm:"column:item_id;not null"`
	ItemType  string `json:"item_type" gorm:"column:item_type;not null"`
	RankType  int    `json:"rank_type" gorm:"column:rank_type;not null"`
	GachaType int    `json:"gacha_type" gorm:"column:gacha_type;not null"`
	Count     int    `json:"count" gorm:"column:count;not null"`
	Lang      string `json:"lang" gorm:"column:lang"`
}

// TableName 指定表名
func (GenshinGachaItem) TableName() string {
	return "GenshinGachaItem"
}

type GenshinGachaInfo struct {
	Id          int    `json:"id" gorm:"primaryKey;column:id;not null" db:"id"`
	Name        string `json:"name" gorm:"column:name" db:"name"`
	Icon        string `json:"icon" gorm:"column:icon" db:"icon"`
	Element     int    `json:"element" gorm:"column:element;not null" db:"element"`
	Level       int    `json:"level" gorm:"column:level;not null" db:"level"`
	CatId       int    `json:"cat_id" gorm:"column:cat_id;not null" db:"cat_id"`
	WeaponCatId int    `json:"weapon_cat_id" gorm:"column:weapon_cat_id;not null" db:"weapon_cat_id"`
}

// TableName 指定表名
func (GenshinGachaInfo) TableName() string {
	return "GenshinGachaInfo"
}
