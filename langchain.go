package main

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"github.com/zbysir/writeflow/pkg/export"
	"github.com/zbysir/writeflow_plugin_llm/sashabaranov"
	"github.com/zbysir/writeflow_plugin_llm/util"
	"reflect"
)

type PluginLLM interface {
	NewOpenAICmd() export.CMDer
	CallOpenAICmd() export.CMDer
	SupportStream() bool
}

type LangChain struct {
	pluginLLM PluginLLM
}

func NewLangChain(pluginLLM PluginLLM) export.Plugin {
	return &LangChain{pluginLLM: pluginLLM}
}

func (l *LangChain) Info() export.PluginInfo {
	return export.PluginInfo{
		NameSpace: "langchain",
	}
}

func (l *LangChain) Categories() []export.Category {
	return []export.Category{
		{
			Key: "llm",
			Name: map[string]string{
				"zh-CN": "LLM",
			},
			Desc: nil,
		},
	}
}

func (l *LangChain) Components() []export.Component {
	var langchainCallInputParams []export.NodeInputParam
	if l.pluginLLM.SupportStream() {
		langchainCallInputParams = append(langchainCallInputParams, export.NodeInputParam{
			Name: map[string]string{
				"zh-CN": "流式返回",
			},
			Value: true,
			Key:   "stream",
			Type:  "bool",
		})
	}

	return []export.Component{
		{
			Id:       0,
			Type:     "new_openai",
			Category: "llm",
			Data: export.ComponentData{
				Name: map[string]string{
					"zh-CN": "OpenAI",
				},
				Source: export.ComponentSource{
					CmdType:    "builtin",
					BuiltinCmd: "new_openai",
				},
				InputParams: []export.NodeInputParam{
					{
						Name: map[string]string{"zh-CN": "ApiKey"},
						Key:  "api_key",
						Type: "string",
					},
					{
						Name: map[string]string{"zh-CN": "BaseURL"},
						Key:  "base_url",
						Type: "string",
					},
				},
				OutputAnchors: []export.NodeOutputAnchor{
					{
						Name: map[string]string{"zh-CN": "Default"},
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
			Data: export.ComponentData{
				Name: map[string]string{"zh-CN": "ChatMemory"},
				Source: export.ComponentSource{
					CmdType:    "builtin",
					BuiltinCmd: "chat_memory",
				},
				InputParams: []export.NodeInputParam{
					{
						Name:     map[string]string{"zh-CN": "SessionID"},
						Key:      "session_id",
						Type:     "string",
						Optional: true,
					},
				},
				OutputAnchors: []export.NodeOutputAnchor{
					{
						Name: map[string]string{"zh-CN": "Default"},
						Key:  "default",
						Type: "langchain/chat_memory",
					},
				},
			},
		},
		{
			Type:     "langchain_call",
			Category: "llm",
			Data: export.ComponentData{
				Name:        map[string]string{"zh-CN": "LangChain"},
				Icon:        "",
				Description: map[string]string{},
				Source: export.ComponentSource{
					CmdType:    "builtin",
					BuiltinCmd: "langchain_call",
				},
				InputParams: append([]export.NodeInputParam{
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
						Name:      map[string]string{"zh-CN": "Functions"},
						Key:       "functions",
						Type:      "string",
						Optional:  true,
					},
					{
						InputType: "anchor",
						Name: map[string]string{
							"zh-CN": "Prompt",
						},
						Key:  "prompt",
						Type: "string",
					},
				}, langchainCallInputParams...),
				OutputAnchors: []export.NodeOutputAnchor{
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

func coverMessageToBase(a openai.ChatCompletionMessage) util.Message {
	var fc *util.FunctionCall
	if a.FunctionCall != nil {
		fc = &util.FunctionCall{
			Name:      a.FunctionCall.Name,
			Arguments: a.FunctionCall.Arguments,
		}
	}
	return util.Message{
		Role:         a.Role,
		Content:      a.Content,
		FunctionCall: fc,
		Name:         a.Name,
	}
}

func coverMessageToSDK(a util.Message) openai.ChatCompletionMessage {
	var fc *openai.FunctionCall
	if a.FunctionCall != nil {
		fc = &openai.FunctionCall{
			Name:      a.FunctionCall.Name,
			Arguments: a.FunctionCall.Arguments,
		}
	}
	return openai.ChatCompletionMessage{
		Role:         a.Role,
		Content:      a.Content,
		FunctionCall: fc,
		Name:         a.Name,
	}
}

func coverMessageListToSDK(as []util.Message) []openai.ChatCompletionMessage {
	var bs []openai.ChatCompletionMessage
	for _, a := range as {
		bs = append(bs, coverMessageToSDK(a))
	}
	return bs
}

func (l *LangChain) Cmd() map[string]export.CMDer {
	return map[string]export.CMDer{
		"new_openai":     l.pluginLLM.NewOpenAICmd(),
		"langchain_call": l.pluginLLM.CallOpenAICmd(),
		// chat_memory 存储对话记录
		"chat_memory": util.NewFun(func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error) {
			idi := params["session_id"]
			if idi == nil {
				return map[string]interface{}{"default": util.NewMemoryChatMemory("")}, nil
			}
			id := idi.(string)

			memory := util.NewMemoryChatMemory(id)
			return map[string]interface{}{"default": memory}, nil
		}),
	}
}

func (l *LangChain) GoSymbols() map[string]map[string]reflect.Value {
	return nil
}

var _ export.Plugin = (*LangChain)(nil)

// Register function is used to register the plugin
func Register(r export.Register) {
	r.RegisterPlugin(NewLangChain(sashabaranov.NewPlugin()))
}
