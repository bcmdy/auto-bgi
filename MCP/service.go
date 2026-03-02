package MCP

import (
	"auto-bgi/OneLong"
	"auto-bgi/bgiStatus"
	"github.com/iancoleman/orderedmap"
	"time"
)

type Res struct {
	Time    string
	Msg     string
	MapData []*orderedmap.OrderedMap
}

var OneLongService OneLong.OneLong

func GetBgiIndex() (Res, error) {

	info := bgiStatus.BgiLogStatusInfo
	data := make([]*orderedmap.OrderedMap, 8)
	data[0] = orderedmap.New()
	data[0].Set("当前配置组：", info.Group+" ["+info.GroupProgress+"]")
	data[0].Set("预计：", info.Timestamp)
	data[0].Set("当前路线：", info.MapTrackingLine)
	data[0].Set("进度：", info.ConfigurationGroupExecutionProgress)
	data[0].Set("bgi运行状态：", info.Running)
	data[0].Set("js运行进度：", info.JSProgress)
	var res Res
	res.Time = time.Now().Format("2006-01-02")
	res.Msg = "bgi运行情况"
	res.MapData = data
	return res, nil
}

// 启动一条龙
func StartOneLong(name string) (Res, error) {

	OneLongService.StartOneLong(name)
	var res Res
	res.Time = time.Now().Format("2006-01-02")
	res.Msg = "启动一条龙" + name
	return res, nil
}
