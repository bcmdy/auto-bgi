package TaskCron

import (
	"auto-bgi/autoLog"
	"fmt"
	"github.com/robfig/cron/v3"
)

// 插入任务：写入 DB 并回填 DBID
func (tm *TaskManager) saveToDB(t *TaskCron) error {
	res, err := tm.db.Exec(
		"INSERT INTO TaskCron(name, spec, data, status) VALUES (?, ?, ?, 1)",
		t.Name, t.Spec, t.Data,
	)
	if err != nil {
		return err
	}
	lastID, _ := res.LastInsertId()
	t.DBID = int(lastID)
	t.Status = 1
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
	rows, err := tm.db.Query("SELECT id, name, spec, data, status FROM TaskCron")
	if err != nil {
		//fmt.Println("加载任务失败:", err)
		autoLog.Sugar.Errorf("加载任务失败: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t TaskCron
		if err := rows.Scan(&t.DBID, &t.Name, &t.Spec, &t.Data, &t.Status); err != nil {
			//fmt.Println("读取行失败:", err)
			autoLog.Sugar.Errorf("读取行失败:%v", err)
			continue
		}

		// 如果是暂停状态（status=0），只放在内存里/或者干脆忽略，下面给两个方案
		if t.Status == 0 {
			// 简单版：直接略过，后面 List 从 DB 里查
			//fmt.Printf("任务[%s] 当前为暂停状态(仅入库，不注册到调度器) (DBID: %d)\n", t.Name, t.DBID)
			autoLog.Sugar.Infof("任务[%s] 当前为暂停状态(仅入库，不注册到调度器) (DBID: %d)\n", t.Name, t.DBID)
			continue
		}

		fn, ok := task[t.Name]
		if !ok {
			//fmt.Printf("跳过未知任务函数[%s] (DBID: %d)\n", t.Name, t.DBID)
			autoLog.Sugar.Infof("跳过未知任务函数[%s] (DBID: %d)\n", t.Name, t.DBID)
			// 3. 从数据库删除对应记录
			tm.deleteFromDB(t.DBID)
			continue
		}

		d := t.Data
		f := fn
		eid, err := tm.c.AddFunc(t.Spec, func() { f(d) })
		if err != nil {
			//fmt.Printf("恢复任务[%s] 失败 (DBID: %d): %v\n", t.Name, t.DBID, err)
			autoLog.Sugar.Errorf("恢复任务[%s] 失败 (DBID: %d): %v\n", t.Name, t.DBID, err)
			continue
		}

		t.ID = int(eid) // 对外展示的 EntryID
		tm.jobs[eid] = t

		//fmt.Printf("任务[%s] 已从数据库恢复 (EntryID: %d, DBID: %d)\n",
		//	t.Name, eid, t.DBID)
		autoLog.Sugar.Infof("任务[%s] 已从数据库恢复 (EntryID: %d, DBID: %d)\n",
			t.Name, eid, t.DBID)
	}

	if err := rows.Err(); err != nil {
		//fmt.Println("扫描行失败:", err)
		autoLog.Sugar.Errorf("扫描行失败: %v", err)
		return
	}

	//fmt.Println("已从数据库加载任务")
	autoLog.Sugar.Infof("已从数据库加载任务")
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
func (tm *TaskManager) Remove(id cron.EntryID, dbid int) {
	if id != 0 {
		task, ok := tm.jobs[id]
		if !ok {
			autoLog.Sugar.Errorf("删除任务失败: EntryID=%d 不存在\n", id)
			//return
		}
		// 1. 从调度器删除
		tm.c.Remove(id)
		// 2. 从内存删除
		delete(tm.jobs, id)

		autoLog.Sugar.Infof("任务[%s] 已删除 (EntryID: %d, DBID: %d)\n", task.Name, id, task.DBID)
	}
	// 3. 从数据库删除对应记录
	tm.deleteFromDB(dbid)
	autoLog.Sugar.Infof("任务已从数据库删除 (DBID: %d)\n", dbid)
}

func (tm *TaskManager) List() []TaskCron {
	rows, err := tm.db.Query("SELECT id, name, spec, data, status FROM TaskCron")
	if err != nil {
		fmt.Println("查询任务列表失败:", err)
		return nil
	}
	defer rows.Close()

	var list []TaskCron

	for rows.Next() {
		var t TaskCron
		if err := rows.Scan(&t.DBID, &t.Name, &t.Spec, &t.Data, &t.Status); err != nil {
			fmt.Println("读取行失败:", err)
			continue
		}

		// 默认没有下一次执行时间
		t.Next = ""

		if t.Status == 1 {
			// 运行中的任务，尝试从内存中找到 EntryID 和 Next
			if eid, ok := tm.entryIDByDBID(t.DBID); ok {
				t.ID = int(eid)
				entry := tm.c.Entry(eid)
				if !entry.Next.IsZero() {
					t.Next = entry.Next.Format("2006-01-02 15:04:05")
				}
			} else {
				// 理论上不应该出现，可能是异常情况
				t.ID = 0
			}
		} else {
			// 暂停任务：不在调度器中，ID=0，Next 为空
			t.ID = 0
		}

		list = append(list, t)
	}

	return list
}

func (tm *TaskManager) entryIDByDBID(dbid int) (cron.EntryID, bool) {
	for eid, t := range tm.jobs {
		if t.DBID == dbid {
			return eid, true
		}
	}
	return 0, false
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

// 暂停任务：根据 DBID（数据库主键）
func (tm *TaskManager) PauseByDBID(dbid int) error {
	// 先查当前状态
	var status int
	err := tm.db.QueryRow("SELECT status FROM TaskCron WHERE id = ?", dbid).Scan(&status)
	if err != nil {
		return fmt.Errorf("查询任务失败: %v", err)
	}
	if status == 0 {
		return fmt.Errorf("任务已经是暂停状态")
	}

	// 如果在调度器中，就移除
	if eid, ok := tm.entryIDByDBID(dbid); ok {
		tm.c.Remove(eid)
		delete(tm.jobs, eid)
	}

	// 更新 DB 状态
	if _, err := tm.db.Exec("UPDATE TaskCron SET status = 0 WHERE id = ?", dbid); err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	//fmt.Printf("任务(DBID=%d) 已暂停\n", dbid)
	autoLog.Sugar.Infof("任务(DBID=%d) 已暂停\n", dbid)
	return nil
}

// 恢复任务：根据 DBID（数据库主键）
func (tm *TaskManager) ResumeByDBID(dbid int) (cron.EntryID, error) {
	var t TaskCron
	err := tm.db.QueryRow(
		"SELECT id, name, spec, data, status FROM TaskCron WHERE id = ?",
		dbid,
	).Scan(&t.DBID, &t.Name, &t.Spec, &t.Data, &t.Status)
	if err != nil {
		return 0, fmt.Errorf("查询任务失败: %v", err)
	}
	if t.Status == 1 {
		return 0, fmt.Errorf("任务当前已是运行状态")
	}

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
	t.ID = int(eid)
	t.Status = 1
	tm.jobs[eid] = t

	// 更新 DB 状态
	if _, err := tm.db.Exec("UPDATE TaskCron SET status = 1 WHERE id = ?", dbid); err != nil {
		return 0, fmt.Errorf("更新任务状态失败: %v", err)
	}

	//fmt.Printf("任务[%s] 已恢复 (EntryID: %d, DBID: %d)\n", t.Name, eid, t.DBID)
	autoLog.Sugar.Infof("任务[%s] 已恢复 (EntryID: %d, DBID: %d)\n", t.Name, eid, t.DBID)
	return eid, nil
}
