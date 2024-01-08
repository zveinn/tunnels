package tunnels

import (
	"encoding/binary"
	"net"
	"syscall"
	"unsafe"

	"github.com/jackpal/gateway"
)

// HostToNetShort converts a 16-bit integer from host to network byte order, aka "htons"
func HostToNetShort(i uint16) uint16 {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return binary.BigEndian.Uint16(b)
}

func FindGateway() (net.IP, error) {
	return gateway.DiscoverGateway()
}

//	func CreateBuffer(length int) (buffer []byte, buffPtr uintptr, buffLenPtr uintptr) {
//		buffer = make([]byte, length)
//		buffPtr = uintptr(unsafe.Pointer(&buffer[0]))
//		buffLenPtr = uintptr(len(buffer))
//		return
//	}

type SockaddrInet4 struct {
	Port uint16
	Addr [4]byte
	raw  RawSockaddrInet4
}

type RawSockaddrInet4 struct {
	Family uint16
	Port   uint16
	Addr   [4]byte /* in_addr */
	Zero   [8]uint8
}

func (sa *SockaddrInet4) sockaddr() (unsafe.Pointer, uint32, error) {
	// if sa.Port < 0 || sa.Port > 0xFFFF {
	// 	return nil, 0, syscall.EINVAL
	// }
	sa.raw.Family = syscall.AF_INET
	p := (*[2]byte)(unsafe.Pointer(&sa.raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	sa.raw.Addr = sa.Addr
	return unsafe.Pointer(&sa.raw), 0x10, nil
}
