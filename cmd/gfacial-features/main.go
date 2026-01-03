package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"google.golang.org/genai"
)

func main() {
	// 1. Récupération de la clé API
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error : env var GEMINI_API_KEY is not set")
	}

	// 2. Initialisation du client Gemini
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- Available Models ---")
	models, err := client.Models.List(ctx, &genai.ListModelsConfig{})
	if err != nil {
		log.Fatal(err)
	}

	for _, model := range models.Items {
		if slices.Contains(model.SupportedActions, "generateContent") {
			fmt.Println(model.Name)
		}
	}

	// 3. Configuration du modèle
	// Gemini 2.5 Flash est rapide et excellent pour la vision
	modelName := "gemini-2.5-flash"

	// 4. Chargement de l'image
	imagePath := "visage.jpg" // Assurez-vous que ce fichier existe
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		log.Fatalf("Image unavailable : %v", err)
	}

	// Détermination du format de l'image (jpeg, png, etc.)
	ext := strings.ToLower(filepath.Ext(imagePath))
	format := strings.TrimPrefix(ext, ".")
	if format == "jpg" {
		format = "jpeg"
	}

	fmt.Printf("🔍 Image analysis with %s...\n", modelName)

	// 5. Création du Prompt Multimodal (Texte + Image)
	promptText := `Provide a structured response including the following points:
- **Apparent Emotion:** What mood is being expressed?
- **Demographic Estimate:** Approximate age and apparent gender.
- **Physical Characteristics:**
    - Eye color
    - Hair color and length
    - Beard/mustache
- **Accessories:**
    - Glasses?
    - Piercings?`

	parts := []*genai.Part{
		{Text: promptText},
		{InlineData: &genai.Blob{Data: imgData, MIMEType: "image/jpeg"}},
	}

	// 6. Envoi de la requête et affichage de la réponse
	// On ajoute une instruction système pour cadrer la réponse
	// On baisse la "temperature" pour avoir une réponse plus factuelle et moins créative
	// On demande une réponse en JSON structuré pour faciliter le parsing
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: "Act like a physiognomy expert analyzing human faces from images. Answer concisely."}}},
		Temperature:       genai.Ptr[float32](0.4),
		ResponseMIMEType:  "application/json",
	}
	start := time.Now()
	resp, err := client.Models.GenerateContent(ctx, modelName, []*genai.Content{{Parts: parts}}, config)
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("Error during GenerateContent (after %s): %v", elapsed, err)
	}

	fmt.Printf("Analyse time : %s\n", elapsed)

	printResponse(resp)
}

// Fonction utilitaire pour afficher la réponse proprement
func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println("\n--- Analyse result ---")
				fmt.Println(part.Text)
				fmt.Println("-----------------------------")
			}
		}
	}
}
