package polling

import (
	"sync"
	"syscall"
	"testing"

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
		expectedErr error
		eventForPolling models.PollingUnit
		mockSys mockSyscalls
	}{
		{
			name: "success connect",
			expectedErr: nil,
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
			expectedErr: nil,
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
			expectedErr: nil,
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
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			testPoller.sys = &tt.mockSys

			err := testPoller.Add(tt.eventForPolling)
			if err != tt.expectedErr {
				t.Error(err)
			}

			err = <-tt.eventForPolling.ResultChan
			if err != tt.expectedErr {
				t.Error(err)
			}

			err = testPoller.GetError()
			if err != tt.expectedErr {
				t.Error(err)
			}
		})
	}
}