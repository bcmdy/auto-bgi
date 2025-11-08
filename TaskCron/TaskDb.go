package TaskCron

import (
	"auto-bgi/autoLog"
	"fmt"
	"github.com/robfig/cron/v3"
)

func (tm *TaskManager) saveToDB(name, spec, data string) {
	_, err := tm.db.Exec("INSERT OR REPLACE INTO tasks(name, spec, data) VALUES (?, ?, ?)", name, spec, data)
	if err != nil {
		autoLog.Sugar.Errorf("保存任务到数据库失败: %v", err)
	}
}

func (tm *TaskManager) deleteFromDB(name string) {
	_, err := tm.db.Exec("DELETE FROM tasks WHERE name = ?", name)
	if err != nil {
		autoLog.Sugar.Errorf("删除任务失败: %v", err)
	}
}

// 只注册调度器（用于启动恢复，不写库）
func (tm *TaskManager) addScheduleOnly(spec, name, data string, fn func(string)) (cron.EntryID, error) {
	d := data
	f := fn
	id, err := tm.c.AddFunc(spec, func() { f(d) })
	if err != nil {
		return 0, err
	}
	tm.jobs[id] = TaskCron{Name: name, Spec: spec, Data: data}
	autoLog.Sugar.Infof("任务[%s] 已恢复到调度器 (ID: %d)\n", name, id)
	return id, nil
}

// 从数据库加载任务
func (tm *TaskManager) loadTasksFromDB() {
	rows, err := tm.db.Query("SELECT name, spec, data FROM tasks")
	if err != nil {
		fmt.Println("加载任务失败:", err)
		return
	}
	defer rows.Close()

	var todo []TaskCron
	for rows.Next() {
		var t TaskCron
		if err := rows.Scan(&t.Name, &t.Spec, &t.Data); err != nil {

			autoLog.Sugar.Errorf("loadTasksFromDB读取行失败:", err)
			continue
		}
		todo = append(todo, t)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("扫描行失败:", err)
		autoLog.Sugar.Errorf("扫描行失败:", err)
		return
	}

	for _, t := range todo {
		if fn, ok := task[t.Name]; ok {
			if _, err := tm.addScheduleOnly(t.Spec, t.Name, t.Data, fn); err != nil {

				autoLog.Sugar.Infof("恢复任务[%s] 失败: %v\n", t.Name, err)
			}
		} else {
			autoLog.Sugar.Infof("跳过未知任务函数[%s]\n", t.Name)
		}
	}

	autoLog.Sugar.Infof("已从数据库加载任务")
}

// 删除任务
func (tm *TaskManager) Remove(id cron.EntryID) {
	task := tm.jobs[id]
	tm.c.Remove(id)
	delete(tm.jobs, id)
	tm.deleteFromDB(task.Name)
	autoLog.Sugar.Infof("任务 ID=%d 已删除\n", id)
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
func (tm *TaskManager) Add(spec, name, data string, fn func(string)) cron.EntryID {
	d := data
	f := fn
	id, err := tm.c.AddFunc(spec, func() { f(d) })
	if err != nil {

		autoLog.Sugar.Errorf("添加任务失败:%v", err)
		return 0
	}
	tm.jobs[id] = TaskCron{Name: name, Spec: spec, Data: data}
	tm.saveToDB(name, spec, data)

	autoLog.Sugar.Infof("任务[%s] 已添加 (ID: %d)\n", name, id)
	return id
}
