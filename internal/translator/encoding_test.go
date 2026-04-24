package translator

import (
	"os"
	"slices"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

var (
	testEncoder = &Translator{}
	encodedArrWithLongString = []byte{
		'*', '1', '\r', '\n',
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
)

func TestEncode(t *testing.T) {
	longString, _ := os.ReadFile("/home/rseleznev/response.json")
	
	testData := []struct{
		name string
		buf *models.SendBuf
		params []any
		expectedErr error
		expectedResult []byte
	}{
		{
			name: "success strings",
			buf: &models.SendBuf{
				WritePos: 0,
				Buf: make([]byte, 200),
			},
			params: []any{"GET", "1"},
			expectedErr: nil,
			expectedResult: []byte{'*', '2', '\r', '\n', '$', '3', '\r', '\n', 'G', 'E', 'T', '\r', '\n', '$', '1', '\r', '\n', '1', '\r', '\n'},
		},
		{
			name: "success bytes",
			buf: &models.SendBuf{
				WritePos: 0,
				Buf: make([]byte, 200),
			},
			params: []any{[]byte{'S', 'E', 'T'}},
			expectedErr: nil,
			expectedResult: []byte{'*', '1', '\r', '\n', '$', '3', '\r', '\n', 'S', 'E', 'T', '\r', '\n'},
		},
		{
			name: "success long value",
			buf: &models.SendBuf{
				WritePos: 0,
				Buf: make([]byte, 2000),
			},
			params: []any{longString},
			expectedErr: nil,
			expectedResult: encodedArrWithLongString,
		},
		{
			name: "fail ErrSendBufTooShort",
			buf: &models.SendBuf{
				WritePos: 0,
				Buf: make([]byte, 2),
			},
			params: []any{},
			expectedErr: models.ErrSendBufTooShort,
			expectedResult: []byte{},
		},
		{
			name: "fail ErrUnsupportedDataType",
			buf: &models.SendBuf{
				WritePos: 0,
				Buf: make([]byte, 200),
			},
			params: []any{map[string]string{
				"name": "test",
			}},
			expectedErr: models.ErrUnsupportedDataType,
			expectedResult: []byte{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := testEncoder.Encode(tt.buf, tt.params)
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
			if slices.Compare(tt.buf.Buf[:tt.buf.WritePos], tt.expectedResult) != 0 {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, tt.buf.Buf[:tt.buf.WritePos])
			}
		})
	}
}