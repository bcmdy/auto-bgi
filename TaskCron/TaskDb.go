package TaskCron

import (
	"auto-bgi/autoLog"
	"auto-bgi/models"
	"fmt"
	"github.com/robfig/cron/v3"
	"strconv"
)

// 插入任务：写入 DB 并回填 EntryID
func (tm *TaskManager) saveToDB(t *models.TaskCron) error {

	//数据库新增定时任务
	create := tm.db.Model(&models.TaskCron{}).Create(t)
	if create.Error != nil {
		autoLog.Sugar.Errorf("创建定时任务失败: %v", create.Error)
		return create.Error
	}
	t.EntryID = t.ID
	t.Status = 1

	return nil

}

// 更新 DB 中的任务
func (tm *TaskManager) updateDB(req models.TaskCron) error {

	//修改定时任务
	updates := tm.db.Model(&models.TaskCron{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"spec":     req.Spec,
		"data":     req.Data,
		"entry_id": req.EntryID,
	})
	if updates.Error != nil {
		autoLog.Sugar.Errorf("更新定时任务失败: %v", updates.Error)
		return updates.Error
	}

	return nil

}

// 删除 DB 中的任务（按主键 id）
func (tm *TaskManager) deleteFromDB(id int) {
	//_, err := tm.db.Exec("DELETE FROM TaskCron WHERE id = ?", id)
	//if err != nil {
	//	autoLog.Sugar.Errorf("删除任务失败（DB）: %v", err)
	//}

	//查询
	var task models.TaskCron
	err := tm.db.Where("id = ?", id).First(&task).Error
	if err != nil {
		autoLog.Sugar.Errorf("任务不存在（DB）: %v", err)
		return
	}

	task, ok := tm.jobs[cron.EntryID(task.EntryID)]
	if !ok {
		autoLog.Sugar.Errorf("删除任务失败: EntryID=%d 不存在\n", id)
		//return
	}
	// 1. 从调度器删除
	tm.c.Remove(cron.EntryID(id))
	// 2. 从内存删除
	delete(tm.jobs, cron.EntryID(id))

	//删除定时任务
	deleteD := tm.db.Model(&models.TaskCron{}).Where("id = ?", id).Delete(&models.TaskCron{})
	if deleteD.Error != nil {
		autoLog.Sugar.Errorf("删除任务失败（DB）: %v", deleteD.Error)
		return
	}

	autoLog.Sugar.Infof("任务[%s] 已删除 (EntryID: %d, DBID: %d)\n", task.Name, task.EntryID, id)

}

// 从数据库加载任务
// 从数据库加载任务，恢复到调度器
func (tm *TaskManager) loadTasksFromDB() {

	//查询所有定时任务
	var tasks []models.TaskCron
	err := tm.db.Find(&tasks).Error
	if err != nil {
		autoLog.Sugar.Errorf("查询定时任务失败: %v", err)
		return
	}

	for i := range tasks {

		// 如果是暂停状态（status=0），只放在内存里/或者干脆忽略，下面给两个方案
		if tasks[i].Status == 0 {
			// 简单版：直接略过，后面 List 从 DB 里查
			//fmt.Printf("任务[%s] 当前为暂停状态(仅入库，不注册到调度器) (EntryID: %d)\n", t.Name, t.EntryID)
			autoLog.Sugar.Infof("任务[%s] 当前为暂停状态(仅入库，不注册到调度器) (EntryID: %d)\n", tasks[i].Name, tasks[i].EntryID)
			continue
		}

		fn, ok := task[tasks[i].Name]
		if !ok {
			//fmt.Printf("跳过未知任务函数[%s] (EntryID: %d)\n", t.Name, t.EntryID)
			autoLog.Sugar.Infof("跳过未知任务函数[%s] (EntryID: %d)\n", tasks[i].Name, tasks[i].EntryID)
			// 3. 从数据库删除对应记录
			tm.deleteFromDB(tasks[i].EntryID)
			continue
		}

		d := tasks[i].Data
		f := fn
		eid, err := tm.c.AddFunc(tasks[i].Spec, func() { f(d) })
		if err != nil {
			//fmt.Printf("恢复任务[%s] 失败 (DBID: %d): %v\n", t.Name, t.DBID, err)
			autoLog.Sugar.Errorf("恢复任务[%s] 失败 (DBID: %d): %v\n", tasks[i].Name, tasks[i].EntryID, err)
			continue
		}

		tasks[i].EntryID = int(eid)
		tm.jobs[eid] = tasks[i]

		autoLog.Sugar.Infof("任务[%s] 已从数据库恢复 (id: %d, EntryID: %d)\n",
			tasks[i].Name, tasks[i].ID, tasks[i].EntryID)
	}

	//批量更新数据库
	err = models.UpdateTaskCron(tasks)
	if err != nil {
		autoLog.Sugar.Errorf("批量更新数据库失败: %v", err)
		return
	}

	autoLog.Sugar.Infof("已从数据库加载任务")
}

