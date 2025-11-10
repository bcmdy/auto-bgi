package abgiAi

import (
	"auto-bgi/config"
	"context"
	"fmt"
	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
	"trpc.group/trpc-go/trpc-agent-go/runner"
)

var abgiAgent runner.Runner

func InitAi() {
	// Create model.
	modelInstance := openai.New(
		config.Cfg.AbgiAiConfig.Model,
		openai.WithAPIKey(config.Cfg.AbgiAiConfig.ApiKey),
		openai.WithBaseURL(config.Cfg.AbgiAiConfig.BaseURL),
	)

	// Enable streaming output.
	genConfig := model.GenerationConfig{
		Stream: false,
	}

	agent := llmagent.New("assistant",
		llmagent.WithModel(modelInstance),
		//llmagent.WithTools(createFunctionTool()),
		llmagent.WithGenerationConfig(genConfig),
	)

	// Create Runner.
	abgiAgent = runner.NewRunner("calculator-app", agent)

}

// 对话
func Conversation(ask string) (string, error) {
	ctx := context.Background()
	events, err := abgiAgent.Run(ctx,
		"user-001",
		"session-001",
		model.NewUserMessage("你是一个机器人，你的名字叫老王，你必须使用工具来回答我的问题，我的问题是："+ask))

	if err != nil {
		return "", err
	}

	for event := range events {
		if event.Response != nil && len(event.Response.Choices) > 0 {
			fmt.Println(event.Response.Choices[0].Message.Content)
			return event.Response.Choices[0].Message.Content, nil
		}
	}
	fmt.Println()
	return "", fmt.Errorf("对话失败")
}

func JsConversation(ask string) (string, error) {
	ctx := context.Background()
	events, err := abgiAgent.Run(ctx,
		"user-001",
		"session-001",
		model.NewUserMessage(ask))

	if err != nil {
		return "", err
	}

	for event := range events {
		if event.Response != nil && len(event.Response.Choices) > 0 {
			fmt.Println(event.Response.Choices[0].Message.Content)
			return event.Response.Choices[0].Message.Content, nil
		}
	}
	fmt.Println()
	return "", fmt.Errorf("对话失败")
}
