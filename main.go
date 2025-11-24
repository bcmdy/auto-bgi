package main

import (
	"auto-bgi/menu"
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(menu.OnReady, menu.OnExit)
}

//前端打包
//cd web
//npm run build

//后端打包：
//go build

//打包脚本
//  build.bat

//go build -ldflags="-H=windowsgui"
//go build -ldflags="-H=windowsgui" -o auto-bgi无窗口.exe
