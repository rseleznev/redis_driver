package polling

import (
	"context"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/rseleznev/redis_driver/internal/models"
)

func TestNewPoller(t *testing.T) {
	var err error
	
	_, err = NewPoller()
	if err != nil {
		t.Error(err)
	}
}

type mockSyscalls struct {
	waitFunc func(int, []syscall.EpollEvent, int) (int, error)
	getSocketOptFunc func(int, int, int) (int, error)
	ctlFunc func(int, int, int, *syscall.EpollEvent) error
}

func (ms *mockSyscalls) Wait(eFd int, events []syscall.EpollEvent, timeout int) (int, error) {
	return ms.waitFunc(eFd, events, timeout)
}

func (ms *mockSyscalls) GetSocketOpt(sFd, l, o int) (int, error) {
	return ms.getSocketOptFunc(sFd, l, o)
}

func (ms *mockSyscalls) Ctl(eFd, o, sFd int, event *syscall.EpollEvent) error {
	return ms.ctlFunc(eFd, o, sFd, event)
}

var testPoller = epoll{
	fd: 2,
	mu: sync.Mutex{},
	sockets: map[int]models.PollingUnit{},
}

func TestAdd(t *testing.T) {
	testData := []struct{
		name string
		setUpFunc func()
		cleanUpFunc func()
		expectedMethodErr error
		expectedChanErr error
		expectedPollerErr error
		eventForPolling models.PollingUnit
		mockSys mockSyscalls
	}{
		{
			name: "success connect",
			expectedMethodErr: nil,
			expectedChanErr: nil,
			expectedPollerErr: nil,

			eventForPolling: models.PollingUnit{
				SocketFd: 5,
				EventType: "connect",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return nil
				},
			},
		},
		{
			name: "success income",
			expectedMethodErr: nil,
			expectedChanErr: nil,
			expectedPollerErr: nil,

			eventForPolling: models.PollingUnit{
				SocketFd: 5,
				EventType: "income",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return nil
				},
			},
		},
		{
			name: "success outcome",
			expectedMethodErr: nil,
			expectedChanErr: nil,
			expectedPollerErr: nil,

			eventForPolling: models.PollingUnit{
				SocketFd: 5,
				EventType: "outcome",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return nil
				},
			},
		},
		{
			name: "fail ErrSocketAlreadyAdded",
			setUpFunc: func() {
				testPoller.sockets[7] = models.PollingUnit{}
			},
			cleanUpFunc: func() {
				delete(testPoller.sockets, 7)
			},
			expectedMethodErr: models.ErrSocketAlreadyAdded,
			expectedChanErr: nil,
			expectedPollerErr: nil,

			eventForPolling: models.PollingUnit{
				SocketFd: 7,
				EventType: "outcome",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return nil
				},
			},
		},
		{
			name: "fail ErrPollUnknownEventType",
			expectedMethodErr: models.ErrPollUnknownEventType,
			expectedChanErr: nil,
			expectedPollerErr: nil,

			eventForPolling: models.PollingUnit{
				SocketFd: 5,
				EventType: "test",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return nil
				},
			},
		},
		{
			name: "fail ctl ErrPollBadFD",
			expectedMethodErr: models.ErrPollBadFD,
			expectedChanErr: nil,
			expectedPollerErr: nil,

			eventForPolling: models.PollingUnit{
				SocketFd: 5,
				EventType: "connect",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return syscall.EBADF
				},
			},
		},
	}
	
	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testPoller.sys = &tt.mockSys

			if tt.setUpFunc != nil {
				tt.setUpFunc()
			}

			err := testPoller.Add(tt.eventForPolling)
			if err != tt.expectedMethodErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedMethodErr, err)
			}

			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*1)
			
			select {
			case err = <-tt.eventForPolling.ResultChan:
				if err != tt.expectedChanErr {
					t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedChanErr, err)
				}

			case <-ctx.Done():
				t.Log("Вышли из select по таймауту")
				
			}

			cancelFunc()

			err = testPoller.GetError()
			if err != tt.expectedPollerErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedPollerErr, err)
			}

			if tt.cleanUpFunc != nil {
				tt.cleanUpFunc()
			}
		})
	}
}

func Test_wait(t *testing.T) {
	testData := []struct{
		name string
		expectedChanErr error
		expectedPollerErr error
		eventForPolling models.PollingUnit
		mockSys mockSyscalls
	}{
		{
			name: "success",
			expectedChanErr: nil,
			expectedPollerErr: nil,
			eventForPolling: models.PollingUnit{
				SocketFd: 1,
				EventType: "outcome",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return nil
				},
			},
		},
		{
			name: "fail",
			expectedChanErr: models.ErrPollNoMemory,
			expectedPollerErr: models.ErrPollNoMemory,
			eventForPolling: models.PollingUnit{
				SocketFd: 1,
				EventType: "outcome",
				ResultChan: make(chan error),
			},
			mockSys: mockSyscalls{
				waitFunc: func(_ int, _ []syscall.EpollEvent, _ int) (int, error) {
					return 1, models.ErrPollNoMemory
				},
				getSocketOptFunc: func(_, _, _ int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(_, _, _ int, _ *syscall.EpollEvent) error {
					return nil
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testPoller.sys = &tt.mockSys

			testPoller.addOutcomeEvent(1)
			testPoller.addSocketInPolling(tt.eventForPolling)

			testPoller.wait()
			err := <-tt.eventForPolling.ResultChan
			if err != tt.expectedChanErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedChanErr, err)
			}

			err = testPoller.GetError()
			if err != tt.expectedPollerErr {
				t.Errorf("Ожидаемая ошибка %s, получено %s", tt.expectedPollerErr, err)
			}
		})
	}
}