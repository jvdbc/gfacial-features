package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type PhysicalCharacteristics struct {
	EyeColor   string `json:"eye_color" jsonschema_description:"The color of the person's eyes"`
	HairColor  string `json:"hair_color" jsonschema_description:"The color of the person's hair"`
	HairLength string `json:"hair_length" jsonschema_description:"The length of the person's hair (e.g., short, medium, long)"`
	Beard      string `json:"beard" jsonschema_description:"Whether the person has a beard or mustache, and if so, its style"`
}

type Accessories struct {
	Glasses   bool `json:"glasses" jsonschema_description:"Whether the person is wearing glasses"`
	Piercings bool `json:"piercings" jsonschema_description:"Whether the person has piercings"`
}

type DemographicEstimate struct {
	Age    string `json:"age" jsonschema_description:"An approximate estimate of the person's age (e.g., child, teenager, adult, senior)"`
	Gender string `json:"gender" jsonschema_description:"An approximate estimate of the person's gender"`
}

type FacialAppearance struct {
	Physical_Characteristics PhysicalCharacteristics `json:"physical_characteristics" jsonschema_description:"The physical characteristics of the person"`
	Accessories              Accessories             `json:"accessories" jsonschema_description:"The accessories the person is wearing"`
	ApparentEmotion          string                  `json:"apparent_emotion" jsonschema_description:"The mood being expressed by the person"`
	DemographicEstimate      DemographicEstimate     `json:"demographic_estimate" jsonschema_description:"An approximate estimate of the person's age and gender"`
}

// https://github.com/openai/openai-go/blob/main/examples/structured-outputs/main.go
func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

// Generate the JSON schema at initialization time
var FacialAppearanceRespSchema = GenerateSchema[FacialAppearance]()

func main() {
	// 1. Récupération de la clé API
	apiKey := os.Getenv("SCW_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("Error : env var SCW_SECRET_KEY is not set")
	}

	// 2. Chargement de l'image
	imagePath := "visage.jpg"
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		log.Fatalf("Image unavailable : %v", err)
	}

	// Encoder l'image en base64
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

	// 3. Création du client OpenAI
	client := openai.NewClient(
		option.WithBaseURL("https://api.scaleway.ai/314acaf2-5b9b-4c8d-94bf-67a059237bb2/v1"),
		option.WithAPIKey(apiKey),
	)

	ctx := context.Background()
	start := time.Now()

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       "pixtral-12b-2409",
		Temperature: openai.Float(0.1),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(unionParts),
		},
		// https://opensource.googleblog.com/2026/01/a-json-schema-package-for-go.html

		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "facial_appearance",
					Schema: FacialAppearanceRespSchema,
				},
			},
		},
	})

	elapsed := time.Since(start)

	if err != nil {
		log.Fatalf("Error during client.Chat.Completions (after %s): %v", elapsed, err)
	}

	fmt.Printf("Elapsed time: %.3f s\n", elapsed.Seconds())
	println(resp.Choices[0].Message.Content)
}
