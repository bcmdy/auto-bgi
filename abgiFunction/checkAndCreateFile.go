package abgiFunction

import (
	"auto-bgi/autoLog"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	vbsFilePath    = "run_auto_bgi.vbs"
	batFilePath    = "run_auto_bgi_hidden.bat"
	vbsFileContent = `CreateObject("Wscript.Shell").Run "run_auto_bgi_hidden.bat", 0, False`
	batFileContent = `@echo off
		timeout /t 10 /nobreak >nul
		start "" /min "%~dp0auto-bgi.exe"`
)

// 检查文件是否存在，如果不存在则创建并写入内容
func checkAndCreateFile(filePath, content string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 如果文件不存在，创建文件并写入内容
		err := ioutil.WriteFile(filePath, []byte(content), 0644) // 0644 权限：所有者可读写，其他用户只读
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
		autoLog.Sugar.Infof("文件 %s 已创建并写入内容", filePath)
	}
	return nil
}

func check() {
	// 检查并创建 vbs 和 bat 文件
	if err := checkAndCreateFile(vbsFilePath, vbsFileContent); err != nil {
		fmt.Printf("创建文件 %s 失败: %v\n", vbsFilePath, err)
		autoLog.Sugar.Error()
		return
	}
	if err := checkAndCreateFile(batFilePath, batFileContent); err != nil {
		fmt.Printf("创建文件 %s 失败: %v\n", batFilePath, err)
		return
	}
}
