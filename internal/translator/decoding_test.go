package translator

import (
	"maps"
	"os"
	"slices"
	"testing"
)

var (
	testDecoder = &Translator{}
	encodedRESPLongString = []byte{
		'$', '3', '6', '4', '\r', '\n', 
		'{', '"', 's', 't', 'o', 'r', 'e', 's', 'C', 'o', 'u', 'n', 't', 'e', 'd',
		'"', ':', '"', '8', '"', ',', '"', 'f', 'i', 'l', 't', 'e', 'r', 'e', 'd', 'R', 'e', 'g', 'i', 'o', 'n',
		'C', 'o', 'd', 'e', 's', '"', ':', 'n', 'u', 'l', 'l', ',', '"', 's', 't', 'o', 'r', 'e', 's', '"', ':', '[',
		'{', '"', 's', 't', 'o', 'r', 'e', 'I', 'd', '"', ':', '"', '2', '4', '5', '4', '"', ',', '"', 'r', 'e', 'g',
		'i', 'o', 'n', 'C', 'o', 'd', 'e', '"', ':', '"', 'M', 'S', 'K', '"', '}', ',', '{', '"', 's', 't', 'o', 'r',
		'e', 'I', 'd', '"', ':', '"', '3', '8', '7', '5', '"', ',', '"', 'r', 'e', 'g', 'i', 'o', 'n', 'C', 'o', 'd',
		'e', '"', ':', '"', 'R', 'Z', 'N', '"', '}', ',', '{', '"', 's', 't', 'o', 'r', 'e', 'I', 'd', '"', ':', '"',
		'B', '0', '5', '6', '"', ',', '"', 'r', 'e', 'g', 'i', 'o', 'n', 'C', 'o', 'd', 'e', '"', ':', '"', 'M', 'S', 'K',
		'"', '}', ',', '{', '"', 's', 't', 'o', 'r', 'e', 'I', 'd', '"', ':', '"', 'D', '5', '0', '0', '"', ',', '"', 'r',
		'e', 'g', 'i', 'o', 'n', 'C', 'o', 'd', 'e', '"', ':', '"', 'M', 'S', 'K', '"', '}', ',', '{', '"', 's', 't', 'o',
		'r', 'e', 'I', 'd', '"', ':', '"', 'L', '5', '0', '0', '"', ',', '"', 'r', 'e', 'g', 'i', 'o', 'n', 'C', 'o',
		'd', 'e', '"', ':', '"', 'T', 'V', 'R', '"', '}', ',', '{', '"', 's', 't', 'o', 'r', 'e', 'I', 'd', '"', ':', '"',
		'L', '7', '0', '0', '"', ',', '"', 'r', 'e', 'g', 'i', 'o', 'n', 'C', 'o', 'd', 'e', '"', ':', '"', 'T', 'V', 'R',
		'"', '}', ',', '{', '"', 's', 't', 'o', 'r', 'e', 'I', 'd', '"', ':', '"', 'R', '2', '0', '0', '"', ',', '"', 'r',
		'e', 'g', 'i', 'o', 'n', 'C', 'o', 'd', 'e', '"', ':', '"', 'S', 'H', 'L', '"', '}', ',', '{', '"', 's', 't', 'o',
		'r', 'e', 'I', 'd', '"', ':', '"', 'R', '3', '2', '3', '2', '"', ',', '"', 'r', 'e', 'g', 'i', 'o', 'n', 'C',
		'o', 'd', 'e', '"', ':', '"', 'Y', 'A', 'K', '"', '}', ']', '}', '\r', '\n', 
	}
	encodedRESPSimpleMap = []byte{
		'%', '2', '\r', '\n',
		'+', 'f', 'i', 'r', 's', 't', '\r', '\n',
		':', '1', '\r', '\n',
		'+', 's', 'e', 'c', 'o', 'n', 'd', '\r', '\n',
		':', '2', '\r', '\n',
	}
)

func TestDecode(t *testing.T) {
	longString, _ := os.ReadFile("/home/rseleznev/response.json")
	
	testData := []struct{
		name string
		data []byte
		expectedBytesResult []byte
		expectedMapResult map[string]string
		expectedErr error
	}{
		{
			name: "success long string",
			data: encodedRESPLongString,
			expectedBytesResult: longString,
			expectedErr: nil,
		},
		{
			name: "success simple map",
			data: encodedRESPSimpleMap,
			expectedMapResult: map[string]string{
				"first": "1",
				"second": "2",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			r, err := testDecoder.Decode(tt.data)
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
			
			switch res := r.(type) {
			case []byte:
				if slices.Compare(res, tt.expectedBytesResult) != 0 {
					t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedBytesResult, res)
				}

			case map[string]string:
				if !maps.Equal(res, tt.expectedMapResult) {
					t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedMapResult, res)
				}

			}
		})
	}
}

