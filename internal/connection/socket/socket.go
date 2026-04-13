package socket

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rseleznev/redis_driver/internal/models"
)

type Socket int

// New создает новый сокет
func New(tcpSendBufLen, tcpRcvBufLen int) (Socket, error) {
	// Создаем сокет
	socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM | syscall.SOCK_NONBLOCK, syscall.IPPROTO_TCP)
	if err != nil {
		// EACCES Permission to create a socket of the specified type and/or protocol is denied.
		if errors.Is(err, syscall.EACCES) {
			return 0, models.ErrSocketNoAccess
		}

		// EMFILE The per-process limit on the number of open file descriptors has been reached.
		if errors.Is(err, syscall.EMFILE) {
			return 0, models.ErrTooManyFilesInProcess
		}

		// ENFILE The system-wide limit on the total number of open files has been reached.
		if errors.Is(err, syscall.ENFILE) {
			return 0, models.ErrTooManyFilesInSystem
		}

		// ENOBUFS or ENOMEM
		// 		Insufficient  memory  is available.  The socket cannot be created until sufficient resources are
		// 		freed.
		if errors.Is(err, syscall.ENOBUFS) {
			return 0, models.ErrSocketNoMemory
		}
		if errors.Is(err, syscall.ENOMEM) {
			return 0, models.ErrSocketNoMemory
		}

		// Еще могут быть:
		// EAFNOSUPPORT The implementation does not support the specified address family.
		// EINVAL Unknown protocol, or protocol family not available.
		// EINVAL Invalid flags in type.
		// EPROTONOSUPPORT The protocol type or the specified protocol is not supported within this domain.

		return 0, fmt.Errorf("socket creation err: %w", err)
	}

	// Включаем keep alive
	syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)

	// Настраиваем проверку keep alive
	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, 300) // через 5 минут без активности...
	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, 60) // отправляется тестовый пакет, ждем 60 секунд...
	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, 5) // после 5 неудачных попыток соединение закрывается

	syscall.SetsockoptInt(socketFd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1) // отключаем задержки

	// Настройки TCP-буферов ядра
	// надо покрыть тестами!
	if tcpSendBufLen > 0 {
		err = syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, tcpSendBufLen)
		if err != nil {
			return 0, err
		}
	}
	if tcpRcvBufLen > 0 {
		err = syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, tcpRcvBufLen)
		if err != nil {
			return 0, err
		}
	}
	

	return Socket(socketFd), nil
}

func (s Socket) GetSocketFd() int {
	return int(s)
}

// Connect подключает созданный сокет
func (s Socket) Connect(opts *models.Options) error {
	port := opts.RedisPort
	ip := opts.RedisIp
	
	// Адрес сервера
	var addr syscall.SockaddrInet4
	addr.Port = port // нужно будет сделать валидацию
	addr.Addr = ip // нужно будет сделать валидацию

	// Подключение сокета
	err := syscall.Connect(s.GetSocketFd(), &addr)
	if err != nil {
		// EACCES For  UNIX  domain  sockets,  which are identified by pathname: Write permission is denied on the
		// 		socket file, or search permission is denied for one of the directories in the path prefix.  (See
		// 		also path_resolution(7).)
		if errors.Is(err, syscall.EACCES) {
			return models.ErrSocketNoAccess
		}

		// EACCES, EPERM
		// 		The  user  tried  to connect to a broadcast address without having the socket broadcast flag en‐
		// 		abled or the connection request failed because of a local firewall rule.

		// 		EACCES can also be returned if an SELinux policy denied a connection (for example, if there is a
		// 		policy saying that an HTTP proxy can only connect to ports associated with HTTP servers, and the
		// 		proxy tries to connect to a different port).  dd
		if errors.Is(err, syscall.EPERM) {
			return models.ErrSocketNoAccess
		}

		// EADDRINUSE Local address is already in use.
		if errors.Is(err, syscall.EADDRINUSE) {
			return models.ErrSocketLocalPortInUse // точно правильно понял ошибку?
		}

		// EADDRNOTAVAIL
		// 		(Internet domain sockets) The socket referred to by sockfd had not previously been bound  to  an
		// 		address  and,  upon  attempting to bind it to an ephemeral port, it was determined that all port
		// 		numbers  in  the  ephemeral  port  range  are  currently  in  use.   See   the   discussion   of
		// 		/proc/sys/net/ipv4/ip_local_port_range in ip(7).
		if errors.Is(err, syscall.EADDRNOTAVAIL) {
			return models.ErrSocketNoLocalPorts // точно правильно понял ошибку?
		}

		// EAFNOSUPPORT The passed address didn't have the correct address family in its sa_family field.
		if errors.Is(err, syscall.EAFNOSUPPORT) {
			return models.ErrAddrBadParams
		}

		// EAGAIN For  nonblocking  UNIX  domain  sockets, the socket is nonblocking, and the connection cannot be
		// 		completed immediately.  For other socket families, there are insufficient entries in the routing
		// 		cache.
		// игнорируем эту ошибку

		// EALREADY The socket is nonblocking and a previous connection attempt has not yet been completed.
		if errors.Is(err, syscall.EALREADY) {
			return models.ErrConnectionInProcess
		}

		// EBADF  sockfd is not a valid open file descriptor.
		if errors.Is(err, syscall.EBADF) {
			return models.ErrSocketBadFD
		}

		// ECONNREFUSED A connect() on a stream socket found no one listening on the remote address.
		// в неблокирующем режиме этой ошибки быть не должно

		// EFAULT The socket structure address is outside the user's address space.
		if errors.Is(err, syscall.EFAULT) {
			return models.ErrSocketNoAccess
		}

		// EINPROGRESS
		// 		The  socket  is  nonblocking  and  the connection cannot be completed immediately.  (UNIX domain
		// 		sockets failed with EAGAIN instead.)  It is possible to select(2) or poll(2) for  completion  by
		// 		selecting  the  socket for writing.  After select(2) indicates writability, use getsockopt(2) to
		// 		read the SO_ERROR option at level SOL_SOCKET to determine whether connect()  completed  success‐
		// 		fully  (SO_ERROR  is  zero)  or  unsuccessfully (SO_ERROR is one of the usual error codes listed
		// 		here, explaining the reason for the failure).
		// игнорируем эту ошибку

		// EINTR  The system call was interrupted by a signal that was caught; see signal(7).
		if errors.Is(err, syscall.EINTR) {
			return models.ErrSignalInterruption
		}

		// EISCONN The socket is already connected.
		if errors.Is(err, syscall.EISCONN) {
			return nil
		}

		// ENETUNREACH Network is unreachable.
		// в неблокирующем режиме этой ошибки быть не должно

		// ENOTSOCK The file descriptor sockfd does not refer to a socket.
		if errors.Is(err, syscall.ENOTSOCK) {
			return models.ErrSocketBadFD
		}

		// EPROTOTYPE
		// 		The socket type does not support the requested communications protocol.  This error  can  occur,
		// 		for example, on an attempt to connect a UNIX domain datagram socket to a stream socket.

		// ETIMEDOUT
		// 		Timeout  while  attempting  connection.   The  server may be too busy to accept new connections.
		// 		Note that for IP sockets the timeout may be very long when syncookies are enabled on the server.
		// в неблокирующем режиме этой ошибки быть не должно
		
		return fmt.Errorf("socket connection err: %w", err)
	}
	
	return nil
}