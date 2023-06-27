package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cast"
	"github.com/zbysir/writeflow/pkg/plugin"
	"reflect"
)

type ExecFun func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error)

func (e ExecFun) Exec(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error) {
	return e(ctx, params)
}

func NewFun(fun func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error)) plugin.CMDer {
	return ExecFun(fun)
}

type LangChain struct {
}

func NewLangChain() plugin.Plugin {
	return &LangChain{}
}

func (l *LangChain) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		NameSpace: "langchain",
	}
}

func (l *LangChain) Categories() []plugin.Category {
	return []plugin.Category{
		{
			Key: "llm",
			Name: map[string]string{
				"zh-CN": "LLM",
			},
			Desc: nil,
		},
	}
}

func (l *LangChain) Components() []plugin.Component {
	return []plugin.Component{
		{
			Id:       0,
			Type:     "new_openai",
			Category: "llm",
			Data: plugin.ComponentData{
				Name: map[string]string{
					"zh-CN": "OpenAI",
				},
				Icon:        "",
				Description: map[string]string{},
				Source: plugin.ComponentSource{
					CmdType:    "builtin",
					BuiltinCmd: "new_openai",
				},
				InputParams: []plugin.NodeInputParam{
					{
						Name: map[string]string{
							"zh-CN": "ApiKey",
						},
						Key:      "api_key",
						Type:     "string",
						Optional: false,
					},
					{
						Name: map[string]string{
							"zh-CN": "BaseURL",
						},
						Key:      "base_url",
						Type:     "string",
						Optional: false,
					},
				},
				OutputAnchors: []plugin.NodeOutputAnchor{
					{
						Name: map[string]string{
							"zh-CN": "Default",
						},
						Key:  "default",
						Type: "langchain/llm",
					},
				},
			},
		},
		{
			Id:       0,
			Type:     "chat_memory",
			Category: "llm",
			Data: plugin.ComponentData{
				Name: map[string]string{
					"zh-CN": "ChatMemory",
				},
				Icon:        "",
				Description: map[string]string{},
				Source: plugin.ComponentSource{
					CmdType:    "builtin",
					BuiltinCmd: "chat_memory",
				},
				InputParams: []plugin.NodeInputParam{
					{
						Name: map[string]string{
							"zh-CN": "SessionID",
						},
						Key:      "session_id",
						Type:     "string",
						Optional: true,
					},
				},
				OutputAnchors: []plugin.NodeOutputAnchor{
					{
						Name: map[string]string{
							"zh-CN": "Default",
						},
						Key:  "default",
						Type: "langchain/chat_memory",
					},
				},
			},
		},
		{
			Type:     "langchain_call",
			Category: "llm",
			Data: plugin.ComponentData{
				Name: map[string]string{
					"zh-CN": "LangChain",
				},
				Icon:        "",
				Description: map[string]string{},
				Source: plugin.ComponentSource{
					CmdType:    "builtin",
					BuiltinCmd: "langchain_call",
				},
				InputParams: []plugin.NodeInputParam{
					{
						InputType: "anchor",
						Name: map[string]string{
							"zh-CN": "LLM",
						},
						Key:  "llm",
						Type: "langchain/llm",
					},
					{
						InputType: "anchor",
						Name: map[string]string{
							"zh-CN": "ChatMemory",
						},
						Key:      "chat_memory",
						Type:     "langchain/chat_memory",
						Optional: true,
					},
					{
						InputType: "anchor",
						Name: map[string]string{
							"zh-CN": "Functions",
						},
						Key:      "functions",
						Type:     "string",
						Optional: true,
					},
					{
						InputType: "anchor",
						Name: map[string]string{
							"zh-CN": "Prompt",
						},
						Key:  "prompt",
						Type: "string",
					},
					{
						Name: map[string]string{
							"zh-CN": "流式返回",
						},
						Value: true,
						Key:   "stream",
						Type:  "bool",
					},
				},
				OutputAnchors: []plugin.NodeOutputAnchor{
					{
						Name: map[string]string{
							"zh-CN": "Default",
						},
						Key:  "default",
						Type: "string",
					},
					{
						Name: map[string]string{
							"zh-CN": "FunctionCall",
						},
						Key:  "function_call",
						Type: "any",
					},
				},
			},
		},
	}
}

func (l *LangChain) Cmd() map[string]plugin.CMDer {
	return map[string]plugin.CMDer{
		"new_openai": NewFun(func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error) {
			key := params["api_key"].(string)
			baseUrl := cast.ToString(params["base_url"])
			config := openai.DefaultConfig(key)
			if baseUrl != "" {
				config.BaseURL = baseUrl
			}
			client := openai.NewClientWithConfig(config)
			return map[string]interface{}{"default": client}, nil
		}),
		// chat_memory 存储对话记录
		"chat_memory": NewFun(func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error) {
			idi := params["session_id"]
			if idi == nil {
				return map[string]interface{}{"default": NewMemoryChatMemory("")}, nil
			}
			id := idi.(string)

			memory := NewMemoryChatMemory(id)
			return map[string]interface{}{"default": memory}, nil
		}),
		"langchain_call": NewFun(func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error) {
			//log.Infof("langchain_call")
			openaiClient := params["llm"].(*openai.Client)
			promptI := params["prompt"]
			functionI := params["functions"]
			if promptI == nil {
				return nil, fmt.Errorf("prompt is nil")
			}
			prompt := promptI.(string)
			var functions []*openai.FunctionDefine
			if functionI != nil {
				function := functionI.(string)
				err = json.Unmarshal([]byte(function), &functions)
				if err != nil {
					return nil, err
				}
			}

			var messages []openai.ChatCompletionMessage
			var chatMemory ChatMemory
			if params["chat_memory"] != nil {
				chatMemory = params["chat_memory"].(ChatMemory)
			}

			if chatMemory != nil {
				messages = append(messages, chatMemory.GetHistory(ctx)...)
			}

			userMsg := openai.ChatCompletionMessage{Content: prompt, Role: openai.ChatMessageRoleUser}
			if chatMemory != nil {
				chatMemory.AppendHistory(ctx, userMsg)
			}
			messages = append(messages, userMsg)

			rsp2, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:            "gpt-3.5-turbo-0613",
				Messages:         messages,
				MaxTokens:        2000,
				Temperature:      0,
				TopP:             0,
				N:                0,
				Stream:           false,
				Stop:             nil,
				PresencePenalty:  0,
				FrequencyPenalty: 0,
				LogitBias:        nil,
				User:             "",
				Functions:        functions,
				FunctionCall:     "",
			})
			if err != nil {
				return nil, err
			}

			content := rsp2.Choices[0].Message.Content
			if chatMemory != nil {
				chatMemory.AppendHistory(ctx, rsp2.Choices[0].Message)
			}

			return map[string]interface{}{"default": content, "function_call": rsp2.Choices[0].Message.FunctionCall}, nil

		}),
	}
}

func (l *LangChain) GoSymbols() map[string]map[string]reflect.Value {
	return nil
}

var _ plugin.Plugin = (*LangChain)(nil)

func Register(r plugin.Register) {
	r.RegisterPlugin(NewLangChain())
}