func Test_parseBulkString(t *testing.T) {
	testData := []struct{
		name string
		data []byte
		expectedResult []byte
	}{
		{
			name: "success 1",
			data: []byte{'$', '2', '\r', '\n', 'O', 'K', '\r', '\n',},
			expectedResult: []byte("OK"),
		},
		{
			name: "success 2",
			data: []byte{'$', '5', '\r', '\n', 'O', 'K', ' ', 'K', 'O', '\r', '\n',},
			expectedResult: []byte("OK KO"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testDecoder.setDecodingData(tt.data)

			_, res := testDecoder.parseBulkString(0)
			resBytes, ok := res.([]byte)
			if !ok {
				t.Error("Ошибка приведения типа к []byte")
			}

			if slices.Compare(resBytes, tt.expectedResult) != 0 {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, res)
			}
		})
	}
}

func Test_parseMap(t *testing.T) {
	testData := []struct{
		name string
		data []byte
		expectedResult map[string]string
	}{
		{
			name: "success 1",
			data: []byte{
				'%', '2', '\r', '\n',
				'+', 'f', 'i', 'r', 's', 't', '\r', '\n',
				':', '1', '\r', '\n',
				'+', 's', 'e', 'c', 'o', 'n', 'd', '\r', '\n',
				':', '2', '\r', '\n'},
			expectedResult: map[string]string{
				"first": "1",
				"second": "2",
			},
		},
		{
			name: "success 2",
			data: []byte{
				'%', '3', '\r', '\n',
				'+', 'f', 'i', 'r', 's', 't', '\r', '\n',
				':', '1', '\r', '\n',
				'+', 's', 'e', 'c', 'o', 'n', 'd', '\r', '\n',
				':', '2', '\r', '\n',
				'$', '3', '\r', '\n',
				'G', 'E', 'T', '\r', '\n',
				'$', '2', '\r', '\n',
				'O', 'K', '\r', '\n',},
			expectedResult: map[string]string{
				"first": "1",
				"second": "2",
				"GET": "OK",
			},
		},
		{
			name: "success 3 with arr",
			data: []byte{
				'%', '2', '\r', '\n',
				'$', '5', '\r', '\n',
				'f', 'i', 'r', 's', 't', '\r', '\n',
				':', '1', '\r', '\n',
				'$', '5', '\r', '\n',
				'M', 'o', 'd', 'e', 's', '\r', '\n',
				'*', '0', '\r', '\n',},
			expectedResult: map[string]string{
				"first": "1",
				"Modes": "",
			},
		},
		{
			name: "success 4 zero len map",
			data: []byte{
				'%', '0', '\r', '\n'},
			expectedResult: map[string]string{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testDecoder.setDecodingData(tt.data)

			_, m := testDecoder.parseMap(0)

			res, ok := m.(map[string]string)
			if !ok {
				t.Error("Ошибка приведения типа к map[string]string")
			}

			if !maps.Equal(res, tt.expectedResult) {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, res)
			}
		})
	}
}

func Test_parseInteger(t *testing.T) {
	testData := []struct{
		name string
		data []byte
		expectedBytesResult []byte
		expectedStringResult string
	}{
		{
			name: "success 1dg int",
			data: []byte{':', '7', '\r', '\n'},
			expectedBytesResult: []byte{'7'},
			expectedStringResult: "7",
		},
		{
			name: "success 2dg int",
			data: []byte{':', '4', '2', '\r', '\n'},
			expectedBytesResult: []byte{'4', '2'},
			expectedStringResult: "42",
		},
		{
			name: "success long int",
			data: []byte{':', '1', '8', '4', '4', '3', '\r', '\n'},
			expectedBytesResult: []byte{'1', '8', '4', '4', '3'},
			expectedStringResult: "18443",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testDecoder.setDecodingData(tt.data)

			_, r := testDecoder.parseInteger(0)

			resBytes, ok := r.([]byte)
			if !ok {
				t.Error("Ошибка приведения типа к []byte")
			}

			if slices.Compare(resBytes, tt.expectedBytesResult) != 0 {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedBytesResult, resBytes)
			}

			resString := string(resBytes)
			if resString != tt.expectedStringResult {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedStringResult, resString)
			}
		})
	}
}

func Test_parseArray(t *testing.T) {
	longValue, _ := os.ReadFile("/home/rseleznev/response.json")
	
	testData := []struct{
		name string
		data []byte
		expectedResult []byte
	}{
		{
			name: "success bulk string",
			data: []byte{'*', '2', '\r', '\n', '$', '3', '\r', '\n', 'G', 'E', 'T', '\r', '\n', '$', '1', '\r', '\n', '1', '\r', '\n'},
			expectedResult: []byte("GET1"),
		},
		{
			name: "success simple string",
			data: []byte{'*', '1', '\r', '\n', '+', 'O', 'K', '\r', '\n'},
			expectedResult: []byte("OK"),
		},
		{
			name: "success long string",
			data: encodedArrWithLongString, // из файла encoding_test.go
			expectedResult: longValue,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testDecoder.setDecodingData(tt.data)
			
			_, r := testDecoder.parseArray(0)
			
			resArrAny, ok := r.([]any)
			if !ok {
				t.Error("Ошибка приведения типа к []any")
			}
			var resArr []byte

			for _, v := range resArrAny {
				arrPart, ok := v.([]byte)
				if !ok {
					t.Error("Ошибка приведения типа к []byte")
				}
				resArr = append(resArr, arrPart...)
			}

			if slices.Compare(resArr, tt.expectedResult) != 0 {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, resArr)
			}
		})
	}
}