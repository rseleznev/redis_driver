package translator

import (
	"slices"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

var testEncoder = Translator{}

func TestEncode(t *testing.T) {
	testData := []struct{
		name string
		buf []byte
		params []any
		expectedErr error
		expectedResult []byte
	}{
		{
			name: "success",
			buf: make([]byte, 0, 10),
			params: []any{"GET"},
			expectedErr: nil,
			expectedResult: []byte{'*', '1', '\r', '\n', '$', '3', '\r', '\n', 'G', 'E', 'T', '\r', '\n'},
		},
		{
			name: "fail build",
			buf: make([]byte, 0, 10),
			params: []any{5},
			expectedErr: models.ErrUnsupportedDataType,
			expectedResult: []byte{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
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