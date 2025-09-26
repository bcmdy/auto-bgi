package CDAwareAutoGather

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UidInfo struct {
	UID               string              // 账户id
	CDAwareAutoGather []CDAwareAutoGather // 文件详情
}

type CDAwareAutoGather struct {
	TextName string   // txt名称
	Detail   []Detail // 文件详情
}

type Detail struct {
	FileName  string // 文件名
	CDTime    string // CD时间
	CDExpired bool   // CD是否到期
}

// ReadInfo 读取采集cd数据
// ReadInfo 函数用于读取并整理记录信息
// 返回一个包含所有记录的 UidInfo 切片
func (u *UidInfo) ReadInfo(status string) []UidInfo {
	// 构建基础目录路径，指向记录文件所在目录
	baseDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "CD-Aware-AutoGather", "record")
	var allRecords []UidInfo

	// 遍历 record 目录下的所有子文件夹（账号）
	subDirs, err := os.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}

	for _, sub := range subDirs {
		if !sub.IsDir() {
			continue
		}
		uid := sub.Name()

		var uidInfo UidInfo
		uidInfo.UID = uid
		uidInfo.CDAwareAutoGather = []CDAwareAutoGather{}

		accountDir := filepath.Join(baseDir, uid)

		// 遍历该账号文件夹下的所有 txt 文件
		err := filepath.Walk(accountDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
				record := readAndFormatTxt(path, info.Name(), status)
				if record != nil {
					uidInfo.CDAwareAutoGather = append(uidInfo.CDAwareAutoGather, *record)
				}
			}
			return nil
		})
		if err != nil {

			autoLog.Sugar.Errorf("CD遍历账号文件夹出错:%s,%d", uid, err)
		}

		allRecords = append(allRecords, uidInfo)
	}

	// 输出查看
	return allRecords
}

func readAndFormatTxt(filePath string, txtName, status string) *CDAwareAutoGather {
	file, err := os.Open(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("打开文件失败:%s", err)
		return nil
	}
	defer file.Close()

	cda := CDAwareAutoGather{
		TextName: txtName,
		Detail:   []Detail{},
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		filename := parts[0]
		timeStr := parts[len(parts)-1]

		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			autoLog.Sugar.Errorf("解析时间失败:%v", err)
			continue
		}

		detail := Detail{
			FileName:  filename,
			CDTime:    t.Format("2006-01-02 15:04:05"),
			CDExpired: t.Before(time.Now()),
		}

		// 根据 status 筛选
		if (status == "1" && detail.CDExpired) ||
			(status == "2" && !detail.CDExpired) ||
			(status == "3") {
			cda.Detail = append(cda.Detail, detail)
		}
	}

	if err := scanner.Err(); err != nil {
		autoLog.Sugar.Errorf("读取文件出错:%v", err)
	}

	return &cda
}
