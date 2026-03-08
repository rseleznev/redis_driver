package message

import "github.com/rseleznev/redis_driver/internal/models"

func Deserialize(input []models.ParsedResponsePart) any {
	var result any

	for i, v := range input {
		if i == 0 {
			if v.PartType == "string" {
				result = string(v.Value)
			}
		}
	}
	
	return result
}