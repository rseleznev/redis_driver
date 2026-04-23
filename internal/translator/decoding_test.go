package translator

import (
	"slices"
	"testing"
)

var (
	testDecoder = &Translator{}
)

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