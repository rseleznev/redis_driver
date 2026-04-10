package command

import (
	"context"
	"testing"
	"time"
)

var testBuilder = &commandBuilder{}

type mockProcessor struct {
	sendAndReceiveFunc func(*command)
}

func (mp *mockProcessor) sendAndReceive(cmd *command) {
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
				sendAndReceiveFunc: func(c *command) {
					c.resultValueChan <- nil
				},
			},
		},
		{
			name: "fail err",
			expectedErr: testErr,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c *command) {
					c.resultErrChan <- testErr
				},
			},
		},
		{
			name: "fail timeout",
			expectedErr: context.DeadlineExceeded,
			mockProc: mockProcessor{
				sendAndReceiveFunc: func(c *command) {
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