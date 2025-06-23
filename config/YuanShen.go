package config

import (
	"auto-bgi/autoLog"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

var GameRoles gameRolesRes

func init() {

	GetGenShinGameRolesAsync()

	file, err := os.Open("GameInfo.json")
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	if err := json.Unmarshal(bytes, &GameRoles); err != nil {
		return
	}
}

type gameRolesRes struct {
	RetCode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		List []struct {
			GameId     string `json:"game_uid"`
			Region     string `json:"region"`
			GameRoleId string `json:"game_role_id"`
			GameBiz    string `json:"game_biz"`
			NicName    string `json:"nickname"`
			Level      int    `json:"level"`
			IsChosen   bool   `json:"is_chosen"`
			RegionName string `json:"region_name"`
		} `json:"list"`
	} `json:"data"`
}

func CreateSecret2(apiSalt2, urlStr string) string {
	t := time.Now().Unix()
	r := rand.Intn(100000) + 100000 // 100000-199999
	b := ""
	q := ""

	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		// handle error if necessary
		return ""
	}

	if parsedUrl.RawQuery != "" {
		queryParams := strings.Split(parsedUrl.RawQuery, "&")
		sort.Strings(queryParams)
		q = strings.Join(queryParams, "&")
	}

	data := fmt.Sprintf("salt=%s&t=%d&r=%d&b=%s&q=%s", apiSalt2, t, r, b, q)
	hashBytes := md5.Sum([]byte(data))
	check := hex.EncodeToString(hashBytes[:])

	result := fmt.Sprintf("%d,%d,%s", t, r, check)
	return result
}

// 获取原神账号信息
func GetGenShinGameRolesAsync() {

	var result gameRolesRes

	ApiSalt2 := "xV8v4Qu54lUKrEYFZkJhB8cuOh9Asafs"

	// 指定要请求的 URL
	url := "https://api-takumi.mihoyo.com/binding/api/getUserGameRolesByCookie?game_biz=hk4e_cn"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating POST request: %v\n", err)
		return
	}
	req.Header.Set("cookie", Cfg.Cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DS", CreateSecret2(ApiSalt2, url))
	req.Header.Set("X-Requested-With", "com.mihoyo.hyperion")
	req.Header.Set("x-rpc-app_version", "2.71.1")
	req.Header.Set("x-rpc-client_type", "5")
	req.Header.Set("Referer", "https://webstatic.mihoyo.com/")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending POST request: %v\n", err)
		return
	}
	defer resp.Body.Close() // 请求完成后关闭响应体
	body, _ := ioutil.ReadAll(resp.Body)

	//转成GameRolesRes
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("原神获取信息转换错误: %v\n", err)
		return
	}

	// 保存游戏角色信息到文件
	// 打开或创建文件
	file, err := os.OpenFile("GameInfo.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("game.json打开失败:", err)
	}
	defer file.Close()
	// 写入 JSON 数据到文件
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("JSON 格式化失败:", err)
	}

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("写入文件失败:", err)
	}

	return
}

type TravelsDiaryDetail struct {
	Uid       int                      `json:"uid"`
	Region    string                   `json:"region"`
	AccountId int                      `json:"account_id"`
	Nickname  string                   `json:"nickname"`
	Date      string                   `json:"date"`
	List      []TravelsDiaryDetailList `json:"list"`
}

type TravelsDiaryDetailList struct {
	ActionID int    `json:"action_id"`
	Action   string `json:"action"`
	Time     string `json:"time"`
	Num      int    `json:"num"`
}

