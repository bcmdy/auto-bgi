package task

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/models"
	"github.com/gin-gonic/gin"
)

func QueryMoraleRecord(c *gin.Context) {

	if len(config.GameRoles.Data.List) == 0 {
		autoLog.Sugar.Errorf("客官，请先去bgi配置好您的米游社cookie哦")
		c.JSON(400, gin.H{"error": "客官，请先去bgi配置好您的米游社cookie哦"})
		return
	}

	dateQuery := c.Query("date")

	if dateQuery == "" {
		c.JSON(400, gin.H{"error": "必须传递 date 参数，格式如: 2026-01-03, 2026-01, 或 2026"})
		return
	}

	// 2. 定义返回结构
	type ResponseData struct {
		TargetDate  string                `json:"target_date"`
		TotalMorale int                   `json:"total_morale"`
		Items       []models.MoraleRecord `json:"items"`
	}
	var res ResponseData
	res.TargetDate = dateQuery

	// 3. 构建模糊查询条件
	// 如果是 2026-01-03，LIKE '2026-01-03%' 会匹配当天所有秒数
	// 如果是 2026-01，LIKE '2026-01%' 会匹配当月所有天
	likeCondition := dateQuery + "%"

	// 4. 查询该时间段的总收益
	// 使用 COALESCE 处理没有数据的情况，防止 Scan 报错
	err := models.DB.Model(&models.MoraleRecord{}).
		Select("COALESCE(SUM(num), 0)").
		Where("time LIKE ? and uid = ?", likeCondition, config.GameRoles.Data.List[0].GameId).
		Row().Scan(&res.TotalMorale)

	if err != nil {
		c.JSON(500, gin.H{"error": "统计收益失败", "details": err.Error()})
		return
	}

	// 5. 查询该时间段的明细
	err = models.DB.Model(&models.MoraleRecord{}).
		Where("time LIKE ?  and uid = ?", likeCondition, config.GameRoles.Data.List[0].GameId).
		Order("time DESC").
		Find(&res.Items).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "查询明细失败", "details": err.Error()})
		return
	}

	// 6. 返回结果
	c.JSON(200, gin.H{
		"data": res,
	})
}

// 更新摩拉记录
func UpdateMoraleRecord(c *gin.Context) {
	rank := Rank()

	c.JSON(200, gin.H{"message": rank})

}
