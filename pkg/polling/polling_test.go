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
	}{
		{
			name: "success connect",
			expectedErr: nil,
			eventForPolling: models.PollingUnit{
				SocketFd: 5,
				EventType: "connect",
				ResultChan: make(chan error),
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			mockSys := mockSyscalls{
				waitFunc: func(eFd int, event []syscall.EpollEvent, timeout int) (int, error) {
					return 1, nil
				},
				getSocketOptFunc: func(i1, i2, i3 int) (int, error) {
					return 0, nil
				},
				ctlFunc: func(i1, i2, i3 int, ee *syscall.EpollEvent) error {
					return nil
				},
			}

			testPoller.sys = &mockSys

			err := testPoller.Add(tt.eventForPolling)
			if err != tt.expectedErr {
				t.Error(err)
			}

			err = <-tt.eventForPolling.ResultChan
			if err != tt.expectedErr {
				t.Error(err)
			}
		})
	}
}