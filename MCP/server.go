package MCP

import (
	"auto-bgi/auth"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"sync"
)

// MCP JSON-RPC structures
type JsonRpcRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type JsonRpcResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
	ID      interface{}   `json:"id,omitempty"`
}

type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type McpTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema,omitempty"`
}

type ListToolsResult struct {
	Tools []McpTool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Session management
type Session struct {
	ID       string
	SendChan chan []byte
}

var sessions = make(map[string]*Session)
var sessionsMutex sync.Mutex

func StartMCPServer(router *gin.Engine) {
	mcpGroup := router.Group("/mcp")
	{
		mcpGroup.GET("/sse", handleSSE)

		mcpGroup.POST("/sse", handleSSE)

		// 支持路径参数 :sessionId
		mcpGroup.POST("/messages", handleMessages)
	}
}

func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback-id"
	}
	return hex.EncodeToString(bytes)
}

func handleSSE(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := generateID()
	session := &Session{
		ID:       sessionID,
		SendChan: make(chan []byte, 10),
	}

	sessionsMutex.Lock()
	sessions[sessionID] = session
	sessionsMutex.Unlock()

	defer func() {
		sessionsMutex.Lock()
		delete(sessions, sessionID)
		sessionsMutex.Unlock()
		close(session.SendChan)
	}()

	// Send endpoint event
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	absoluteURL := fmt.Sprintf("%s://%s/mcp/messages?sessionId=%s", scheme, host, sessionID)

	endpointEvent := fmt.Sprintf("event: endpoint\ndata: %s\n\n", absoluteURL)
	c.Writer.Write([]byte(endpointEvent))
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-session.SendChan:
			if !ok {
				return false
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

func handleMessages(c *gin.Context) {

	apiKey := c.GetHeader("apiKey")
	if apiKey == "" || apiKey != auth.User.Auth.APIKey {
		c.JSON(401, gin.H{"code": 401, "error": "未授权"})
		return
	}

	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "Missing sessionId"})
		return
	}

	sessionsMutex.Lock()
	session, exists := sessions[sessionID]
	sessionsMutex.Unlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Session not found"})
		return
	}

	var req JsonRpcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON-RPC request"})
		return
	}

	go processRequest(session, req)

	c.Status(http.StatusAccepted)
}

func processRequest(session *Session, req JsonRpcRequest) {
	var response JsonRpcResponse
	response.Jsonrpc = "2.0"
	response.ID = req.ID

	switch req.Method {
	case "initialize":
		response.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "abgi-info-mcp",
				"version": "1.0.0",
			},
		}
	case "notifications/initialized":
		return
	case "tools/list":
		response.Result = ListToolsResult{
			Tools: []McpTool{
				{
					Name:        "findBgiIndex",
					Description: "查询bgi当前进度",
					InputSchema: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					Name:        "StartOneLong",
					Description: "启动自动化流程。严禁私自默认选项。如果用户未指明，你必须先询问用户具体执行哪一个，严禁直接调用此工具。",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"enum":        OneLongService.OneLongAllName(),
								"description": "必须且只能从枚举值中选择。如果用户提供的名称（如'aaaa'）不在列表中，严禁擅自匹配，必须回复用户并要求其从中枚举值选择一个。",
							},
						},
					},
				},
			},
		}
	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			response.Error = &JsonRpcError{Code: -32602, Message: "Invalid params"}
			break
		}

		var result interface{}
		var err error

		switch params.Name {
		case "findBgiIndex":
			result, err = GetBgiIndex()
		case "StartOneLong":
			//参数校验
			longAllName := OneLongService.OneLongAllName()
			name := params.Arguments["name"].(string)
			//判断参数是否包含
			if !slices.Contains(longAllName, name) {
				response.Error = &JsonRpcError{Code: -32602, Message: "没有这个一条龙，你的一条龙有：" + strings.Join(longAllName, "、")}
				break
			}
			result, err = StartOneLong(params.Arguments["name"].(string))
		default:
			response.Error = &JsonRpcError{Code: -32601, Message: "Tool not found"}
			sendResponse(session, response)
			return
		}

		if err != nil {
			response.Result = CallToolResult{
				Content: []ContentItem{{Type: "text", Text: "Error: " + err.Error()}},
				IsError: true,
			}
		} else {
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")
			response.Result = CallToolResult{
				Content: []ContentItem{{Type: "text", Text: string(jsonBytes)}},
				IsError: false,
			}
		}

	default:
		if req.ID == nil {
			return
		}
		response.Error = &JsonRpcError{Code: -32601, Message: "Method not found"}
	}

	sendResponse(session, response)
}

func sendResponse(session *Session, response JsonRpcResponse) {
	respBytes, _ := json.Marshal(response)
	session.SendChan <- respBytes
}
