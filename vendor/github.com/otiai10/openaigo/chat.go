package openaigo

import "encoding/json"

// ChatCompletionRequestBody:
// https://platform.openai.com/docs/guides/chat/chat-completions-beta
// https://platform.openai.com/docs/api-reference/chat
type ChatCompletionRequestBody struct {

	// Model: ID of the model to use.
	// Currently, only gpt-3.5-turbo and gpt-3.5-turbo-0301 are supported.
	Model string `json:"model"`

	// Messages: The messages to generate chat completions for, in the chat format.
	// https://platform.openai.com/docs/guides/chat/introduction
	// Including the conversation history helps when user instructions refer to prior messages.
	// In the example above, the user’s final question of “Where was it played?” only makes sense in the context of the prior messages about the World Series of 2020.
	// Because the models have no memory of past requests, all relevant information must be supplied via the conversation.
	// If a conversation cannot fit within the model’s token limit, it will need to be shortened in some way.
	Messages []Message `json:"messages"`

	// Temperature: What sampling temperature to use, between 0 and 2.
	// Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic.
	// We generally recommend altering this or top_p but not both.
	// Defaults to 1.
	Temperature float32 `json:"temperature,omitempty"`

	// TopP: An alternative to sampling with temperature, called nucleus sampling,
	// where the model considers the results of the tokens with top_p probability mass.
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	// We generally recommend altering this or temperature but not both.
	// Defaults to 1.
	TopP float32 `json:"top_p,omitempty"`

	// N: How many chat completion choices to generate for each input message.
	// Defaults to 1.
	N int `json:"n,omitempty"`

	// Stream: If set, partial message deltas will be sent, like in ChatGPT.
	// Tokens will be sent as data-only server-sent events as they become available,
	// with the stream terminated by a data: [DONE] message.
	Stream bool `json:"stream,omitempty"`

	// StreamCallback is a callback funciton to handle stream response.
	// If provided, this library automatically set `Stream` `true`.
	// This field is added by github.com/otiai10/openaigo only to handle Stream.
	// Thus, it is omitted when the client excute HTTP request.
	StreamCallback func(res ChatCompletionResponse, done bool, err error) `json:"-"`

	// Stop: Up to 4 sequences where the API will stop generating further tokens.
	// Defaults to null.
	Stop []string `json:"stop,omitempty"`

	// MaxTokens: The maximum number of tokens allowed for the generated answer.
	// By default, the number of tokens the model can return will be (4096 - prompt tokens).
	MaxTokens int `json:"max_tokens,omitempty"`

	// PresencePenalty: Number between -2.0 and 2.0.
	// Positive values penalize new tokens based on whether they appear in the text so far,
	// increasing the model's likelihood to talk about new topics.
	// See more information about frequency and presence penalties.
	// https://platform.openai.com/docs/api-reference/parameter-details
	PresencePenalty float32 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty: Number between -2.0 and 2.0.
	// Positive values penalize new tokens based on their existing frequency in the text so far,
	// decreasing the model's likelihood to repeat the same line verbatim.
	// See more information about frequency and presence penalties.
	// https://platform.openai.com/docs/api-reference/parameter-details
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// LogitBias: Modify the likelihood of specified tokens appearing in the completion.
	// Accepts a json object that maps tokens (specified by their token ID in the tokenizer)
	// to an associated bias value from -100 to 100.
	// Mathematically, the bias is added to the logits generated by the model prior to sampling.
	// The exact effect will vary per model, but values between -1 and 1 should decrease or increase likelihood of selection;
	// values like -100 or 100 should result in a ban or exclusive selection of the relevant token.
	LogitBias map[string]int `json:"logit_bias,omitempty"`

	// User: A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse. Learn more.
	// https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids
	User string `json:"user,omitempty"`

	// Functions: A list of functions which GPT is allowed to request to call.
	// Functions []Function `json:"functions,omitempty"`
	Functions json.Marshaler `json:"functions,omitempty"`

	// FunctionCall: You ain't need it. Default is "auto".
	FunctionCall string `json:"function_call,omitempty"`
}

type Functions []Function

func (funcs Functions) MarshalJSON() ([]byte, error) {
	if len(funcs) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]Function(funcs))
}

type Function struct {
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Parameters  Parameters `json:"parameters,omitempty"`
}

type Parameters struct {
	Type       string                    `json:"type,omitempty"` // Must be "object"
	Properties map[string]map[string]any `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// ChatRequest is just an alias of ChatCompletionRequestBody.
type ChatRequest ChatCompletionRequestBody

// Message: An element of messages parameter.
// The main input is the messages parameter. Messages must be an array of message objects,
// where each object has a role (either “system”, “user”, or “assistant”)
// and content (the content of the message).
// Conversations can be as short as 1 message or fill many pages.
// See https://platform.openai.com/docs/api-reference/chat/create#chat/create-messages
type Message struct {

	// Role: Either of "system", "user", "assistant".
	// Typically, a conversation is formatted with a system message first, followed by alternating user and assistant messages.
	// The system message helps set the behavior of the assistant. In the example above, the assistant was instructed with “You are a helpful assistant.”
	// The user messages help instruct the assistant. They can be generated by the end users of an application, or set by a developer as an instruction.
	// The assistant messages help store prior responses. They can also be written by a developer to help give examples of desired behavior.
	Role string `json:"role"`

	// Content: A content of the message.
	Content string `json:"content"`

	// FunctionCall requested by ChatGPT.
	// Only appears in a response from ChatGPT in which ChatGPT wants to call a function.
	FunctionCall *FunctionCall `json:"function_call,omitempty"`

	// Name of the function called, to tell this message is a result of function_call.
	// Only appears in a request from us when the previous message is "function_call" requested by ChatGPT.
	Name string `json:"name,omitempty"`
}

type FunctionCall struct {
	Name         string `json:"name,omitempty"`
	ArgumentsRaw string `json:"arguments,omitempty"`
	// Arguments map[string]any `json:"arguments,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Delta        Message `json:"delta"` // Only appears in stream response
}