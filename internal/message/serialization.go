package message

import (
	"strconv"

	"github.com/rseleznev/redis_driver/internal/models"
)

func SerializeGetCommand(key string) []byte {
	arrLen := 3 + 16 + len(key)
	keyLen := len(key)

	if keyLen > 9 { // если длина ключа - двузначное число (пока как максимум)
		arrLen++
	}

	result := make([]byte, 0, arrLen)

	fixBytes := []byte{
		'*', '2', '\r', '\n',
		'$', '3', '\r', '\n',
		'G', 'E', 'T', '\r', '\n',
		'$', '3', '2', '\r', '\n',
		'd', '4', '1', 'd', '8', 'c', 'd', '9', '8', 'f', '0', '0', 'b', '2', '0', '4', 'e', '9', '8', '0', '0', '9', '9', '8', 'e', 'c', 'f', '8', '4', '2', '7', 'e', '\r', '\n',
		// d41d8cd98f00b204e9800998ecf8427e
	}
	result = append(result, fixBytes...)

	// if keyLen < 9 {
	// 	result = append(result, byte(keyLen))
	// } else {
	// 	fstDg := keyLen / 10
	// 	sndDg := keyLen % 10
	// 	result = append(result, byte(fstDg), byte(sndDg))
	// }
	// result = append(result, '\r', '\n')

	// for range keyLen {
	// 	result = append(result, key)
	// }

	// result = append(result, '\r', '\n')

	return result
}

func SerializeSetCommand(command string, key string, value any, dur int) []byte {
	// Добавляем команду
	fixBytes := []byte{
		'*', '5', '\r', '\n',
		'$', '3', '\r', '\n',
		'S', 'E', 'T', '\r', '\n',
		// key '\r', '\n',
		// value '\r', '\n',
		// 'E', 'X', '\r', '\n',
		// dur '\r', '\n',
	}

	// Добавляем ключ
	k := buildDOM(key)
	kBytes := transformDOMToRESP(k)

	// Добавляем значение
	v := buildDOM(value)
	vBytes := transformDOMToRESP(v)

	// Добавляем длительность
	exBytes := []byte{'$', '2', '\r', '\n', 'E', 'X', '\r', '\n'}
	dr := buildDOM(dur)
	drBytes := transformDOMToRESP(dr)

	totalLen := len(fixBytes) + len(kBytes) + len(vBytes) + len(exBytes) + len(drBytes)
	result := make([]byte, 0, totalLen)

	result = append(result, fixBytes...)
	result = append(result, kBytes...)
	result = append(result, vBytes...)
	result = append(result, exBytes...)
	result = append(result, drBytes...)
	
	return result
}

func buildDOM(input any) []models.DOMPart {
	var result []models.DOMPart

	switch input := input.(type) {
	case string:		
		part := models.DOMPart{
			PartType: "string",
			ValueLen: len(input),
			Value: []byte(input),
		}

		result = append(result, part)

	case int:
		vl := 1

		if input > 9 && input < 99 {
			vl = 2
		}
		v := strconv.Itoa(input)
		
		part := models.DOMPart{
			PartType: "int",
			ValueLen: vl,
			Value: []byte(v),
		}

		result = append(result, part)

	case []byte:

	case map[string]string:

	}

	return result
}

func transformDOMToRESP(input []models.DOMPart) []byte {
	var result []byte
	
	for _, v := range input {
		b := transformDOMPartToRESP(v)
		result = append(result, b...)
	}

	return result
}

func transformDOMPartToRESP(input models.DOMPart) []byte {
	var result []byte

	switch input.PartType {
	case "string":
		result = append(result, '$')

		vl := input.ValueLen
		if vl > 99 {
			panic("слишком длинная строка!") // заглушка
		}
		if vl > 9 {
			fstDg := vl / 10
			scndDg := vl % 10
			fstDgS := strconv.Itoa(fstDg)
			scndDgS := strconv.Itoa(scndDg)
			result = append(result, fstDgS[0], scndDgS[0], '\r', '\n')
		} else {
			vls := strconv.Itoa(vl)
			result = append(result, vls[0], '\r', '\n')
		}
		result = append(result, input.Value...)
		result = append(result, '\r', '\n')

	case "int":
		result = append(result, ':')
		result = append(result, input.Value...)
		result = append(result, '\r', '\n')

	}

	return result
}