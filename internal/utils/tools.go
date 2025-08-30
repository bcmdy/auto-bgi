package utils

import (
	"auto-bgi/autoLog"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	httpClient "auto-bgi/internal/http"
	"auto-bgi/internal/mysConfig"
)

const (
	// 米游社签名盐值 - 与Python版本保持一致
	MihoyobbsSalt       = "ss6ZUzUlaWv6iDe0SHPSdCZYr0RSKPdi"
	MihoyobbsSaltWeb    = "gW20AtTxpc0V5GR3SmsytCLhVBgXtW6I"
	MihoyobbsSaltX4     = "xV8v4Qu54lUKrEYFZkJhB8cuOh9Asafs"
	MihoyobbsSaltX6     = "t0qEgfub6cvueAPgR5m9aQWWVciEer7v"
	MihoyobbsVersion    = "2.85.1"
	MihoyobbsClientType = "2"
)

// InitRandom 初始化随机数种子
func InitRandom() {
	rand.Seed(time.Now().UnixNano())
}

// MD5 计算MD5哈希
func MD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// RandomText 生成指定长度的随机文本
func RandomText(num int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, num)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// Timestamp 获取当前时间戳
func Timestamp() int64 {
	return time.Now().Unix()
}

// GetDS 获取米游社的签名字符串
func GetDS(web bool) string {
	var salt string
	if web {
		salt = MihoyobbsSaltWeb
	} else {
		salt = MihoyobbsSalt
	}

	t := strconv.FormatInt(Timestamp(), 10)
	r := RandomText(6)
	c := MD5(fmt.Sprintf("salt=%s&t=%s&r=%s", salt, t, r))
	return fmt.Sprintf("%s,%s,%s", t, r, c)
}

// GetDS2 获取米游社的签名字符串（带查询参数和请求体）
func GetDS2(query, body string) string {
	t := strconv.FormatInt(Timestamp(), 10)
	r := strconv.Itoa(rand.Intn(100000) + 100001)
	c := MD5(fmt.Sprintf("salt=%s&t=%s&r=%s&b=%s&q=%s", MihoyobbsSaltX6, t, r, body, query))
	return fmt.Sprintf("%s,%s,%s", t, r, c)
}

// GetDeviceID 使用cookie生成设备ID
func GetDeviceID(cookie string) string {
	// 简化版本，使用cookie的MD5作为设备ID
	return MD5(cookie)[:16]
}

// GetItem 获取签到奖励信息
func GetItem(rawData map[string]interface{}) string {
	name, _ := rawData["name"].(string)
	cnt, _ := rawData["cnt"].(float64)
	return fmt.Sprintf("「%s」x%.0f", name, cnt)
}

// GetNextDayTimestamp 获取明天凌晨的时间戳
func GetNextDayTimestamp() int64 {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location()).Unix()
}

// GetUserAgent 获取用户代理字符串
func GetUserAgent(baseUA string) string {
	if baseUA == "" {
		baseUA = "Mozilla/5.0 (Linux; Android 12; Unspecified Device) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/103.0.5060.129 Mobile Safari/537.36"
	}
	return fmt.Sprintf("%s miHoYoBBS/%s", baseUA, MihoyobbsVersion)
}

// ExtractUIDFromCookie 从cookie中提取UID，与Python版本保持一致
func ExtractUIDFromCookie(cookie string) string {
	// 匹配Python中的正则：(account_id|ltuid|login_uid|ltuid_v2|account_id_v2)=(\d+)
	patterns := []string{
		`account_id=(\d+)`,
		`ltuid=(\d+)`,
		`login_uid=(\d+)`,
		`ltuid_v2=(\d+)`,
		`account_id_v2=(\d+)`,
	}

	for _, pattern := range patterns {
		if strings.Contains(cookie, pattern[:strings.Index(pattern, "=")]) {
			start := strings.Index(cookie, pattern[:strings.Index(pattern, "=")])
			if start >= 0 {
				part := cookie[start:]
				end := strings.Index(part, ";")
				if end == -1 {
					end = len(part)
				}
				keyValue := part[:end]
				if strings.Contains(keyValue, "=") {
					value := keyValue[strings.Index(keyValue, "=")+1:]
					return value
				}
			}
		}
	}

	return ""
}

