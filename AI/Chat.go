package AI

import (
	"auto-bgi/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"io"
	"net/http"
)

var model *ark.ChatModel

func init() {

	var err error

	ctx := context.Background()
	model, err = ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  config.Cfg.AIConfig.APIKey,
		Model:   config.Cfg.AIConfig.Model,
		BaseURL: "https://api.deepseek.com/v1",
	})
	if err != nil {
		panic(err)
	}

}

func GetModel() *ark.ChatModel {
	return model
}

// StreamChat 流式输出到 http.ResponseWriter（SSE）
func StreamChat(ctx context.Context, messages []*schema.Message, w http.ResponseWriter) error {
	if model == nil {
		return errors.New("model not initialized")
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	sr, err := model.Stream(ctx, messages)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return err
	}
	defer sr.Close()

	for {
		recv, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Fprintf(w, "event: done\ndata: [EOF]\n\n")
			flusher.Flush()
			break
		}
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
			flusher.Flush()
			return err
		}

		payload := map[string]string{
			"role":    string(recv.Role),
			"content": recv.Content,
		}
		b, _ := json.Marshal(payload)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}
