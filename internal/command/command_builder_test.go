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
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testBuilder.proc = &tt.mockProc
			
			var err error
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			go func() {
				err = testBuilder.Ping(ctx)
			}()
			if err != tt.expectedErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedErr, err)
			}

			cancel()		
		})
	}
}