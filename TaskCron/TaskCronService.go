package TaskCron

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"strconv"
)

// 列出任务
func List(c *gin.Context) {
	list := Tm.List()
	if list != nil {
		c.JSON(200, list)
		return
	}
	c.JSON(400, []TaskCron{})
}

func CronAdd(c *gin.Context) {
	var taskCron TaskCron
	if err := c.ShouldBindJSON(&taskCron); err != nil {
		c.String(400, "参数错误: %v", err)
		return
	}
	fn, ok := task[taskCron.Name]
	if !ok {
		c.String(400, "任务名称不存在")
		return
	}
	if _, err := Tm.Add(taskCron.Spec, taskCron.Name, taskCron.Data, fn); err != nil {
		c.String(500, "添加失败: %v", err)
		return
	}
	c.String(200, "任务已添加")
}

func CronRemove(c *gin.Context) {
	data := c.Query("id")
	if data == "" {
		c.String(400, "缺少参数 id")
		return
	}
	id, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		c.String(400, "id 解析失败")
		return
	}

	Tm.Remove(cron.EntryID(id))
	c.String(200, "任务已删除")

}

func GetTask(context *gin.Context) {
	var taskList []string
	for k := range task {
		taskList = append(taskList, k)
	}
	context.JSON(200, taskList)
}

func Update(c *gin.Context) {
	var req TaskCron
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, "参数错误: %v", err)
		return
	}

	if req.ID == 0 {
		c.String(400, "缺少任务 ID")
		return
	}
	if req.Spec == "" {
		c.String(400, "缺少 cron 表达式 spec")
		return
	}

	newID, err := Tm.Update(cron.EntryID(req.ID), req.Spec, req.Data)
	if err != nil {
		c.String(500, "更新失败: %v", err)
		return
	}

	// 返回新的 EntryID，前端可以据此刷新列表
	c.JSON(200, gin.H{
		"msg": "任务已更新",
		"id":  newID,
	})

}