// 添加任务：允许同名多条，DB 用自增 id 区分
func (tm *TaskManager) Add(spec, name, data string, fn func(string)) (cron.EntryID, error) {
	// 先准备 TaskCron 对象（不含 ID / DBID）
	t := models.TaskCron{
		Name:   name,
		Spec:   spec,
		Status: 1,
		Data:   data,
	}

	// 2. 再注册到调度器
	d := data
	f := fn
	id, err := tm.c.AddFunc(spec, func() { f(d) })
	if err != nil {
		autoLog.Sugar.Errorf("添加任务到调度器失败: %v", err)
		// 一般最好做一次回滚：把刚插入的 DB 记录删掉
		tm.deleteFromDB(t.EntryID)
		return 0, err
	}
	// 3. 完善任务信息并放入内存映射
	//t.ID = int(id)  // 对外展示使用调度器 EntryID
	t.EntryID = int(id) // 对内使用 DBID
	tm.jobs[id] = t     // key 是 EntryID，value 里包含 DBID

	// 1. 先写入数据库，得到 DBID
	if err := tm.saveToDB(&t); err != nil {
		autoLog.Sugar.Errorf("保存任务到数据库失败: %v", err)
		return 0, err
	}

	autoLog.Sugar.Infof("任务[%s] 已添加 (EntryID: %d, DBID: %d)\n", name, id, t.EntryID)

	return id, nil
}

// 删除任务：根据 EntryID 删除调度器 + 对应 DB 记录
func (tm *TaskManager) Remove(id string) {
	atoi, err := strconv.Atoi(id)
	if err != nil {
		autoLog.Sugar.Errorf("删除任务失败: EntryID=%s 格式错误\n", id)
		return
	}
	//if atoi != 0 {
	//	task, ok := tm.jobs[id]
	//	if !ok {
	//		autoLog.Sugar.Errorf("删除任务失败: EntryID=%d 不存在\n", id)
	//		//return
	//	}
	//	// 1. 从调度器删除
	//	tm.c.Remove(id)
	//	// 2. 从内存删除
	//	delete(tm.jobs, id)
	//
	//	autoLog.Sugar.Infof("任务[%s] 已删除 (EntryID: %d, DBID: %d)\n", task.Name, id, task.EntryID)
	//}
	// 3. 从数据库删除对应记录
	tm.deleteFromDB(atoi)

}

func (tm *TaskManager) List() []models.TaskCron {

	// 查询数据库中的任务列表
	var tasks []models.TaskCron
	err := tm.db.Find(&tasks).Error
	if err != nil {
		autoLog.Sugar.Errorf("查询定时任务失败: %v", err)
		return tasks
	}
	for i := range tasks {
		// 默认没有下一次执行时间
		tasks[i].Next = ""

		if tasks[i].Status == 1 {
			// 运行中的任务，尝试从内存中找到 EntryID 和 Next
			if eid, ok := tm.entryIDByID(tasks[i].EntryID); ok {

				entry := tm.c.Entry(eid)
				if !entry.Next.IsZero() {
					tasks[i].Next = entry.Next.Format("2006-01-02 15:04:05")
				}
			} else {
				fmt.Println(tasks[i])
			}
		}
	}

	return tasks

}

func (tm *TaskManager) entryIDByID(ID int) (cron.EntryID, bool) {

	// 先找到旧任务
	oldTask, ok := tm.jobs[cron.EntryID(ID)]
	if !ok {
		autoLog.Sugar.Errorf("任务不存在: EntryID=%d", ID)
		return 0, false
	}
	return cron.EntryID(oldTask.EntryID), true

	//for _, t := range tm.jobs {
	//	if t.EntryID == ID {
	//		return cron.EntryID(t.EntryID), true
	//	}
	//}
	//return 0, false
}

