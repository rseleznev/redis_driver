package message

import (
	"errors"
	"strings"

	"github.com/rseleznev/redis_driver/internal/models"
)

// Deserialize десериализует DOM-объект в тип данных Go
func Deserialize(domObj models.DOMPart) any {
	var result any

	switch domObj.PartType {
	case "string":
		return domObj.Value

	case "error": // это надо переделать покрасивее
		s := string(domObj.Value)
		strParts := strings.Fields(s)
		if strParts[1] == "Protocol" && strParts[2] == "error:" {
			errString := strParts[3] + strParts[4] + strParts[5] + strParts[6]
			err := errors.New(errString)
			return errors.Join(models.ErrRedisProtocol, err)
		}
		if strParts[1] == "Unknown" && strParts[2] == "value" && strParts[3] == "type:" {
			strParts = strParts[4:]
			var errString string
			for _, v := range strParts {
				errString = errString + v + " "
			}
			err := errors.New(errString)

			return errors.Join(models.ErrUnknownValueType, err)
		}

		return errors.New(s)

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

	case "null":
		return models.ErrNoValue

	}

	return result
}