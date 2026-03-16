package OneLongPlan

import (
	"auto-bgi/config"
	"auto-bgi/models"
	"auto-bgi/task"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var weekdayAllowed = map[string]bool{
	"周一": true,
	"周二": true,
	"周三": true,
	"周四": true,
	"周五": true,
	"周六": true,
	"周日": true,
}

var weekdayCN = map[time.Weekday]string{
	time.Monday:    "周一",
	time.Tuesday:   "周二",
	time.Wednesday: "周三",
	time.Thursday:  "周四",
	time.Friday:    "周五",
	time.Saturday:  "周六",
	time.Sunday:    "周日",
}

func parseMaterialValue(value string) (string, int64, error) {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, "-", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("材料值格式错误，应为：材料名称-数量")
	}
	name := strings.TrimSpace(parts[0])
	if name == "" {
		return "", 0, fmt.Errorf("材料名称不能为空")
	}
	numStr := strings.TrimSpace(parts[1])
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || num < 0 {
		return "", 0, fmt.Errorf("材料数量必须为数字")
	}
	return name, num, nil
}

func parseIDQuery(c *gin.Context) (int, bool) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "无效的ID"})
		return 0, false
	}
	return id, true
}

func validatePlanForWrite(c *gin.Context, req *models.OneLongPlan) bool {
	req.Type = strings.TrimSpace(req.Type)
	req.Value = strings.TrimSpace(req.Value)
	req.GroupName = strings.TrimSpace(req.GroupName)

	if req.Type != "日期" && req.Type != "材料" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "类型必须为：日期 或 材料"})
		return false
	}
	if req.GroupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "配置组名称不能为空"})
		return false
	}
	if req.Type == "日期" {
		if !weekdayAllowed[req.Value] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "日期值必须为：周一到周日"})
			return false
		}
		return true
	}
	if _, _, err := parseMaterialValue(req.Value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": err.Error()})
		return false
	}
	return true
}

func List(c *gin.Context) {
	var plans []models.OneLongPlan
	if err := models.DB.Order("id desc").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": plans})
}

func Detail(c *gin.Context) {
	id, ok := parseIDQuery(c)
	if !ok {
		return
	}

	var plan models.OneLongPlan
	if err := models.DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "msg": "计划不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": plan})
}

func Add(c *gin.Context) {
	var req models.OneLongPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "参数格式错误", "error": err.Error()})
		return
	}
	req.ID = 0
	if !validatePlanForWrite(c, &req) {
		return
	}

	if err := models.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": req})
}

func Update(c *gin.Context) {
	var req models.OneLongPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "参数格式错误", "error": err.Error()})
		return
	}
	if req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "缺少ID"})
		return
	}

	var plan models.OneLongPlan
	if err := models.DB.First(&plan, req.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "msg": "计划不存在"})
		return
	}

	if !validatePlanForWrite(c, &req) {
		return
	}

	updates := map[string]interface{}{
		"type":       req.Type,
		"value":      req.Value,
		"group_name": req.GroupName,
	}

	if err := models.DB.Model(&models.OneLongPlan{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
		return
	}

	if err := models.DB.First(&plan, req.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": plan})
}

func Delete(c *gin.Context) {
	id, ok := parseIDQuery(c)
	if !ok {
		return
	}
	if err := models.DB.Delete(&models.OneLongPlan{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "删除成功"})
}

func Run(c *gin.Context) {
	id, ok := parseIDQuery(c)
	if !ok {
		return
	}

	var plan models.OneLongPlan
	if err := models.DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "msg": "计划不存在"})
		return
	}

	now := time.Now()
	if strings.TrimSpace(plan.GroupName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "配置组名称不能为空"})
		return
	}

	if plan.Type == "日期" {
		today := weekdayCN[now.Weekday()]
		if today == plan.Value {
			if err := task.StartGroups([]string{plan.GroupName}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{
				"startedGroups": []string{plan.GroupName},
				"type":          plan.Type,
				"value":         plan.Value,
				"msg":           "已启动配置组",
			}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{
			"startedGroups": []string{},
			"type":          plan.Type,
			"value":         plan.Value,
			"msg":           "未满足启动条件",
		}})
		return
	}

	if plan.Type == "材料" {
		materialName, threshold, err := parseMaterialValue(plan.Value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": err.Error()})
			return
		}
		_, materialMap := config.GetBagInfo()
		if materialMap == nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "获取背包信息失败，请检查米游社Cookie"})
			return
		}
		count := materialMap[materialName]
		if count < threshold {
			if err := task.StartGroups([]string{plan.GroupName}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{
				"startedGroups":  []string{plan.GroupName},
				"type":           plan.Type,
				"value":          plan.Value,
				"materialName":   materialName,
				"materialCount":  count,
				"materialTarget": threshold,
				"msg":            "已启动配置组",
			}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{
			"startedGroups":  []string{},
			"type":           plan.Type,
			"value":          plan.Value,
			"materialName":   materialName,
			"materialCount":  count,
			"materialTarget": threshold,
			"msg":            "未满足启动条件",
		}})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "未知类型"})
}

