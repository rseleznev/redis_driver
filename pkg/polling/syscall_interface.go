package polling

import "syscall"

type epollSyscallInterface interface {
	New(int) (int, error)
	Wait(int, []syscall.EpollEvent, int) (int, error)
	GetSocketOpt(int, int, int) (int, error)
	Ctl(int, int, int, *syscall.EpollEvent) error
}

type epollSyscalls struct {}

func (es epollSyscalls) New(size int) (int, error) {
	return syscall.EpollCreate(size)
}

func (es epollSyscalls) Wait(eFd int, events []syscall.EpollEvent, timeout int) (int, error) {
	return syscall.EpollWait(eFd, events, timeout)
}

func (es epollSyscalls) GetSocketOpt(sFd, l, o int) (int, error) {
	return syscall.GetsockoptInt(sFd, l, o)
}

func (es epollSyscalls) Ctl(eFd, o, sFd int, event *syscall.EpollEvent) error {
	return syscall.EpollCtl(eFd, o, sFd, event)
}