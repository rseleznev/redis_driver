package message

import "github.com/rseleznev/redis_driver/internal/models"

func Deserialize(input []models.DOMPart) any {
	var result any

	if len(input) == 1 {
		result = input[0].Value
		return result
	}

	switch input[0].PartType {
	case "map":
		m := map[string]string{}
		index := 1

		for range input[0].ContentLen {
			key := string(input[index].Value)
			value := string(input[index+1].Value)
			index += 2

			m[key] = value
		}
		result = m

	}
	
	return result
}