// 旅行札记收入详情
func GetTravelsDiaryDetailAsync(month int, type_ int, page int) (TravelsDiaryDetail, error) {

	//捕获异常
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("旅行札记收入详情异常详情: %v\n", r)
		}
	}()

	GetTravelsDiaryDetailUrl := fmt.Sprintf("https://hk4e-api.mihoyo.com/event/ys_ledger/monthDetail?"+
		"page=%d"+
		"&month=%d"+
		"&limit=100"+
		"&type=%d"+
		"&bind_uid=%s"+
		"&bind_region=%s"+
		"&bbs_presentation_style=fullscreen&bbs_auth_required=true&utm_source=bbs&utm_medium=mys&utm_campaign=icon",
		page, month, type_, GameRoles.Data.List[0].GameId, GameRoles.Data.List[0].Region)

	req, err := http.NewRequest("GET", GetTravelsDiaryDetailUrl, nil)
	if err != nil {
		fmt.Printf("请求接口，POST request: %v\n", err)
		return TravelsDiaryDetail{}, err
	}
	req.Header.Set("cookie", Cfg.Cookie)
	req.Header.Set("Referer", "https://webstatic.mihoyo.com/")
	req.Header.Set("X-Requested-With", "com.mihoyo.hyperion")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending POST request: %v\n", err)
		return TravelsDiaryDetail{}, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// 定义临时结构体来解析到 data 这一层
	var res map[string]interface{}

	err2 := json.Unmarshal(body, &res)
	if err2 != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return TravelsDiaryDetail{}, err
	}
	data := res["data"]
	//转成TravelsDiaryDetail
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON 格式化失败:", err)
		return TravelsDiaryDetail{}, err
	}
	var travelsDiaryDetail TravelsDiaryDetail
	err = json.Unmarshal(jsonData, &travelsDiaryDetail)
	if err != nil {
		fmt.Println("JSON 格式化失败:", err)
		return TravelsDiaryDetail{}, err
	}

	return travelsDiaryDetail, nil

}

// 根据时间过滤数据
func FilterByTime(data []TravelsDiaryDetailList, start, end string) []TravelsDiaryDetailList {
	var result []TravelsDiaryDetailList

	// 定义时间格式
	layout := "2006-01-02 15:04:05"

	// 解析起始和结束时间
	startTime, err := time.Parse(layout, start)
	if err != nil {
		fmt.Println("起始时间解析错误:", err)
		return result
	}
	endTime, err := time.Parse(layout, end)
	if err != nil {
		fmt.Println("结束时间解析错误:", err)
		return result
	}

	for _, item := range data {
		itemTime, err := time.Parse(layout, item.Time)
		if err != nil {
			// 时间格式不正确，跳过
			continue
		}
		if !itemTime.Before(startTime) && !itemTime.After(endTime) {
			result = append(result, item)
		}
	}
	return result
}

// 原神签到
func GenShinSign() {

	mapData := make(map[string]interface{})
	mapData["act_id"] = "e202311201442471"
	mapData["region"] = GameRoles.Data.List[0].Region
	mapData["uid"] = GameRoles.Data.List[0].GameId
	mapData["lang"] = "zh-cn"
	jsonData, err := json.Marshal(mapData)
	if err != nil {
		fmt.Println("JSON 格式化失败:", err)
	}

	// 定义请求的 URL
	signUrl := "https://api-takumi.mihoyo.com/event/luna/sign"

	req, err := http.NewRequest("POST", signUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating POST request: %v\n", err)

	}

	req.Header.Set("cookie", Cfg.Cookie)
	req.Header.Set("x-rpc-signgame", "hk4e")
	req.Header.Set("x-rpc-client_type", "5")
	req.Header.Set("x-rpc-app_version", "2.71.1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending POST request: %v\n", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var res map[string]interface{}
	err2 := json.Unmarshal(body, &res)
	if err2 != nil {
		autoLog.Sugar.Errorf("签到转换错误: %v\n", err)
	}
	if res["message"].(string) == "OK" {
		autoLog.Sugar.Infof("原神签到成功")
	} else if res["message"].(string) == "已签到" {
		autoLog.Sugar.Infof("原神签到成功")
	} else {
		autoLog.Sugar.Errorf("原神签到失败: %v\n", res["message"].(string))
	}
}

// GetCookieHeader 接收一个 cookie 字符串，返回一个 map[string]string，表示键值对
func GetCookieHeader() map[string]string {
	result := make(map[string]string)

	// 按分号分割 cookie 条目
	pairs := strings.Split(Cfg.Cookie, ";")
	for _, pair := range pairs {
		// 去除前后空格
		trimmedPair := strings.TrimSpace(pair)
		// 按等号分割键值
		kv := strings.Split(trimmedPair, "=")
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}
