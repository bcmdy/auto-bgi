package models

import "fmt"

func UpdateTaskCron(tasks []TaskCron) error {

	//批量更新任务
	for _, task := range tasks {

		err := DB.Model(&TaskCron{}).Where("id = ?", task.ID).Update("entry_id", task.EntryID).Error
		if err != nil {
			return fmt.Errorf("updateTaskCron err:%v", err)
		}
	}

	return nil

}
