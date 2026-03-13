package message

import "github.com/rseleznev/redis_driver/internal/models"

// Deserialize десериализует DOM-объект в тип данных Go
func Deserialize(domObj models.DOMPart) any {
	var result any

	switch domObj.PartType {
	case "string":
		return domObj.Value

	case "map":
		m := map[string]string{}
		var key, value string

		for _, v := range domObj.Content {
			if key == "" {
				key = string(v.Value)
				continue
			}
			value = string(v.Value)
			m[key] = value

			key = ""
			value = ""
		}

		return m
	}

	return result
}