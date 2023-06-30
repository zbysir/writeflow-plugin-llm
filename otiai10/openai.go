package otiai10

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/otiai10/openaigo"
	"github.com/spf13/cast"
	"github.com/zbysir/writeflow/pkg/export"
	"github.com/zbysir/writeflow_plugin_llm/util"
)

// Plugin implement PluginLLM
// yaegi is aslo can't support otiai10/openaigo, so just write it but not use it.
type Plugin struct {
}

func (p *Plugin) NewOpenAICmd() export.CMDer {
	return util.NewFun(func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error) {
		key := params["api_key"].(string)
		baseUrl := cast.ToString(params["base_url"])
		client := openaigo.NewClient(key)
		if baseUrl != "" {
			client.BaseURL = baseUrl
		}
		return map[string]interface{}{"default": client}, nil
	})
}

func (p *Plugin) CallOpenAICmd() export.CMDer {
	return util.NewFun(func(ctx context.Context, params map[string]interface{}) (rsp map[string]interface{}, err error) {
		//log.Infof("langchain_call")
		openaiClient := params["llm"].(*openaigo.Client)
		promptI := params["prompt"]
		functionI := params["functions"]
		if promptI == nil {
			return nil, fmt.Errorf("prompt is nil")
		}
		enableSteam := cast.ToBool(params["stream"])
		prompt := promptI.(string)
		var functions json.Marshaler = nil
		if functionI != nil {
			function := functionI.(string)
			functions = json.RawMessage(function)
		}

		var messages []openaigo.Message
		var chatMemory util.ChatMemory
		if params["chat_memory"] != nil {
			chatMemory = params["chat_memory"].(util.ChatMemory)
		}

		if chatMemory != nil {
			messages = append(messages, coverMessageListToSDK(chatMemory.GetHistory(ctx))...)
		}

		userMsg := openaigo.Message{Content: prompt, Role: "user"}
		if chatMemory != nil {
			chatMemory.AppendHistory(ctx, coverMessageToBase(userMsg))
		}
		messages = append(messages, userMsg)

		if enableSteam {
			steam := util.NewSteamResponse()
			go func() {
				content := ""
				_, _ = openaiClient.ChatCompletion(ctx, openaigo.ChatCompletionRequestBody{
					Model:            "gpt-3.5-turbo-0613",
					Messages:         messages,
					MaxTokens:        2000,
					Temperature:      0,
					TopP:             0,
					N:                0,
					Stop:             nil,
					PresencePenalty:  0,
					FrequencyPenalty: 0,
					LogitBias:        nil,
					User:             "",
					Functions:        functions,
					FunctionCall:     "",
					StreamCallback: func(res openaigo.ChatCompletionResponse, done bool, err error) {
						if err != nil {
							steam.Close(err)
							return
						}
						if done {
							steam.Close(nil)

							if chatMemory != nil {
								if content != "" {
									chatMemory.AppendHistory(ctx, util.Message{
										Role:    "assistant",
										Content: content,
									})
								}
							}
							return
						}
						if len(res.Choices) > 0 && res.Choices[0].Delta.Content != "" {
							steam.Append(res.Choices[0].Delta.Content)
							content += res.Choices[0].Delta.Content
						}
					},
				})
			}()

			return map[string]interface{}{"default": steam, "function_call": ""}, nil
		} else {
			res, err := openaiClient.ChatCompletion(ctx, openaigo.ChatCompletionRequestBody{
				Model:            "gpt-3.5-turbo-0613",
				Messages:         messages,
				MaxTokens:        2000,
				Temperature:      0,
				TopP:             0,
				N:                0,
				Stop:             nil,
				PresencePenalty:  0,
				FrequencyPenalty: 0,
				LogitBias:        nil,
				User:             "",
				Functions:        functions,
				FunctionCall:     "",
			})
			if err != nil {
				return map[string]interface{}{}, err
			}

			if chatMemory != nil {
				chatMemory.AppendHistory(ctx, coverMessageToBase(res.Choices[0].Message))
			}

			return map[string]interface{}{"default": res.Choices[0].Message.Content, "function_call": res.Choices[0].Message.FunctionCall}, nil
		}
	})
}

func (p *Plugin) SupportStream() bool {
	return true
}

func coverMessageToBase(a openaigo.Message) util.Message {
	var fc *util.FunctionCall
	if a.FunctionCall != nil {
		fc = &util.FunctionCall{
			Name:      a.FunctionCall.Name,
			Arguments: a.FunctionCall.ArgumentsRaw,
		}
	}
	return util.Message{
		Role:         a.Role,
		Content:      a.Content,
		FunctionCall: fc,
		Name:         a.Name,
	}
}

func coverMessageToSDK(a util.Message) openaigo.Message {
	var fc *openaigo.FunctionCall
	if a.FunctionCall != nil {
		fc = &openaigo.FunctionCall{
			Name:         a.FunctionCall.Name,
			ArgumentsRaw: a.FunctionCall.Arguments,
		}
	}
	return openaigo.Message{
		Role:         a.Role,
		Content:      a.Content,
		FunctionCall: fc,
		Name:         a.Name,
	}
}

func coverMessageListToSDK(as []util.Message) []openaigo.Message {
	var bs []openaigo.Message
	for _, a := range as {
		bs = append(bs, coverMessageToSDK(a))
	}
	return bs
}
