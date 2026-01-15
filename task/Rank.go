package task

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/models"
	"auto-bgi/tools"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	downloadLock   sync.Mutex
	lastInvokeTime time.Time
	// 2分钟的时间间隔
	invokeInterval = 5 * time.Minute
)

// 摩拉排行榜
func Rank() string {

	//捕获异常
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("摩拉排行榜异常:%v", r)
			return
		}
	}()

	// 加锁检查调用频率
	downloadLock.Lock()
	// 检查是否在5分钟内重复调用
	if time.Since(lastInvokeTime) < invokeInterval && !lastInvokeTime.IsZero() {
		downloadLock.Unlock()
		return "客官，不可以，您太着急了，请稍后重试哦"

	}
	// 更新最后调用时间
	lastInvokeTime = time.Now()
	downloadLock.Unlock()

	if len(config.GameRoles.Data.List) == 0 {
		return "客官，请先去bgi配置好您的米游社cookie哦"
	}

	over := false

	var records []models.MoraleRecord
	//今天的月
	m := time.Now().Format("01")
	today := models.GetLastMoraleRecord()
	for i := range 50 {
		i++
		fmt.Println("第============", i)
		if over {
			break
		}

		//转成int
		mInt, _ := strconv.Atoi(m)
		async, err := config.GetTravelsDiaryDetailAsync(mInt, 2, i)
		if err != nil {
			autoLog.Sugar.Infof("GetTravelsDiaryDetailAsync%s", err)
			break
		}

		for _, list := range async.List {

			var record models.MoraleRecord
			less, err := tools.IsDateLess(list.Time, today)
			if err != nil {
				autoLog.Sugar.Errorf("IsDateLess:%s", err)
			}
			if less {
				over = true
				break
			}
			fmt.Println(list.ActionID, "======", list.Time, "=======", list.Num, "======", list.Action)
			record.UID = config.GameRoles.Data.List[0].GameId
			record.Name = config.GameRoles.Data.List[0].NicName
			record.Time = list.Time
			record.Num = list.Num
			record.Action = list.Action
			records = append(records, record)
		}
		time.Sleep(time.Second * 3)
	}

	err := models.BatchAddMoraleRecords(records)
	if err != nil {
		autoLog.Sugar.Errorf("BatchAddMoraleRecords==============%s", err)
	}

	SendMoraleRank()

	return fmt.Sprintf("客官，摩拉记录已更新，共%d条新记录", len(records))

}
