package volcengine

import (
	"cmp"
	"context"
	_ "embed"
	"errors"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"

	"github.com/lftk/anki-vocab/internal/dict"
)

//go:embed prompt.txt
var prompt string

type Config struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
	Prompt string `yaml:"prompt"`
}

func New(cfg *Config) *dict.Dict {
	client := arkruntime.NewClientWithApiKey(cfg.APIKey)
	d := &Dict{
		client: client,
		model:  cfg.Model,
		prompt: cmp.Or(cfg.Prompt, prompt),
	}
	caps := &dict.Capabilities{
		Query: &dict.QueryCapabilities{
			AI: true,
		},
	}
	return &dict.Dict{Queryer: d, Capabilities: caps}
}

type Dict struct {
	client *arkruntime.Client
	model  string
	prompt string
}

func (d *Dict) Query(ctx context.Context, word string) ([]byte, error) {
	req := model.CreateChatCompletionRequest{
		Model: d.model,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: &d.prompt,
				},
			},
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: &word,
				},
			},
		},
		ResponseFormat: &model.ResponseFormat{
			Type: model.ResponseFormatJsonObject,
		},
	}
	resp, err := d.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) < 1 {
		return nil, errors.New("no choices in response")
	}

	val := resp.Choices[0].Message.Content.StringValue
	if val == nil {
		return nil, errors.New("empty content in response")
	}

	return []byte(*val), nil
}