// 添加任务到调度器 & DB（用于新增）

// 修改任务（根据 EntryID 修改 Spec / Data，不改 Name）
func (tm *TaskManager) Update(req models.TaskCron) (cron.EntryID, error) {

	// 先找到旧任务
	oldTask, ok := tm.jobs[cron.EntryID(req.EntryID)]
	if !ok {
		return 0, fmt.Errorf("任务不存在: EntryID=%d", req.EntryID)
	}

	// 根据旧任务的 Name 找到对应函数
	fn, ok := task[oldTask.Name]
	if !ok {
		return 0, fmt.Errorf("任务函数不存在: name=%s", oldTask.Name)
	}

	// 2. 更新调度器：移除旧任务
	tm.c.Remove(cron.EntryID(req.EntryID))
	delete(tm.jobs, cron.EntryID(req.EntryID))

	// 3. 重新注册新任务
	d := req.Data
	f := fn
	newID, err := tm.c.AddFunc(req.Spec, func() { f(d) })
	if err != nil {
		autoLog.Sugar.Errorf("更新任务失败（重新注册调度器失败）: %v", err)
		return 0, err
	}

	// 4. 更新内存中的任务信息（保留 DBID 和 Name）
	oldTask.Spec = req.Spec
	oldTask.Data = req.Data
	oldTask.EntryID = int(newID)
	tm.jobs[newID] = oldTask
	req.EntryID = int(newID)

	// 1. 更新数据库中该条记录
	if err := tm.updateDB(req); err != nil {

		autoLog.Sugar.Errorf("更新任务数据库失败:%v", err)
		return 0, err
	}

	autoLog.Sugar.Infof("任务[%s] 已更新\n", oldTask.Name)

	return newID, nil
}

// 暂停任务：根据 ID（数据库主键）
func (tm *TaskManager) PauseByID(dbid int) error {
	// 先查当前状态
	var taskCron models.TaskCron
	scan := tm.db.Model(&models.TaskCron{}).Where("id = ?", dbid).Scan(&taskCron)
	if scan.Error != nil {
		return fmt.Errorf("查询任务失败: %v", scan.Error)
	}

	// 如果在调度器中，就移除
	if eid, ok := tm.entryIDByID(taskCron.EntryID); ok {
		tm.c.Remove(eid)
		delete(tm.jobs, eid)
	}

	// 更新 DB 状态
	update := tm.db.Model(&models.TaskCron{}).Where("id = ?", dbid).Update("status", 0)
	if update.Error != nil {
		return fmt.Errorf("更新任务状态失败: %v", update.Error)
	}

	autoLog.Sugar.Infof("任务(DBID=%d) 已暂停\n", dbid)
	return nil
}

// 恢复任务：根据 DBID（数据库主键）
func (tm *TaskManager) ResumeByDBID(dbid int) (cron.EntryID, error) {
	var t models.TaskCron

	first := tm.db.Model(&models.TaskCron{}).Where("id = ?", dbid).First(&t)
	if first.Error != nil {
		return 0, fmt.Errorf("查询任务失败: %v", first.Error)
	}

	//if t.Status == 1 {
	//	return 0, fmt.Errorf("任务当前已是运行状态")
	//}

	fn, ok := task[t.Name]
	if !ok {
		return 0, fmt.Errorf("任务函数不存在: %s", t.Name)
	}

	d := t.Data
	f := fn
	eid, err := tm.c.AddFunc(t.Spec, func() { f(d) })
	if err != nil {
		return 0, fmt.Errorf("恢复任务失败（注册调度器失败）: %v", err)
	}

	// 更新内存
	//t.ID = int(eid)
	t.Status = 1
	t.EntryID = int(eid)
	tm.jobs[eid] = t

	// 更新 DB 状态
	update := tm.db.Model(&models.TaskCron{}).Where("id = ?", dbid).Updates(map[string]interface{}{
		"entry_id": int(eid),
		"status":   1,
	})
	if update.Error != nil {
		return 0, fmt.Errorf("更新任务状态失败: %v", update.Error)
	}

	//fmt.Printf("任务[%s] 已恢复 (EntryID: %d, DBID: %d)\n", t.Name, eid, t.DBID)
	autoLog.Sugar.Infof("任务[%s] 已恢复 (EntryID: %d, DBID: %d)\n", t.Name, t.EntryID, t.ID)
	return eid, nil
}