// ExtractMIDFromCookie 从cookie中提取MID，与Python版本保持一致
func ExtractMIDFromCookie(cookie string) string {
	// 匹配Python中的正则：(account_mid_v2|ltmid_v2|mid)=(.*?)(?:;|$)
	patterns := []string{"account_mid_v2", "ltmid_v2", "mid"}

	for _, pattern := range patterns {
		if strings.Contains(cookie, pattern+"=") {
			start := strings.Index(cookie, pattern+"=")
			if start >= 0 {
				part := cookie[start:]
				end := strings.Index(part, ";")
				if end == -1 {
					end = len(part)
				}
				keyValue := part[:end]
				if strings.Contains(keyValue, "=") {
					value := keyValue[strings.Index(keyValue, "=")+1:]
					return value
				}
			}
		}
	}

	return ""
}

// ExtractStokenFromCookie 从完整的stoken cookie字符串中提取stoken值
func ExtractStokenFromCookie(stokenCookie string) string {
	// 查找stoken=后面的值
	if strings.Contains(stokenCookie, "stoken=") {
		start := strings.Index(stokenCookie, "stoken=")
		if start >= 0 {
			part := stokenCookie[start+7:] // 跳过"stoken="
			end := strings.Index(part, ";")
			if end == -1 {
				end = len(part)
			}
			return part[:end]
		}
	}
	return ""
}

// GetStokenCookie 获取带stoken的cookie，与Python版本保持一致
func GetStokenCookie() string {
	stoken := mysConfig.GlobalConfig.Account.Stoken

	// 如果stoken字段包含完整的cookie字符串，提取正确的stoken
	if strings.Contains(stoken, "stuid=") && strings.Contains(stoken, "stoken=") {
		// 从完整的stoken cookie中提取stuid和stoken
		stuid := ExtractUIDFromCookie(stoken)
		actualStoken := ExtractStokenFromCookie(stoken)
		mid := ExtractMIDFromCookie(stoken)

		cookie := fmt.Sprintf("stuid=%s;stoken=%s", stuid, actualStoken)
		if mid != "" {
			cookie += fmt.Sprintf(";mid=%s", mid)
		}
		return cookie
	}

	// 如果没有stoken，返回普通cookie
	if stoken == "" {
		return TidyCookie(mysConfig.GlobalConfig.Account.Cookie)
	}

	stuid := mysConfig.GlobalConfig.Account.Stuid
	mid := mysConfig.GlobalConfig.Account.Mid

	// 如果stuid为空，尝试从cookie中提取
	if stuid == "" {
		stuid = ExtractUIDFromCookie(mysConfig.GlobalConfig.Account.Cookie)
	}

	// 如果mid为空且stoken以v2_开头，尝试从cookie中提取
	if mid == "" && strings.HasPrefix(stoken, "v2_") {
		mid = ExtractMIDFromCookie(mysConfig.GlobalConfig.Account.Cookie)
	}

	cookie := fmt.Sprintf("stuid=%s;stoken=%s", stuid, stoken)

	// 如果stoken以v2_开头，需要添加mid参数
	if strings.HasPrefix(stoken, "v2_") {
		if mid != "" {
			cookie += fmt.Sprintf(";mid=%s", mid)
		}
	}

	return cookie
}

// TidyCookie 整理cookie格式，与Python版本保持一致
func TidyCookie(cookies string) string {
	if cookies == "" {
		return cookies
	}

	cookieDict := make(map[string]string)
	splitCookies := strings.Split(cookies, ";")

	if len(splitCookies) < 2 {
		return cookies
	}

	for _, cookie := range splitCookies {
		cookie = strings.TrimSpace(cookie)
		if cookie == "" {
			continue
		}
		parts := strings.SplitN(cookie, "=", 2)
		if len(parts) == 2 {
			cookieDict[parts[0]] = parts[1]
		}
	}

	var result []string
	for key, value := range cookieDict {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(result, "; ")
}

// ParseCookie 解析cookie字符串为map
func ParseCookie(cookieStr string) map[string]string {
	cookies := make(map[string]string)
	pairs := strings.Split(cookieStr, ";")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
		}
	}

	return cookies
}

