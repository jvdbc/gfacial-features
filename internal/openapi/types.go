package openapi

import (
	"github.com/invopop/jsonschema"
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

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var FacialAppearanceRespSchema = GenerateSchema[FacialAppearance]()
