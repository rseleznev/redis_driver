package connection

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/rseleznev/redis_driver/internal/models"
)

type mockPoller struct {
	done chan struct{}
	
	addFunc func(models.PollingUnit, <-chan struct{}) error
	getErrorFunc func() error
	deleteSocketFunc func(int)
}
func (mp *mockPoller) Add(unit models.PollingUnit) error {
	return mp.addFunc(unit, mp.done)
}
func (mp *mockPoller) GetError() error {
	return mp.getErrorFunc()
}
func (mp *mockPoller) DeleteSocketFromPolling(n int) {
	mp.closeTestChan()
	mp.deleteSocketFunc(n)
}
func (mp *mockPoller) closeTestChan() {
	close(mp.done)
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


type mockCoder struct {
	encodeFunc func(*models.SendBuf, []any) error
	decodeFunc func([]byte) (any, error)
}
func (mc mockCoder) Encode(buf *models.SendBuf, params []any) error {
	return mc.encodeFunc(buf, params)
}
func (mc mockCoder) Decode(d []byte) (any, error) {
	return mc.decodeFunc(d)
}


type mockMessenger struct {
	sendFunc func([]byte) (int, error)
	receiveFunc func(*models.RecvBuf) error
	changeSocketFunc func(int)
}
func (mm mockMessenger) Send(d []byte) (int, error) {
	return mm.sendFunc(d)
}
func (mm mockMessenger) Receive(buf *models.RecvBuf) error {
	return mm.receiveFunc(buf)
}
func (mm mockMessenger) ChangeSocketFd(n int) {
	mm.changeSocketFunc(n)
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
				addFunc: func(pu models.PollingUnit, done <-chan struct{}) error {
					go func ()  {
						select {
						case pu.ResultChan <- nil:

						case <-done:

						}
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
				addFunc: func(pu models.PollingUnit, done <-chan struct{}) error {
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
			testConnection.poller = &tt.mockPoll
			testConnection.socket = tt.mockSock

			err := testConnection.connect()
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
		})
	}
}

func Test_poll(t *testing.T) {
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
				PollingTimeout: time.Millisecond*500,
			},
			expectedErr: nil,
			mockPoll: mockPoller{
				addFunc: func(pu models.PollingUnit, done <-chan struct{}) error {
					go func ()  {
						select {
						case pu.ResultChan <- nil:

						case <-done:

						}
					}()

					return nil
				},
			},
			mockSock: mockSocket{
				getSocketFdFunc: func() int {
					return 2
				},
			},
		},
		{
			name: "fail sync timeout",
			opts: &models.Options{
				PollingTimeout: time.Millisecond*500,
			},
			expectedErr: models.ErrPollTimeout,
			mockPoll: mockPoller{
				addFunc: func(pu models.PollingUnit, done <-chan struct{}) error {
					return models.ErrSocketAlreadyAdded
				},
			},
			mockSock: mockSocket{
				getSocketFdFunc: func() int {
					return 2
				},
			},
		},
		{
			name: "fail async timeout",
			opts: &models.Options{
				PollingTimeout: time.Millisecond*500,
			},
			expectedErr: models.ErrPollTimeout,
			mockPoll: mockPoller{
				done: make(chan struct{}),
				addFunc: func(pu models.PollingUnit, done <-chan struct{}) error {
					go func() {
						time.Sleep(time.Second*1)
						
						select {
						case pu.ResultChan <- nil:

						case <-done:

						}
						
					}()
					
					return nil
				},
				deleteSocketFunc: func(_ int) {},
			},
			mockSock: mockSocket{
				getSocketFdFunc: func() int {
					return 2
				},
			},
		},
		{
			name: "fail sync ErrConnectionClosed",
			opts: &models.Options{
				PollingTimeout: time.Millisecond*500,
			},
			expectedErr: models.ErrConnectionClosed,
			mockPoll: mockPoller{
				addFunc: func(pu models.PollingUnit, done <-chan struct{}) error {
					return fmt.Errorf("test err: %w", syscall.EPIPE)
				},
			},
			mockSock: mockSocket{
				getSocketFdFunc: func() int {
					return 2
				},
			},
		},
		{
			name: "fail async ErrConnectionClosed",
			opts: &models.Options{
				PollingTimeout: time.Millisecond*500,
			},
			expectedErr: models.ErrConnectionClosed,
			mockPoll: mockPoller{
				addFunc: func(pu models.PollingUnit, done <-chan struct{}) error {
					go func() {
						select {
						case pu.ResultChan <- fmt.Errorf("test err: %w", models.ErrSocketHUPEvent):

						case <-done:

						}
					}()
					
					return nil
				},
			},
			mockSock: mockSocket{
				getSocketFdFunc: func() int {
					return 2
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testConnection.opts = tt.opts
			testConnection.poller = &tt.mockPoll
			testConnection.socket = tt.mockSock

			err := testConnection.poll("connect")
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	testData := []struct{
		name string
		opts *models.Options
		expectedErr error
		setUpFunc func()
		cleanUpFunc func()
		mockPoll mockPoller
		mockSock mockSocket
		mockCoder mockCoder
		mockMsgr mockMessenger
		params []any
	}{
		{
			name: "success",
			opts: &models.Options{
				RetryAmount: 3,
				SendBufMinLen: 1024,
				ReceiveBufMinLen: 1024,
			},
			expectedErr: nil,
			mockPoll: mockPoller{},
			mockSock: mockSocket{},
			mockCoder: mockCoder{
				encodeFunc: func(sb *models.SendBuf, a []any) error {
					return nil
				},
				decodeFunc: func(b []byte) (any, error) {
					return "", nil
				},
			},
			mockMsgr: mockMessenger{
				sendFunc: func(b []byte) (int, error) {
					return 10, nil
				},
				receiveFunc: func(rb *models.RecvBuf) error {
					return nil
				},
				changeSocketFunc: func(i int) {},
			},
			params: []any{"GET", "test"},
		},
		{
			name: "fail ErrConnectionCmdInProcess",
			opts: &models.Options{
				RetryAmount: 3,
				SendBufMinLen: 1024,
				ReceiveBufMinLen: 1024,
			},
			expectedErr: models.ErrConnectionCmdInProcess,
			setUpFunc: func() {
				testConnection.processing = true
			},
			cleanUpFunc: func() {
				testConnection.processing = false
			},
			mockPoll: mockPoller{},
			mockSock: mockSocket{},
			mockCoder: mockCoder{
				encodeFunc: func(sb *models.SendBuf, a []any) error {
					return nil
				},
				decodeFunc: func(b []byte) (any, error) {
					return "", nil
				},
			},
			mockMsgr: mockMessenger{
				sendFunc: func(b []byte) (int, error) {
					return 10, nil
				},
				receiveFunc: func(rb *models.RecvBuf) error {
					return nil
				},
				changeSocketFunc: func(i int) {},
			},
			params: []any{"GET", "test"},
		},
		{
			name: "fail on send with bufs clear",
			opts: &models.Options{
				RetryAmount: 3,
				SendBufMinLen: 1024,
				ReceiveBufMinLen: 1024,
			},
			expectedErr: models.ErrSendNoAccess,
			mockPoll: mockPoller{},
			mockSock: mockSocket{},
			mockCoder: mockCoder{
				encodeFunc: func(sb *models.SendBuf, a []any) error {
					testConnection.sendBuf.WritePos = 356
					
					return nil
				},
				decodeFunc: func(b []byte) (any, error) {
					return "", nil
				},
			},
			mockMsgr: mockMessenger{
				sendFunc: func(b []byte) (int, error) {
					return 0, models.ErrSendNoAccess
				},
				receiveFunc: func(rb *models.RecvBuf) error {
					return nil
				},
				changeSocketFunc: func(i int) {},
			},
			params: []any{"GET", "test"},
		},
		{
			name: "fail on decode with bufs clear",
			opts: &models.Options{
				RetryAmount: 3,
				SendBufMinLen: 1024,
				ReceiveBufMinLen: 1024,
			},
			expectedErr: models.ErrSendNoAccess,
			mockPoll: mockPoller{},
			mockSock: mockSocket{},
			mockCoder: mockCoder{
				encodeFunc: func(sb *models.SendBuf, a []any) error {
					testConnection.sendBuf.WritePos = 356
					
					return nil
				},
				decodeFunc: func(b []byte) (any, error) {
					return nil, models.ErrSendNoAccess
				},
			},
			mockMsgr: mockMessenger{
				sendFunc: func(b []byte) (int, error) {
					return 356, nil
				},
				receiveFunc: func(rb *models.RecvBuf) error {
					testConnection.recvBuf.WritePos = 725
					
					return nil
				},
				changeSocketFunc: func(i int) {},
			},
			params: []any{"GET", "test"},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testConnection.opts = tt.opts
			testConnection.poller = &tt.mockPoll
			testConnection.socket = tt.mockSock
			testConnection.coder = tt.mockCoder
			testConnection.msgr = tt.mockMsgr
			testConnection.sendBuf = &models.SendBuf{
				Buf: make([]byte, testConnection.opts.SendBufMinLen),
			}
			testConnection.recvBuf = &models.RecvBuf{
				Buf: make([]byte, testConnection.opts.ReceiveBufMinLen),
			}

			if tt.setUpFunc != nil {
				tt.setUpFunc()
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
			defer cancel()

			_, err := testConnection.Process(ctx, tt.params) // не проверяем результат, т.к. он полностью моковый
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}

			if tt.cleanUpFunc != nil {
				tt.cleanUpFunc()
			}

			if testConnection.sendBuf.WritePos != 0 {
				t.Error("Буфер отправки не сброшен")
			}
			if testConnection.recvBuf.WritePos != 0 {
				t.Error("Буфер получения не сброшен")
			}
		})
	}
}