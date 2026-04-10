package command

import (
	"errors"
	"testing"

	"github.com/rseleznev/redis_driver/internal/models"
)

var testProcessor = &commandProcessor{}

var (
	testErr = errors.New("test error")
)

type mockEnc struct{
	encodeFunc func([]byte, []any) ([]byte, error)
}

func (me *mockEnc) Encode(b []byte, d []any) ([]byte, error) {
	return me.encodeFunc(b, d)
}


type mockDec struct{
	decodeFunc func([]byte) (any, error)
}

func (md *mockDec) Decode(b []byte) (any, error) {
	return md.decodeFunc(b)
}



type mockConn struct{
	getSendBufFunc func() *models.SendBuf
	sendAndReceiveFunc func(*models.SendBuf) (*models.RecvBuf, error)
	drainRecvBufFunc func(*models.RecvBuf)
}

func (mc *mockConn) GetSendBuf() *models.SendBuf {
	return mc.getSendBufFunc()
}

func (mc *mockConn) SendAndReceive(sBuf *models.SendBuf) (*models.RecvBuf, error) {
	return mc.sendAndReceiveFunc(sBuf)
}

func (mc *mockConn) DrainRecvBuf(rBuf *models.RecvBuf) {
	mc.drainRecvBufFunc(rBuf)
}


func Test_sendAndReceive(t *testing.T) {
	testData := []struct{
		name string
		cmd command
		conn mockConn
		encoder mockEnc
		decoder mockDec
		expectedErr error
		expectedResult any
	}{
		{
			name: "success nil",
			cmd: command{
				resultValueChan: make(chan any),
				resultErrChan: make(chan error),
			},
			conn: mockConn{
				getSendBufFunc: func() *models.SendBuf {
					return &models.SendBuf{
						SocketFd: 5,
						Buf: make([]byte, 0, 20),
					}
				},
				sendAndReceiveFunc: func(sb *models.SendBuf) (*models.RecvBuf, error) {
					return &models.RecvBuf{
						SocketFd: 5,
						Buf: make([]byte, 20),
					}, nil
				},
				drainRecvBufFunc: func(rb *models.RecvBuf) {},
			},
			encoder: mockEnc{
				encodeFunc: func(b []byte, a []any) ([]byte, error) {
					testEncodedData := []byte{'T', 'E', 'S', 'T'}
					return testEncodedData, nil
				},
			},
			decoder: mockDec{
				decodeFunc: func(b []byte) (any, error) {
					return nil, nil
				},
			},
			expectedErr: nil,
			expectedResult: nil,
		},
		{
			name: "success OK",
			cmd: command{
				resultValueChan: make(chan any),
				resultErrChan: make(chan error),
			},
			conn: mockConn{
				getSendBufFunc: func() *models.SendBuf {
					return &models.SendBuf{
						SocketFd: 5,
						Buf: make([]byte, 0, 20),
					}
				},
				sendAndReceiveFunc: func(sb *models.SendBuf) (*models.RecvBuf, error) {
					return &models.RecvBuf{
						SocketFd: 5,
						Buf: make([]byte, 20),
					}, nil
				},
				drainRecvBufFunc: func(rb *models.RecvBuf) {},
			},
			encoder: mockEnc{
				encodeFunc: func(b []byte, a []any) ([]byte, error) {
					testEncodedData := []byte{'T', 'E', 'S', 'T'}
					return testEncodedData, nil
				},
			},
			decoder: mockDec{
				decodeFunc: func(b []byte) (any, error) {
					return "OK", nil
				},
			},
			expectedErr: nil,
			expectedResult: "OK",
		},
		{
			name: "fail encode",
			cmd: command{
				resultValueChan: make(chan any),
				resultErrChan: make(chan error),
			},
			conn: mockConn{
				getSendBufFunc: func() *models.SendBuf {
					return &models.SendBuf{
						SocketFd: 5,
						Buf: make([]byte, 0, 20),
					}
				},
				sendAndReceiveFunc: func(sb *models.SendBuf) (*models.RecvBuf, error) {
					return &models.RecvBuf{
						SocketFd: 5,
						Buf: make([]byte, 20),
					}, nil
				},
				drainRecvBufFunc: func(rb *models.RecvBuf) {},
			},
			encoder: mockEnc{
				encodeFunc: func(b []byte, a []any) ([]byte, error) {
					testEncodedData := []byte{'T', 'E', 'S', 'T'}
					return testEncodedData, testErr
				},
			},
			decoder: mockDec{
				decodeFunc: func(b []byte) (any, error) {
					return nil, nil
				},
			},
			expectedErr: testErr,
			expectedResult: nil,
		},
		{
			name: "fail sendAndReceive",
			cmd: command{
				resultValueChan: make(chan any),
				resultErrChan: make(chan error),
			},
			conn: mockConn{
				getSendBufFunc: func() *models.SendBuf {
					return &models.SendBuf{
						SocketFd: 5,
						Buf: make([]byte, 0, 20),
					}
				},
				sendAndReceiveFunc: func(sb *models.SendBuf) (*models.RecvBuf, error) {
					return &models.RecvBuf{
						SocketFd: 5,
						Buf: make([]byte, 20),
					}, testErr
				},
				drainRecvBufFunc: func(rb *models.RecvBuf) {},
			},
			encoder: mockEnc{
				encodeFunc: func(b []byte, a []any) ([]byte, error) {
					testEncodedData := []byte{'T', 'E', 'S', 'T'}
					return testEncodedData, nil
				},
			},
			decoder: mockDec{
				decodeFunc: func(b []byte) (any, error) {
					return nil, nil
				},
			},
			expectedErr: testErr,
			expectedResult: nil,
		},
		{
			name: "fail decode",
			cmd: command{
				resultValueChan: make(chan any),
				resultErrChan: make(chan error),
			},
			conn: mockConn{
				getSendBufFunc: func() *models.SendBuf {
					return &models.SendBuf{
						SocketFd: 5,
						Buf: make([]byte, 0, 20),
					}
				},
				sendAndReceiveFunc: func(sb *models.SendBuf) (*models.RecvBuf, error) {
					return &models.RecvBuf{
						SocketFd: 5,
						Buf: make([]byte, 20),
					}, nil
				},
				drainRecvBufFunc: func(rb *models.RecvBuf) {},
			},
			encoder: mockEnc{
				encodeFunc: func(b []byte, a []any) ([]byte, error) {
					testEncodedData := []byte{'T', 'E', 'S', 'T'}
					return testEncodedData, nil
				},
			},
			decoder: mockDec{
				decodeFunc: func(b []byte) (any, error) {
					return nil, testErr
				},
			},
			expectedErr: testErr,
			expectedResult: nil,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testProcessor.connector = &tt.conn
			testProcessor.enc = &tt.encoder
			testProcessor.dec = &tt.decoder

			go testProcessor.sendAndReceive(&tt.cmd)

			select {
			case err := <-tt.cmd.resultErrChan:
				if err != tt.expectedErr {
					t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
				}

			case res := <-tt.cmd.resultValueChan:
				if res != tt.expectedResult {
					t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, res)
				}

			}
		})
	}
}