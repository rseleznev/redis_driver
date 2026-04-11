package command

import (
	"context"
	"maps"
	"testing"
	"time"
)

var testBuilder = &commandBuilder{}

type mockProcessor struct {
	sendAndReceiveFunc func(command)
}

func (mp *mockProcessor) sendAndReceive(cmd command) {
	mp.sendAndReceiveFunc(cmd)
}

func TestPing(t *testing.T) {
	testData := []struct{
		name string
		expectedErr error
		mockProc mockProcessor
	}{
		{
			name: "success",
			expectedErr: nil,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultValueChan <- nil
				},
			},
		},
		{
			name: "fail err",
			expectedErr: testErr,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultErrChan <- testErr
				},
			},
		},
		{
			name: "fail timeout",
			expectedErr: context.DeadlineExceeded,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					time.Sleep(time.Second*2)
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testBuilder.proc = &tt.mockProc
			
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			err := testBuilder.Ping(ctx)

			defer cancel()

			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}	
		})
	}
}

func TestHello(t *testing.T) {
	testData := []struct{
		name string
		expectedErr error
		expectedResult map[string]string
		mockProc mockProcessor
	}{
		{
			name: "success",
			expectedErr: nil,
			expectedResult: map[string]string{
				"test": "OK",
			},
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultValueChan <- map[string]string{
						"test": "OK",
					}
				},
			},
		},
		{
			name: "fail err",
			expectedErr: testErr,
			expectedResult: nil,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultErrChan <- testErr
				},
			},
		},
		{
			name: "fail timeout",
			expectedErr: context.DeadlineExceeded,
			expectedResult: nil,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					time.Sleep(time.Second*2)
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testBuilder.proc = &tt.mockProc
			
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			res, err := testBuilder.Hello(ctx)

			defer cancel()

			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
			if !maps.Equal(res, tt.expectedResult) {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, res)
			}
		})
	}
}

func TestSet(t *testing.T) {
	testData := []struct{
		name, key string
		value any
		dur time.Duration
		expectedErr error
		mockProc mockProcessor
	}{
		{
			name: "success no duration",
			key: "test",
			value: "OK",
			expectedErr: nil,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultValueChan <- nil
				},
			},
		},
		{
			name: "success with duration",
			key: "test",
			value: "OK",
			dur: time.Second*1,
			expectedErr: nil,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultValueChan <- nil
				},
			},
		},
		{
			name: "fail err",
			key: "test",
			value: "ERR",
			dur: time.Second*1,
			expectedErr: testErr,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultErrChan <- testErr
				},
			},
		},
		{
			name: "fail timeout",
			key: "test",
			value: "ERR",
			dur: time.Second*1,
			expectedErr: context.DeadlineExceeded,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					time.Sleep(time.Second*2)
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testBuilder.proc = &tt.mockProc
			
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			err := testBuilder.Set(ctx, tt.key, tt.value, tt.dur)

			defer cancel()

			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	testData := []struct{
		name, key string
		expectedErr error
		expectedResult any
		mockProc mockProcessor
	}{
		{
			name: "success",
			key: "test",
			expectedErr: nil,
			expectedResult: "OK",
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultValueChan <- "OK"
				},
			},
		},
		{
			name: "fail err",
			key: "test",
			expectedErr: testErr,
			expectedResult: nil,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					c.resultErrChan <- testErr
				},
			},
		},
		{
			name: "fail timeout",
			key: "test",
			expectedErr: context.DeadlineExceeded,
			expectedResult: nil,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c command) {
					time.Sleep(time.Second*2)
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testBuilder.proc = &tt.mockProc

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			res, err := testBuilder.Get(ctx, tt.key)

			defer cancel()

			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}
			if res != tt.expectedResult {
				t.Errorf("Ожидаемый результат %s, получено %s", tt.expectedResult, res)
			}
		})
	}
}