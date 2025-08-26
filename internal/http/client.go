package http

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"auto-bgi/internal/logger"
)

// Client HTTP客户端
type Client struct {
	httpClient *http.Client
	headers    map[string]string
}

// NewClient 创建新的HTTP客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}
}

// SetHeader 设置请求头
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetHeaders 批量设置请求头
func (c *Client) SetHeaders(headers map[string]string) {
	for key, value := range headers {
		c.headers[key] = value
	}
}

// GetHeaders 获取当前请求头
func (c *Client) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for key, value := range c.headers {
		headers[key] = value
	}
	return headers
}

// Get 发送GET请求
func (c *Client) Get(url string) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	logger.Debug("发送GET请求: %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	
	// 检查是否是gzip压缩
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("创建gzip读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}

// Post 发送POST请求
func (c *Client) Post(url string, data interface{}) (*Response, error) {
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化数据失败: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 如果有JSON数据，设置Content-Type
	if data != nil {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	logger.Debug("发送POST请求: %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	
	// 检查是否是gzip压缩
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("创建gzip读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// PostJSON 发送POST请求，直接使用JSON字符串
func (c *Client) PostJSON(url string, jsonData string) (*Response, error) {
	body := bytes.NewBufferString(jsonData)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	logger.Debug("发送POST JSON请求: %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	
	// 检查是否是gzip压缩
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("创建gzip读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// PostJSONWithHeaders 发送POST请求，使用指定的请求头
func (c *Client) PostJSONWithHeaders(url string, jsonData string, headers map[string]string) (*Response, error) {
	body := bytes.NewBufferString(jsonData)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	logger.Debug("发送POST JSON请求: %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	
	// 检查是否是gzip压缩
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("创建gzip读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// Response HTTP响应
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// JSON 将响应体解析为JSON
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String 获取响应体字符串
func (r *Response) String() string {
	return string(r.Body)
}

// IsSuccess 检查响应是否成功
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
} 