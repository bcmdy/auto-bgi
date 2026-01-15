package control

import (
	"auto-bgi/autoLog"
	"github.com/itchyny/volume-go"
)

type Audio struct {
}

// 静音
func (audio *Audio) Mute() {

	err := volume.Mute()
	if err != nil {
		autoLog.Sugar.Infof("静音失败: %+v", err)

	}

}

// 取消静音
func (audio *Audio) UnMute() {
	//取消静音
	err := volume.Unmute()
	if err != nil {
		autoLog.Sugar.Infof("取消静音: %+v", err)
	}
}

// 获取当前音量
func (audio *Audio) GetVolume() int {
	v, err := volume.GetVolume()
	if err != nil {
		autoLog.Sugar.Infof("获取音量失败: %+v", err)
		return 0
	}
	return v
}

// 设置音量
func (audio *Audio) SetVolume(num int) {
	err := volume.SetVolume(num)
	if err != nil {
		autoLog.Sugar.Infof("设置音量失败: %+v", err)
	}
}
