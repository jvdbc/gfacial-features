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
		log.Fatal("Erreur : La variable d'environnement GEMINI_API_KEY n'est pas définie.")
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

	fmt.Println("--- Modèles disponibles ---")
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
		log.Fatalf("Impossible de lire l'image : %v", err)
	}

	// Détermination du format de l'image (jpeg, png, etc.)
	ext := strings.ToLower(filepath.Ext(imagePath))
	format := strings.TrimPrefix(ext, ".")
	if format == "jpg" {
		format = "jpeg"
	}

	fmt.Printf("🔍 Analyse de l'image en cours avec %s...\n", modelName)

	// 5. Création du Prompt Multimodal (Texte + Image)
	promptText := `Act like a physiognomy expert and analyze this photo of a human face.
Provide a structured response including the following points:
- **Apparent Emotion:** What mood is being expressed?
- **Demographic Estimate:** Approximate age and apparent gender.
- **Physical Characteristics:**
    - Eye color
    - Hair color and length
    - Beard/mustache
- **Accessories:**
    - Glasses?
    - Piercings?

Answer concisely.`

	parts := []*genai.Part{
		{Text: promptText},
		{InlineData: &genai.Blob{Data: imgData, MIMEType: "image/jpeg"}},
	}

	// 6. Envoi de la requête et affichage de la réponse
	// On baisse la "temperature" pour avoir une réponse plus factuelle et moins créative
	config := &genai.GenerateContentConfig{Temperature: genai.Ptr[float32](0.4)}
	start := time.Now()
	resp, err := client.Models.GenerateContent(ctx, modelName, []*genai.Content{{Parts: parts}}, config)
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("Erreur pendant l'appel GenerateContent (après %s): %v", elapsed, err)
	}

	fmt.Printf("Temps d'analyse : %s\n", elapsed)

	printResponse(resp)
}

// Fonction utilitaire pour afficher la réponse proprement
func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println("\n--- Résultat de l'analyse ---")
				fmt.Println(part.Text)
				fmt.Println("-----------------------------")
			}
		}
	}
}
