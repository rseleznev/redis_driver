package translator

import (
	"slices"
	"testing"
)

var (
	testDecoder = &Translator{}
)

func Test_parseArray(t *testing.T) {
	testData := []struct{
		name string
		data []byte
		expectedErr error
		expectedResult []byte
	}{
		{
			name: "success",
			data: []byte{'*', '2', '\r', '\n', '$', '3', '\r', '\n', 'G', 'E', 'T', '\r', '\n', '$', '1', '\r', '\n', '1', '\r', '\n'},
			expectedErr: nil,
			expectedResult: []byte("GET1"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// закинуть в decodingData
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