package config

import (
	"auto-bgi/autoLog"
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
	"strconv"
	"strings"
	"time"
)

var GameRoles GameRolesRes

// 初始化游戏角色数据
func InitA() GameRolesRes {

	// 异步获取原神游戏角色数据
	GetGenShinGameRolesAsync()

	// 打开GameInfo.json文件
	file, err := os.Open("GameInfo.json")
	if err != nil {
		return GameRoles
	}
	defer file.Close() // 确保文件在函数返回前被关闭

	// 读取文件内容
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return GameRoles
	}

	// 将JSON数据解析到GameRoles结构体中
	if err := json.Unmarshal(bytes, &GameRoles); err != nil {
		return GameRoles
	}
	return GameRoles
}

type GameRolesRes struct {
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

	var result GameRolesRes

	ApiSalt2 := "xV8v4Qu54lUKrEYFZkJhB8cuOh9Asafs"

	// 指定要请求的 URL
	url := "https://api-takumi.mihoyo.com/binding/api/getUserGameRolesByCookie?game_biz=hk4e_cn"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("cookie", BgiCfg.MiYouSheConfigCookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DS", CreateSecret2(ApiSalt2, url))
	req.Header.Set("X-Requested-With", "com.mihoyo.hyperion")
	req.Header.Set("x-rpc-app_version", "2.71.1")
	req.Header.Set("x-rpc-client_type", "5")
	req.Header.Set("Referer", "https://webstatic.mihoyo.com/")
	if err != nil {
		fmt.Printf("Error creating GET request: %v\n", err)
		return
	}

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
	req.Header.Set("cookie", BgiCfg.MiYouSheConfigCookie)
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

//// 原神签到
//func GenShinSign() {
//
//	mapData := make(map[string]interface{})
//	mapData["act_id"] = "e202311201442471"
//	mapData["region"] = GameRoles.Data.List[0].Region
//	mapData["uid"] = GameRoles.Data.List[0].GameId
//	mapData["lang"] = "zh-cn"
//	jsonData, err := json.Marshal(mapData)
//	if err != nil {
//		fmt.Println("JSON 格式化失败:", err)
//	}
//
//	// 定义请求的 URL
//	signUrl := "https://api-takumi.mihoyo.com/event/luna/sign"
//
//	req, err := http.NewRequest("POST", signUrl, bytes.NewBuffer(jsonData))
//	if err != nil {
//		fmt.Printf("Error creating POST request: %v\n", err)
//
//	}
//
//	req.Header.Set("cookie", Cfg.Cookie)
//	req.Header.Set("x-rpc-signgame", "hk4e")
//	req.Header.Set("x-rpc-client_type", "5")
//	req.Header.Set("x-rpc-app_version", "2.71.1")
//	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		fmt.Printf("Error sending POST request: %v\n", err)
//	}
//	defer resp.Body.Close()
//	body, _ := ioutil.ReadAll(resp.Body)
//
//	var res map[string]interface{}
//	err2 := json.Unmarshal(body, &res)
//	if err2 != nil {
//		autoLog.Sugar.Errorf("签到转换错误: %v\n", err)
//	}
//	if res["message"].(string) == "OK" {
//		autoLog.Sugar.Infof("原神签到成功")
//	} else if res["message"].(string) == "已签到" {
//		autoLog.Sugar.Infof("原神签到成功")
//	} else {
//		autoLog.Sugar.Errorf("原神签到失败: %v\n", res["message"].(string))
//	}
//}

type ChildNode struct {
	ID         int64
	Name       string
	Icon       string
	ParentID   int64
	ParentName string
	Number     int64
}

// 定义API返回的JSON结构（适配接口返回格式）
type TreeResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		Tree []struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			Icon     string `json:"icon"`
			Children []struct {
				ID       int64  `json:"id"`
				Name     string `json:"name"`
				Icon     string `json:"icon"`
				ParentId int64  `json:"parent_id"`
			} `json:"children"`
			// 其他字段可根据实际返回补充，这里只保留核心字段
		} `json:"tree"`
	} `json:"data"`
}

type TreeResponse2 struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		MaterialInfo map[string]int64 `json:"material_info"`
	} `json:"data"`
}

