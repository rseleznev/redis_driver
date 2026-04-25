package connection

import (
	"sync"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

type mockSocket struct {
	getSocketFdFunc func() int
	connectFunc func(*models.Options) error
	closeFunc func()
}

func (ms mockSocket) GetSocketFd() int {
	return ms.getSocketFdFunc()
}

func (ms mockSocket) Connect(opts *models.Options) error {
	return ms.connectFunc(opts)
}

func (ms mockSocket) Close() {
	ms.closeFunc()
}

var testConnection = &Connection{
	mu: sync.Mutex{},
}

func Test_connect(t *testing.T) {
	testData := []struct{
		name string
		opts *models.Options
		expectedErr error
		mockSock mockSocket
	}{
		{
			name: "success",
			opts: &models.Options{
				RetryAmount: 3,
			},
			expectedErr: nil,
			mockSock: mockSocket{
				connectFunc: func(o *models.Options) error {
					return nil
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testConnection.opts = tt.opts
			testConnection.socket = tt.mockSock

			err := testConnection.connect()
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
		})
	}
}