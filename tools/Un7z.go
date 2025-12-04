package tools

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed 7zr.exe
var sevenZipExe embed.FS

// 把内置的 7zr.exe 写到临时文件
func extract7zr() (string, error) {
	f, err := sevenZipExe.Open("7zr.exe")
	if err != nil {
		return "", err
	}
	defer f.Close()

	tmpFile, err := os.CreateTemp("", "7zr-*.exe")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, f); err != nil {
		return "", err
	}

	// 设置可执行权限（Windows忽略，但无害）
	_ = os.Chmod(tmpFile.Name(), 0755)

	return tmpFile.Name(), nil
}

func Un7z(src7z, dstDir string) error {
	// 先把嵌入的 7zr.exe 写出来
	zedExePath, err := extract7zr()
	if err != nil {
		return fmt.Errorf("写出7zr失败: %w", err)
	}
	defer os.Remove(zedExePath) // 用完删除

	// 创建目标目录
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	absDst, _ := filepath.Abs(dstDir)

	// 调用临时目录的 7zr
	cmd := exec.Command(zedExePath, "x", src7z, fmt.Sprintf("-o%s", absDst))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	return nil
}
