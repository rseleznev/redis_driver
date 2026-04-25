package connection

import (
	"fmt"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/rseleznev/redis_driver/internal/models"
)

type mockPoller struct {
	addFunc func(models.PollingUnit) error
	getErrorFunc func() error
	deleteSocketFunc func(int)
}

func (mp mockPoller) Add(unit models.PollingUnit) error {
	return mp.addFunc(unit)
}

func (mp mockPoller) GetError() error {
	return mp.getErrorFunc()
}

func (mp mockPoller) DeleteSocketFromPolling(n int) {
	mp.deleteSocketFunc(n)
}


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
		mockPoll mockPoller
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
		{
			name: "success with poll",
			opts: &models.Options{
				RetryAmount: 3,
				PollingTimeout: time.Millisecond*500,
			},
			expectedErr: nil,
			mockPoll: mockPoller{
				addFunc: func(pu models.PollingUnit) error {
					go func ()  {
						pu.ResultChan <- nil
					}()

					return nil
				},
			},
			mockSock: mockSocket{
				getSocketFdFunc: func() int {
					return 2
				},
				connectFunc: func(o *models.Options) error {
					return fmt.Errorf("test err: %w", syscall.EAGAIN)
				},
			},
		},
		{
			name: "fail ErrConnectionRetriesFailed",
			opts: &models.Options{
				RetryAmount: 3,
				PollingTimeout: time.Millisecond*500,
			},
			expectedErr: models.ErrConnectionRetriesFailed,
			mockPoll: mockPoller{
				addFunc: func(pu models.PollingUnit) error {
					return models.ErrSocketAlreadyAdded
				},
			},
			mockSock: mockSocket{
				getSocketFdFunc: func() int {
					return 2
				},
				connectFunc: func(o *models.Options) error {
					return fmt.Errorf("test err: %w", syscall.EAGAIN)
				},
			},
		},
		{
			name: "fail ErrSocketNoAccess",
			opts: &models.Options{
				RetryAmount: 3,
			},
			expectedErr: models.ErrSocketNoAccess,
			mockSock: mockSocket{
				connectFunc: func(o *models.Options) error {
					return models.ErrSocketNoAccess
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testConnection.opts = tt.opts
			testConnection.poller = tt.mockPoll
			testConnection.socket = tt.mockSock

			err := testConnection.connect()
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
		})
	}
}