package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	BaseURL      = "https://api.scaleway.ai/fb7f7471-4eb8-49a0-bb6b-d8f2655902fd/v1"
	PixtralModel = "pixtral-12b-2409"
)

type Client struct {
	client openai.Client
}

func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client := openai.NewClient(
		option.WithBaseURL(BaseURL),
		option.WithAPIKey(apiKey),
	)

	return &Client{client: client}, nil
}

func (c *Client) AnalyzeFace(ctx context.Context, imagePath string) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	prompt := `Provide a structured response including the following points:
- **Physical Characteristics:**
    - Eye color
    - Hair color and length
    - Beard/mustache
- **Accessories:**
    - Glasses?
    - Piercings?
- **Apparent Emotion:** What mood is being expressed?
- **Demographic Estimate:** Approximate age and apparent gender.`

	promptPart := openai.TextContentPart(prompt)
	imgPart := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
		URL: fmt.Sprintf("data:image/jpeg;base64,%s", encodedImage),
	})

	unionParts := []openai.ChatCompletionContentPartUnionParam{
		promptPart,
		imgPart,
	}

	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       PixtralModel,
		Temperature: openai.Float(0.1),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(unionParts),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "facial_appearance",
					Schema: FacialAppearanceRespSchema,
				},
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to analyze image: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}
