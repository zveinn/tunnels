//go:build freebsd || linux || openbsd

package tunnels

import (
	"encoding/binary"
	"io"
	"syscall"
	"unsafe"
)

type RawSocket struct {
	Name          string
	IPv4Address   string
	IPv6Address   string
	InterfaceName string
	SocketBuffer  []byte

	Domain int
	Type   int
	Proto  int

	RWC io.ReadWriteCloser
}

func (r *RawSocket) Create() (err error) {
	fd, sockErr := syscall.Socket(
		r.Domain,
		r.Type,
		r.Proto,
	)
	if sockErr != nil {
		syscall.Close(fd)
		return sockErr
	}

	err = syscall.SetNonblock(fd, true)
	if err != nil {
		syscall.Close(fd)
		return err
	}

	if err := syscall.SetsockoptInt(
		fd,
		syscall.SOL_SOCKET,
		syscall.SO_REUSEADDR,
		1,
	); err != nil {
		syscall.Close(fd)
		return err
	}

	sfd, sockErr := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_RAW,
		syscall.IPPROTO_RAW,
	)
	if sockErr != nil {
		syscall.Close(fd)
		return sockErr
	}

	err = syscall.BindToDevice(fd, r.InterfaceName)
	if err != nil {
		syscall.Close(fd)
		panic(err)
	}

	addr := syscall.RawSockaddrInet4{
		Family: syscall.AF_INET,
	}

	r.RWC = &RWC{
		fd:         fd,
		fdPtr:      uintptr(fd),
		buffPtr:    uintptr(unsafe.Pointer(&r.SocketBuffer[0])),
		buffLenPtr: uintptr(len(r.SocketBuffer)),

		sfd:        sfd,
		sfdPtr:     uintptr(sfd),
		addr:       &addr,
		addrLenPtr: uintptr(0x10),
		addrPtr:    uintptr(unsafe.Pointer(&addr)),
	}

	return nil
}

type RWC struct {
	fd    int
	fdPtr uintptr

	// used for reading
	r0         uintptr
	e1         syscall.Errno
	buffPtr    uintptr
	buffLenPtr uintptr
	// msg specific reading

	// user for writing
	sfd        int
	sfdPtr     uintptr
	addr       *syscall.RawSockaddrInet4
	addrPtr    uintptr
	addrLenPtr uintptr
}

func (rwc *RWC) Read(data []byte) (n int, err error) {
	rwc.r0, _, rwc.e1 = syscall.Syscall6(
		syscall.SYS_RECVFROM,
		rwc.fdPtr,
		rwc.buffPtr,
		rwc.buffLenPtr,
		0,
		0,
		0,
	)
	n = int(rwc.r0)

	return
}

func (rwc *RWC) Write(data []byte) (n int, err error) {
	rwc.addr.Addr[0] = data[16]
	rwc.addr.Addr[1] = data[17]
	rwc.addr.Addr[2] = data[18]
	rwc.addr.Addr[3] = data[19]
	IHL := ((data[0] << 4) >> 4) * 4
	rwc.addr.Port = binary.BigEndian.Uint16(data[IHL+2 : IHL+4])
	_, _, e1 := syscall.Syscall6(
		syscall.SYS_SENDTO,
		rwc.sfdPtr,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		0,
		rwc.addrPtr,
		rwc.addrLenPtr,
	)
	if e1 != 0 {
		return 0, e1
	}
	return 0, nil
}

func (rwc *RWC) Close() error {
	return syscall.Close(rwc.fd)
}