// 获取背包信息
func GetBagInfo() (map[string][]ChildNode, map[string]int64) {
	// 捕获异常
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("我的背包: %v\n", r)
		}
	}()

	// 1. 构建请求
	req, err := http.NewRequest("GET", "https://waf-api-takumi.mihoyo.com/common/map_user/ys_obc/v2/map/label/tree?map_id=2&app_sn=ys_obc&lang=zh-cn", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return nil, nil
	}

	// 设置请求头（补充必要的User-Agent，提升兼容性）
	req.Header.Set("cookie", BgiCfg.MiYouSheConfigCookie)
	req.Header.Set("Referer", "https://webstatic.mihoyo.com/")
	req.Header.Set("X-Requested-With", "com.mihoyo.hyperion")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 2. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return nil, nil
	}
	defer resp.Body.Close()

	// 3. 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求返回非200状态码: %d\n", resp.StatusCode)
		return nil, nil
	}

	// 4. 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应体失败: %v\n", err)
		return nil, nil
	}

	// 可选：打印原始响应（调试用）
	// fmt.Println("原始响应数据:", string(body))

	// 5. 解析JSON响应
	var treeResp TreeResponse
	err = json.Unmarshal(body, &treeResp)
	if err != nil {
		fmt.Printf("JSON解析失败: %v\n", err)
		return nil, nil
	}

	// 6. 检查接口返回状态
	if treeResp.Retcode != 0 {
		fmt.Printf("接口返回错误: retcode=%d, message=%s\n", treeResp.Retcode, treeResp.Message)
		return nil, nil
	}

	// 7. 转换为ChildNode结构体列表
	var childNodes []ChildNode
	for _, node := range treeResp.Data.Tree {
		for _, child := range node.Children {
			//fmt.Println(child)
			childNodes = append(childNodes, ChildNode{
				ID:         child.ID,
				Name:       child.Name,
				Icon:       child.Icon,
				ParentID:   node.ID,
				ParentName: node.Name,
			})

		}
	}

	//数量请求
	// 1. 构建请求
	req2, err2 := http.NewRequest("GET", "https://api-takumi.mihoyo.com/common/map_user/ys_obc/v1/user/sync_game_material_info?map_id=2&app_sn=ys_obc&lang=zh-cn", nil)
	if err2 != nil {
		fmt.Printf("创建请求失败: %v\n", err2)
		return nil, nil
	}

	// 设置请求头（补充必要的User-Agent，提升兼容性）
	req2.Header.Set("cookie", BgiCfg.MiYouSheConfigCookie)
	req2.Header.Set("Referer", "https://webstatic.mihoyo.com/")
	req2.Header.Set("X-Requested-With", "com.mihoyo.hyperion")
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 2. 发送请求
	client2 := &http.Client{}
	resp2, err2 := client2.Do(req2)
	if err2 != nil {
		fmt.Printf("发送请求失败: %v\n", err2)
		return nil, nil
	}
	defer resp2.Body.Close()

	// 3. 检查响应状态码
	if resp2.StatusCode != http.StatusOK {
		fmt.Printf("请求返回非200状态码: %d\n", resp2.StatusCode)
		return nil, nil
	}

	// 4. 读取响应体
	body2, err2 := ioutil.ReadAll(resp2.Body)
	if err2 != nil {
		fmt.Printf("读取响应体失败: %v\n", err2)
		return nil, nil
	}
	var treeResp2 TreeResponse2

	err = json.Unmarshal(body2, &treeResp2)
	if err != nil {
		fmt.Printf("JSON解析失败: %v\n", err)
		return nil, nil
	}

	mapData := make(map[string][]ChildNode)

	m := make(map[string]int64)

	//var BackpackRecords []models.BackpackRecord
	// 8. 输出解析结果（可选，验证用）
	fmt.Printf("成功解析到 %d 个背包标签节点\n", len(childNodes))
	for _, cn := range childNodes {
		aa := treeResp2.Data.MaterialInfo[strconv.FormatInt(cn.ID, 10)]
		cn.Number = aa
		if cn.ParentName == "区域特产" || cn.ParentName == "木材" || cn.ParentName == "矿物" || cn.ParentName == "背包/素材" || cn.ParentName == "贵重收集物" {
			mapData[cn.ParentName] = append(mapData[cn.ParentName], cn)
			m[cn.Name] += aa
		}
	}

	return mapData, m
}

// 查询实时便签
func DailyNote() error {

	// 捕获异常
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("我的背包: %v\n", r)
		}
	}()

	// 1. 构建请求
	req, err := http.NewRequest("GET", "https://api-takumi-record.mihoyo.com/game_record/app/genshin/api/dailyNote?role_id=103740894&server=cn_gf01", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return nil
	}

	ApiSalt2 := "xV8v4Qu54lUKrEYFZkJhB8cuOh9Asafs"

	secret2 := CreateSecret2(ApiSalt2, "https://api-takumi-record.mihoyo.com/game_record/app/genshin/api/dailyNote?role_id=103740894&server=cn_gf01")

	fmt.Println("======secret2=====", secret2)

	// 设置请求头（补充必要的User-Agent，提升兼容性）
	req.Header.Set("cookie", "ltuid=165421629;ltoken=v2_G3_RKbtjHRXW2OhX2vebT8IAXOGesJvDiILoQqsKRXjGNKn7ReCKaUaGXWQqXxax-IremmSwEjdGiHqO3exuqM3-qY5iTJiKjHZjE23JBgdo7PJnlJVBH5s=.CAE=;stuid=165421629;mid=0o3x3blmnx_mhy;stoken=v2_9BHACVTFvdxA5OHtW7oXfo5O-bZHa_eJZLIBxH11Z4AiwptSBf1lXepGYvQGjpFqw0Dye8jJqrL06E2l-xbBfyTKwvccEbDIzaZeaJgs1xCQ-6mB_9UKbhg=.CAE=;device_id=f35e28cc-c3a7-3774-82b2-e62e9e8574a3;device_fp=38d80d7f09968")
	req.Header.Set("DS", secret2)
	req.Header.Set("Referer", "https://api-takumi-record.mihoyo.com/")
	req.Header.Set("X-Requested-With", "api-takumi-record.mihoyo.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 2. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	// 3. 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求返回非200状态码: %d\n", resp.StatusCode)
		return nil
	}

	// 4. 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应体失败: %v\n", err)
		return nil
	}

	// 5. 解析JSON响应
	fmt.Println("==========asdasaas==============", string(body))

	return nil

}
