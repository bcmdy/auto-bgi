package control

import (
	"auto-bgi/autoLog"
	"github.com/go-ole/go-ole"
)
import "github.com/moutend/go-wca/pkg/wca"

type Audio struct {
}

// SetSystemMute 设置系统默认输出设备的静音状态
// mute = true  → 静音
// mute = false → 取消静音
func setSystemMute(mute bool) error {
	// 初始化 COM
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	// 创建设备枚举器
	var enumerator *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&enumerator,
	); err != nil {
		return err
	}
	defer enumerator.Release()

	// 获取默认输出设备（扬声器/耳机）
	var device *wca.IMMDevice
	if err := enumerator.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &device); err != nil {
		return err
	}
	defer device.Release()

	// 激活音量控制接口
	var endpointVol *wca.IAudioEndpointVolume
	if err := device.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &endpointVol); err != nil {
		return err
	}
	defer endpointVol.Release()

	// 设置静音
	if err := endpointVol.SetMute(mute, nil); err != nil {
		return err
	}

	return nil
}

// 静音
func (audio *Audio) Mute() {

	err := setSystemMute(true)
	if err != nil {
		autoLog.Sugar.Errorf("静音失败%v", err)
	}

}

// 取消静音
func (audio *Audio) UnMute() {
	err := setSystemMute(false)
	if err != nil {

		autoLog.Sugar.Errorf("取消静音失败%v", err)
	}
}
