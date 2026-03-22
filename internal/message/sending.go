package message

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/epoll"
	"github.com/rseleznev/redis_driver/internal/models"
)

// Send отправляет данные по указанному сокету
func Send(socketFd int, data []byte) error {
	var err error
	
	for {
		err = trySend(socketFd, data)
		if err != nil {
			if errors.Is(err, syscall.EWOULDBLOCK) {
				epoll.Wait()
				continue
			}
			// также нужно проверять ErrMsgSndTrunc
			fmt.Println(err)
		}
		break
	}

	return err
}

// trySend делает одну попытку отправить данные и выполняет проверки, если данные отправлены
func trySend(socketFd int, data []byte) error {
	// Системный вызов для отправки
	n, err := syscall.SendmsgN(socketFd, data, nil, nil, 0)
	if err != nil {
		// EACCES (For  UNIX domain sockets, which are identified by pathname) Write permission is denied on the destina‐
		// 		tion socket file, or search permission is denied for one of the  directories  the  path  prefix.   (See
		// 		path_resolution(7).)

		// 		(For UDP sockets) An attempt was made to send to a network/broadcast address as though it was a unicast
		// 		address.
		if errors.Is(err, syscall.EACCES) {
			return models.ErrSendNoAccess
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
			return models.ErrSocketBadFD
		}

		// ECONNRESET Connection reset by peer.
		if errors.Is(err, syscall.ECONNRESET) {
			return models.ErrConnectionReset
		}

		// EDESTADDRREQ The socket is not connection-mode, and no peer address is set.

		// EFAULT An invalid user space address was specified for an argument.
		if errors.Is(err, syscall.EFAULT) {
			return models.ErrSpaceAddress
		}

		// EINTR  A signal occurred before any data was transmitted; see signal(7).
		if errors.Is(err, syscall.EINTR) {
			return models.ErrSignalInterruption
		}

		// EINVAL Invalid argument passed.
		if errors.Is(err, syscall.EINVAL) {
			return models.ErrBadValue
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
			return models.ErrNoMemory
		}

		// ENOTCONN The socket is not connected, and no target has been given.
		if errors.Is(err, syscall.ENOTCONN) {
			return models.ErrNotConnected
		}

		// ENOTSOCK The file descriptor sockfd does not refer to a socket.
		if errors.Is(err, syscall.ENOTSOCK) {
			return models.ErrSocketBadFD
		}

		// EOPNOTSUPP Some bit in the flags argument is inappropriate for the socket type.

		// EPIPE  The  local end has been shut down on a connection oriented socket.  In this case, the process will also
		// 		receive a SIGPIPE unless MSG_NOSIGNAL is set
		if errors.Is(err, syscall.EPIPE) {
			return models.ErrConnectionClosed
		}

		
		return fmt.Errorf("sending err: %w", err)
	}
	if n != len(data) {
		return models.ErrSendMsgTrunc
	}
	
	return nil
}