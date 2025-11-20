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
	dbid := c.Query("dbid")

	if data == "" {
		c.String(400, "缺少参数 id")
		return
	}
	id, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		c.String(400, "id 解析失败")
		return
	}
	if dbid == "" {
		c.String(400, "dbid 解析失败")
		return
	}
	dbidint, err := strconv.ParseInt(dbid, 10, 64)
	if err != nil {
		c.String(400, "dbid 解析失败")
		return
	}

	Tm.Remove(cron.EntryID(id), int(dbidint))
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
		c.String(400, "只有启动的任务可以修改")
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

func Pause(c *gin.Context) {
	data := c.Query("dbid")
	if data == "" {
		c.String(400, "缺少参数 dbid")
		return
	}
	dbid, err := strconv.Atoi(data)
	if err != nil {
		c.String(400, "dbid 解析失败")
		return
	}

	if err := Tm.PauseByDBID(dbid); err != nil {
		c.String(400, "暂停失败: %v", err)
		return
	}

	c.String(200, "任务已暂停")
}

func Resume(c *gin.Context) {
	data := c.Query("dbid")
	if data == "" {
		c.String(400, "缺少参数 dbid")
		return
	}
	dbid, err := strconv.Atoi(data)
	if err != nil {
		c.String(400, "dbid 解析失败")
		return
	}

	newID, err := Tm.ResumeByDBID(dbid)
	if err != nil {
		c.String(400, "恢复失败: %v", err)
		return
	}

	c.JSON(200, gin.H{
		"msg": "任务已恢复",
		"id":  newID, // 新的 EntryID（如果前端要展示的话）
	})
}
