package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

// Constants
const (
	uploadDir     = "./uploads"
	maxUploadSize = 5 * 1024 * 1024 // 10MB
)

// analyzeFaceImage analyzes a face image using Scaleway API
func analyzeFaceImage(ctx context.Context, client *openai.Client, imagePath string) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
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

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       "pixtral-12b-2409",
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

// handleUploadFace handles POST requests to /upload-face
func handleUploadFace(client *openai.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// add basic CORS headers so that file:// pages can talk to us
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			// preflight; just return
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the form data
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
			return
		}

		// Retrieve the file from the form
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File not found in form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate file is JPG
		if handler.Header.Get("Content-Type") != "image/jpeg" {
			http.Error(w, "Only JPEG files are allowed", http.StatusBadRequest)
			return
		}

		// Create upload directory if it doesn't exist
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create upload directory: %v", err), http.StatusInternalServerError)
			return
		}

		// Save file to disk
		filename := filepath.Join(uploadDir, handler.Filename)
		dst, err := os.Create(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
			return
		}

		// Analyze the face image
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		start := time.Now()
		result, err := analyzeFaceImage(ctx, client, filename)
		elapsed := time.Since(start)

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to analyze image: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"filename":"%s","analysis_time":"%.3fs","result":%s}`, handler.Filename, elapsed.Seconds(), result)
	}
}

func main() {
	// 1. Récupération de la clé API
	apiKey := os.Getenv("SCW_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("Error : env var SCW_SECRET_KEY is not set")
	}

	// 2. Création du client OpenAI
	client := openai.NewClient(
		option.WithBaseURL("https://api.scaleway.ai/314acaf2-5b9b-4c8d-94bf-67a059237bb2/v1"),
		option.WithAPIKey(apiKey),
	)

	// 3. Configure HTTP handlers
	http.HandleFunc("/upload-face", handleUploadFace(&client))

	// Start server
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