// BuildCookieString 将cookie map构建为字符串
func BuildCookieString(cookies map[string]string) string {
	var pairs []string
	for key, value := range cookies {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(pairs, "; ")
}

// ContainsString 检查字符串切片是否包含指定字符串
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RandomSleep 随机等待一段时间
func RandomSleep(min, max int) {
	duration := rand.Intn(max-min+1) + min
	time.Sleep(time.Duration(duration) * time.Second)
}

// UpdateCookieToken 通过stoken刷新cookie_token
func UpdateCookieToken() error {
	autoLog.Sugar.Infof("米游社-尝试刷新CookieToken")

	// 检查stoken是否存在
	if mysConfig.GlobalConfig.Account.Stoken == "" {
		return fmt.Errorf("Stoken 为空，无法自动更新 CookieToken")
	}

	// 获取stuid，如果配置中为空，尝试从stoken中提取
	stuid := mysConfig.GlobalConfig.Account.Stuid
	if stuid == "" {
		stuid = ExtractUIDFromCookie(mysConfig.GlobalConfig.Account.Stoken)
		if stuid == "" {
			return fmt.Errorf("无法从Stoken中提取Stuid，请手动设置Stuid")
		}
	}

	// 构建stoken cookie
	stokenCookie := fmt.Sprintf("stuid=%s;stoken=%s", stuid, ExtractStokenFromCookie(mysConfig.GlobalConfig.Account.Stoken))
	mid := ExtractMIDFromCookie(mysConfig.GlobalConfig.Account.Stoken)
	if mid != "" {
		stokenCookie += fmt.Sprintf(";mid=%s", mid)
	}

	// 使用自定义HTTP客户端
	httpClient := httpClient.NewClient()

	// 设置请求头
	headers := map[string]string{
		"DS":                GetDS(false), // 使用非web版本的DS
		"x-rpc-app_version": MihoyobbsVersion,
		"User-Agent":        "okhttp/4.9.3",
		"x-rpc-client_type": MihoyobbsClientType,
		"Referer":           "https://webstatic.mihoyo.com",
		"Origin":            "https://webstatic.mihoyo.com",
		"x-rpc-device_id":   mysConfig.GlobalConfig.Device.ID,
		"Content-Type":      "application/json; charset=UTF-8",
		"Accept-Encoding":   "gzip, deflate",
		"Cookie":            stokenCookie,
	}

	httpClient.SetHeaders(headers)

	// 构建请求
	url := "https://api-takumi.mihoyo.com/auth/api/getCookieAccountInfoBySToken"
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}

	// 解析响应
	var response struct {
		Retcode int    `json:"retcode"`
		Message string `json:"message"`
		Data    struct {
			CookieToken string `json:"cookie_token"`
		} `json:"data"`
	}

	if err := resp.JSON(&response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 打印响应内容用于调试

	autoLog.Sugar.Infof("米游社-UpdateCookieToken响应: %s", resp.String())

	if response.Retcode != 0 {

		autoLog.Sugar.Errorf("米游社-stoken 已失效，请重新抓取 cookie")
		return fmt.Errorf("stoken 已失效: %s", response.Message)
	}

	// 更新cookie中的cookie_token
	newCookieToken := response.Data.CookieToken
	oldCookie := mysConfig.GlobalConfig.Account.Cookie

	// 使用正则表达式替换cookie_token
	re := regexp.MustCompile(`cookie_token=([^;]+)`)
	if re.MatchString(oldCookie) {
		// 如果找到旧的cookie_token，替换它
		newCookie := re.ReplaceAllString(oldCookie, fmt.Sprintf("cookie_token=%s", newCookieToken))
		mysConfig.GlobalConfig.Account.Cookie = newCookie

		autoLog.Sugar.Infof("米游社-CookieToken 刷新成功")
		return nil
	} else {
		// 如果没有找到旧的cookie_token，添加新的
		if strings.TrimSpace(oldCookie) != "" {
			mysConfig.GlobalConfig.Account.Cookie = oldCookie + "; cookie_token=" + newCookieToken
		} else {
			mysConfig.GlobalConfig.Account.Cookie = "cookie_token=" + newCookieToken
		}

		autoLog.Sugar.Infof("米游社-CookieToken 添加成功")
		return nil
	}
}
