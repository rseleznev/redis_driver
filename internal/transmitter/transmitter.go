package transmitter

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/models"
	"golang.org/x/sys/unix"
)

type Transmitter struct {
	SocketFd int
}

func NewTransmitter(socketFd int) *Transmitter {
	return &Transmitter{
		SocketFd: socketFd,
	}
}

func (t *Transmitter) Send(data []byte) (int, error) {
	// Системный вызов для отправки
	n, err := syscall.SendmsgN(t.SocketFd, data, nil, nil, 0)
	if err != nil {
		// EACCES (For  UNIX domain sockets, which are identified by pathname) Write permission is denied on the destina‐
		// 		tion socket file, or search permission is denied for one of the  directories  the  path  prefix.   (See
		// 		path_resolution(7).)

		// 		(For UDP sockets) An attempt was made to send to a network/broadcast address as though it was a unicast
		// 		address.
		if errors.Is(err, syscall.EACCES) {
			return n, models.ErrSendNoAccess
		}


		// EAGAIN or EWOULDBLOCK
		// 		The socket is marked nonblocking and the requested operation would block.  POSIX.1-2001  allows  either
		// 		error  to  be returned for this case, and does not require these constants to have the same value, so a
		// 		portable application should check for both possibilities.
		// игнорируем ошибку

		// EAGAIN (Internet domain datagram sockets) The socket referred to by sockfd had not previously been bound to an
		// 		address  and,  upon attempting to bind it to an ephemeral port, it was determined that all port numbers
		// 		in the ephemeral port range are currently in use.   See  the  discussion  of  /proc/sys/net/ipv4/ip_lo‐
		// 		cal_port_range in ip(7).

		// EALREADY Another Fast Open is in progress.

		// EBADF  sockfd is not a valid open file descriptor.
		if errors.Is(err, syscall.EBADF) {
			return n, models.ErrSocketBadFD
		}

		// ECONNRESET Connection reset by peer.
		if errors.Is(err, syscall.ECONNRESET) {
			return n, models.ErrConnectionReset
		}

		// EDESTADDRREQ The socket is not connection-mode, and no peer address is set.

		// EFAULT An invalid user space address was specified for an argument.
		if errors.Is(err, syscall.EFAULT) {
			return n, models.ErrSpaceAddress
		}

		// EINTR  A signal occurred before any data was transmitted; see signal(7).
		if errors.Is(err, syscall.EINTR) {
			return n, models.ErrSignalInterruption
		}

		// EINVAL Invalid argument passed.
		if errors.Is(err, syscall.EINVAL) {
			return n, models.ErrBadValue
		}

		// EISCONN
		// 		The connection-mode socket was connected already but a recipient was specified.  (Now either this error
		// 		is returned, or the recipient specification is ignored.)

		// EMSGSIZE
		// 		The socket type requires that message be sent atomically, and the size of the message to be  sent  made
		// 		this impossible.

		// ENOBUFS
		// 		The  output  queue  for  a network interface was full.  This generally indicates that the interface has
		// 		stopped sending, but may be caused by transient congestion.  (Normally, this does not occur  in  Linux.
		// 		Packets are just silently dropped when a device queue overflows.)

		// ENOMEM No memory available.
		if errors.Is(err, syscall.ENOMEM) {
			return n, models.ErrNoMemory
		}

		// ENOTCONN The socket is not connected, and no target has been given.
		if errors.Is(err, syscall.ENOTCONN) {
			return n, models.ErrNotConnected
		}

		// ENOTSOCK The file descriptor sockfd does not refer to a socket.
		if errors.Is(err, syscall.ENOTSOCK) {
			return n, models.ErrSocketBadFD
		}

		// EOPNOTSUPP Some bit in the flags argument is inappropriate for the socket type.

		// EPIPE  The  local end has been shut down on a connection oriented socket.  In this case, the process will also
		// 		receive a SIGPIPE unless MSG_NOSIGNAL is set
		if errors.Is(err, syscall.EPIPE) {
			return n, models.ErrConnectionClosed
		}

		
		return n, fmt.Errorf("sending err: %w", err)
	}
	if n != len(data) {
		return n, models.ErrSendMsgTrunc
	}
	
	return n, nil
}

func (t *Transmitter) Receive(buf *models.RecvBuf) error {
	// Читаем ответ
	n, _, coreFlags, _, err := syscall.Recvmsg(t.SocketFd, buf.Buf[buf.WritePos:], nil, 0)
	if err != nil {
		// EAGAIN or EWOULDBLOCK
		// 		The  socket  is marked nonblocking and the receive operation would block, or a receive timeout had been
		// 		set and the timeout expired before data was received.  POSIX.1 allows either error to be  returned  for
		// 		this  case,  and  does  not  require  these constants to have the same value, so a portable application
		// 		should check for both possibilities.
		// игнорируем и обрабатываем выше по стеку

		// EBADF  The argument sockfd is an invalid file descriptor.
		if errors.Is(err, syscall.EBADF) {
			return models.ErrSocketBadFD
		}

		// ECONNREFUSED
		// 		A remote host refused to allow the network connection (typically because it  is  not  running  the  re‐
		// 		quested service).
		if errors.Is(err, syscall.ECONNREFUSED) {
			return models.ErrConnectionRefused
		}

		// EFAULT The receive buffer pointer(s) point outside the process's address space.
		if errors.Is(err, syscall.EFAULT) {
			return models.ErrSpaceAddress
		}

		// EINTR  The receive was interrupted by delivery of a signal before any data was available; see signal(7).
		if errors.Is(err, syscall.EINTR) {
			return models.ErrSignalInterruption
		}

		// EINVAL Invalid argument passed.
		if errors.Is(err, syscall.EINVAL) {
			return models.ErrBadValue
		}

		// ENOMEM Could not allocate memory for recvmsg().
		if errors.Is(err, syscall.ENOMEM) {
			return models.ErrNoMemory
		}

		// ENOTCONN
		// 		The socket is associated with a connection-oriented protocol and has not been connected (see connect(2)
		// 		and accept(2)).
		if errors.Is(err, syscall.ENOTCONN) {
			return models.ErrNotConnected
		}

		// ENOTSOCK The file descriptor sockfd does not refer to a socket
		if errors.Is(err, syscall.ENOTSOCK) {
			return models.ErrSocketBadFD
		}
		
		return fmt.Errorf("receiving err: %w", err)
	}

	// Проверяем флаги ядра
	// Проверка, все ли данные влезли в буфер получения
	if coreFlags & syscall.MSG_TRUNC != 0 {
		return models.ErrRecvMsgTrunc
	}
	// Доп проверка, не должна срабатывать
	if coreFlags & syscall.MSG_CTRUNC != 0 {
		return models.ErrRecvMsgCTrunc
	}
	
	// Соединение закрыто сервером
	if n == 0 {
		return models.ErrConnectionClosed
	}

	// Указываем позицию окончания данных
	buf.WritePos += n

	// Проверяем, что в буфере получения не осталось данных
	if n == len(buf.Buf) {
		bytesLeft, _ := unix.IoctlGetInt(t.SocketFd, unix.SIOCINQ)
		if bytesLeft > 0 {
			return models.ErrRecvMsgTrunc
		}
	}

	return nil
}

func (t *Transmitter) ChangeSocketFd(newSocketFd int) {
	t.SocketFd = newSocketFd
}