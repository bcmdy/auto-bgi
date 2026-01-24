package abgiObs

import (
	"auto-bgi/config"
	"auto-bgi/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
)

type VideoInfo struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	SizeMB       float64 `json:"sizeMB"`
	ModifiedTime string  `json:"modifiedTime"`
}

// GetAllRecordingsInfo 获取录制视频列表
// GetAllRecordingsInfo 获取指定目录下的视频信息列表
func (V *VideoInfo) GetAllRecordingsInfo(obsPath string) ([]VideoInfo, error) {
	if obsPath == "" {
		return nil, fmt.Errorf("未指定 OBS 录制路径")
	}

	var videos []VideoInfo
	err := filepath.Walk(obsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		rawSizeMB := float64(info.Size()) / 1024.0 / 1024.0
		finalSizeMB := tools.RoundFloat(rawSizeMB, 2)
		switch ext {
		case ".mp4", ".mkv", ".flv", ".mov":
			videos = append(videos, VideoInfo{
				Name:         info.Name(),
				Path:         path,
				SizeMB:       finalSizeMB,
				ModifiedTime: info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return videos, nil
}

func (V *VideoInfo) DeleteVideo(filename string) error {
	if filename == "" {
		return fmt.Errorf("未指定要删除的视频文件")
	}

	videoPath := filepath.Join(config.Cfg.ScreenRecord.ObsSavePath, filename)

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("视频文件不存在: %s", filename)
	}

	err := os.Remove(videoPath)
	if err != nil {
		return fmt.Errorf("删除视频失败: %v", err)
	}

	return nil
}

func (V *VideoInfo) DeleteAllVideo(context *gin.Context) {
	err := DeleteVideosByAge(0)
	if err != nil {
		context.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	context.JSON(200, gin.H{
		"message": "删除所有视频成功",
	})
	return

}
