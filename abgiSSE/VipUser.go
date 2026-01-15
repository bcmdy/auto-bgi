package abgiSSE

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MoveUserRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	NewGroupID string `json:"new_group_id" binding:"required"`
	SecretKey  string `json:"secret_key" binding:"required"`
}

// 妙妙屋修改用户房间
func ModifyRoom(NewGroupID, SecretKey string) {
	host, err := tools.Decrypt(config.Cfg.Account.SecretKey)
	if err != nil {
		autoLog.Sugar.Error("无法连接")
		return
	}

	payload := map[string]string{
		"uid":            config.Cfg.Account.Uid,
		"new_group_name": NewGroupID,
		"secret_key":     SecretKey,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		autoLog.Sugar.Error("JSON序列化失败:", err)
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://%s/api/vip/ModifyUserRoom", host)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		autoLog.Sugar.Errorf("切换房间失败: %v", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		autoLog.Sugar.Errorf("切换房间失败 %s", string(body))
		return
	}

	body, _ := io.ReadAll(resp.Body)
	autoLog.Sugar.Infof("更新成功，返回内容: %s", string(body))
}
