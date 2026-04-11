package translator

import (
	"slices"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

var (
	testEncoder = Translator{}
)

type mockSerializator struct {
	serializeDOMToRESPFunc func([]byte, models.DOMPart) []byte
}

func (ms mockSerializator) serializeDOMToRESP(buf []byte, input models.DOMPart) []byte {
	return ms.serializeDOMToRESPFunc(buf, input)
}

func TestEncode(t *testing.T) {
	testData := []struct{
		name string
		buf []byte
		params []any
		expectedErr error
		expectedResult []byte
		s mockSerializator
	}{
		{
			name: "success",
			buf: make([]byte, 0, 10),
			params: []any{"GET"},
			expectedErr: nil,
			expectedResult: []byte{'G', 'E', 'T'},
			s: mockSerializator{
				serializeDOMToRESPFunc: func(b []byte, d models.DOMPart) []byte {
					b = append(b, 'G', 'E', 'T')

					return b
				},
			},
		},
		{
			name: "fail build",
			buf: make([]byte, 0, 10),
			params: []any{5},
			expectedErr: models.ErrUnsupportedDataType,
			expectedResult: []byte{},
			s: mockSerializator{
				serializeDOMToRESPFunc: func(b []byte, d models.DOMPart) []byte {
					return b
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testEncoder.serializator = tt.s

			res, err := testEncoder.Encode(tt.buf, tt.params)
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
			if slices.Compare(res, tt.expectedResult) != 0 {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, res)
			}
		})
	}
}