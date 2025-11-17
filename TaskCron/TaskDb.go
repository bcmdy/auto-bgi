package TaskCron

import (
	"auto-bgi/autoLog"
	"fmt"
	"github.com/robfig/cron/v3"
)

// 插入任务：写入 DB 并回填 DBID
func (tm *TaskManager) saveToDB(t *TaskCron) error {
	res, err := tm.db.Exec(
		"INSERT INTO TaskCron(name, spec, data) VALUES (?, ?, ?)",
		t.Name, t.Spec, t.Data,
	)
	if err != nil {
		return err
	}
	lastID, _ := res.LastInsertId()
	t.DBID = int(lastID)
	return nil
}

// 更新 DB 中的任务
func (tm *TaskManager) updateDB(id int, spec, data string) error {
	_, err := tm.db.Exec(
		"UPDATE TaskCron SET spec = ?, data = ? WHERE id = ?",
		spec, data, id,
	)
	return err
}

// 删除 DB 中的任务（按主键 id）
func (tm *TaskManager) deleteFromDB(id int) {
	_, err := tm.db.Exec("DELETE FROM TaskCron WHERE id = ?", id)
	if err != nil {
		autoLog.Sugar.Errorf("删除任务失败（DB）: %v", err)
	}
}

// 从数据库加载任务
// 从数据库加载任务，恢复到调度器
func (tm *TaskManager) loadTasksFromDB() {
	rows, err := tm.db.Query("SELECT id, name, spec, data FROM TaskCron")
	if err != nil {
		autoLog.Sugar.Errorf("加载任务失败: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t TaskCron
		if err := rows.Scan(&t.DBID, &t.Name, &t.Spec, &t.Data); err != nil {
			autoLog.Sugar.Errorf("读取行失败: %v", err)
			continue
		}

		fn, ok := task[t.Name]
		if !ok {
			autoLog.Sugar.Errorf("跳过未知任务函数[%s] (DBID: %d)\n", t.Name, t.DBID)
			continue
		}

		d := t.Data
		f := fn
		eid, err := tm.c.AddFunc(t.Spec, func() { f(d) })
		if err != nil {
			autoLog.Sugar.Errorf("恢复任务[%s] 失败 (DBID: %d): %v\n", t.Name, t.DBID, err)
			continue
		}

		t.ID = int(eid) // 对外展示的 EntryID
		tm.jobs[eid] = t

		autoLog.Sugar.Debugf("任务[%s] 已从数据库恢复 (EntryID: %d, DBID: %d)\n", t.Name, eid, t.DBID)
	}

	if err := rows.Err(); err != nil {
		autoLog.Sugar.Errorf("扫描行失败: %v", err)
		return
	}

	fmt.Println("已从数据库加载任务")
}

// 添加任务：允许同名多条，DB 用自增 id 区分
func (tm *TaskManager) Add(spec, name, data string, fn func(string)) (cron.EntryID, error) {
	// 先准备 TaskCron 对象（不含 ID / DBID）
	t := TaskCron{
		Name: name,
		Spec: spec,
		Data: data,
	}

	// 1. 先写入数据库，得到 DBID
	if err := tm.saveToDB(&t); err != nil {
		autoLog.Sugar.Errorf("保存任务到数据库失败: %v", err)
		return 0, err
	}

	// 2. 再注册到调度器
	d := data
	f := fn
	id, err := tm.c.AddFunc(spec, func() { f(d) })
	if err != nil {
		autoLog.Sugar.Errorf("添加任务到调度器失败: %v", err)
		// 一般最好做一次回滚：把刚插入的 DB 记录删掉
		tm.deleteFromDB(t.DBID)
		return 0, err
	}

	// 3. 完善任务信息并放入内存映射
	t.ID = int(id)  // 对外展示使用调度器 EntryID
	tm.jobs[id] = t // key 是 EntryID，value 里包含 DBID

	autoLog.Sugar.Infof("任务[%s] 已添加 (EntryID: %d, DBID: %d)\n", name, id, t.DBID)

	return id, nil
}

// 删除任务：根据 EntryID 删除调度器 + 对应 DB 记录
func (tm *TaskManager) Remove(id cron.EntryID) {
	task, ok := tm.jobs[id]
	if !ok {
		autoLog.Sugar.Errorf("删除任务失败: EntryID=%d 不存在\n", id)
		return
	}

	// 1. 从调度器删除
	tm.c.Remove(id)
	// 2. 从内存删除
	delete(tm.jobs, id)
	// 3. 从数据库删除对应记录
	tm.deleteFromDB(task.DBID)

	autoLog.Sugar.Infof("任务[%s] 已删除 (EntryID: %d, DBID: %d)\n", task.Name, id, task.DBID)
}

func (tm *TaskManager) List() []TaskCron {
	var taskCrons []TaskCron
	for id, task := range tm.jobs {
		entry := tm.c.Entry(id)
		taskCrons = append(taskCrons, TaskCron{
			ID:   int(id),
			Name: task.Name,
			Spec: task.Spec,
			Next: entry.Next.Format("2006-01-02 15:04:05"),
			Data: task.Data, // NEW
		})
	}
	return taskCrons
}

// 添加任务到调度器 & DB（用于新增）

// 修改任务（根据 EntryID 修改 Spec / Data，不改 Name）
func (tm *TaskManager) Update(id cron.EntryID, spec, data string) (cron.EntryID, error) {
	// 先找到旧任务
	oldTask, ok := tm.jobs[id]
	if !ok {
		return 0, fmt.Errorf("任务不存在: EntryID=%d", id)
	}

	// 根据旧任务的 Name 找到对应函数
	fn, ok := task[oldTask.Name]
	if !ok {
		return 0, fmt.Errorf("任务函数不存在: name=%s", oldTask.Name)
	}

	// 1. 更新数据库中该条记录
	if err := tm.updateDB(oldTask.DBID, spec, data); err != nil {

		autoLog.Sugar.Errorf("更新任务数据库失败:%v", err)
		return 0, err
	}

	// 2. 更新调度器：移除旧任务
	tm.c.Remove(id)
	delete(tm.jobs, id)

	// 3. 重新注册新任务
	d := data
	f := fn
	newID, err := tm.c.AddFunc(spec, func() { f(d) })
	if err != nil {
		autoLog.Sugar.Errorf("更新任务失败（重新注册调度器失败）: %v", err)
		return 0, err
	}

	// 4. 更新内存中的任务信息（保留 DBID 和 Name）
	oldTask.ID = int(newID)
	oldTask.Spec = spec
	oldTask.Data = data
	tm.jobs[newID] = oldTask

	autoLog.Sugar.Infof("任务[%s] 已更新 (旧EntryID: %d, 新EntryID: %d, DBID: %d)\n", oldTask.Name, id, newID, oldTask.DBID)

	return newID, nil
}
