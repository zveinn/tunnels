package tunnel

import (
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

type syscallAddAddr struct {
	Name [16]byte
	Addr [4]byte
	Pad  [8]byte
}

type syscallChangeMTU struct {
	Name [16]byte
	MTU  int32
}

type Interface struct {
	Name    string
	Address string
	RWC     io.ReadWriteCloser
	User    uint
	Group   uint
	FD      uintptr
}

func (IF *Interface) Syscall_MTU(mtu int32) (err error) {
	var ifr syscallChangeMTU
	copy(ifr.Name[:], []byte(IF.Name))
	ifr.MTU = mtu

	if err = ioctl(
		IF.FD,
		syscall.SIOCSIFMTU,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

func (IF *Interface) Syscall_Addr(ip string) (err error) {
	var addr [4]byte
	if ip == "" {
		ip = IF.Address
	}
	copy(addr[:], net.ParseIP(ip).To4())

	var ifr syscallAddAddr
	copy(ifr.Name[:], []byte(IF.Name))
	copy(ifr.Addr[:], addr[:])

	if err = ioctl(
		IF.FD,
		syscall.SIOCSIFADDR,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

func (IF *Interface) Addr(ip string) (err error) {
	out, err := exec.Command(
		"ip",
		"addr",
		"add",
		ip,
		"dev",
		IF.Name,
	).Output()
	if err != nil {
		log.Println("ERROR || ip addr add", ip, "dev", IF.Name, " || err:", err, " || rawout: ", string(out))
		return err
	}
	return
}

func (IF *Interface) Down() (err error) {
	out, err := exec.Command(
		"ip",
		"link",
		"set",
		"dev",
		IF.Name,
		"down",
	).Output()
	if err != nil {
		log.Println("ERROR || ip link set dev", IF.Name, "down || err:", err, " || rawout: ", string(out))
		return err
	}
	return
}

func (IF *Interface) Up() (err error) {
	out, err := exec.Command("ip", "link", "set", "dev", IF.Name, "up").Output()
	if err != nil {
		log.Println("ERROR || ip link set dev", IF.Name, "up || err:", err, " || rawout: ", string(out))
		return err
	}
	return
}

func (i *Interface) SetTXQueueLen(ql string) (err error) {
	out, err := exec.Command("ip", "link", "set", i.Name, "txqueuelen", ql).Output()
	if err != nil {
		log.Println("ERROR || ip link set", i.Name, "txqueuelen", ql, " || err:", err, " || rawout: ", string(out))
		return err
	}
	return
}

func (IF *Interface) SetMTU(mtu string) (err error) {
	out, err := exec.Command("ip", "link", "set", IF.Name, "mtu", mtu).Output()
	if err != nil {
		log.Println("ERROR || ip link set", IF.Name, "mtu", mtu, " || err:", err, " || rawout: ", string(out))
		return err
	}
	return
}

func (IF *Interface) AddRoute(network string, gateway string, metric string) (err error) {
	_ = IF.DelRoute(network, gateway, metric)
	out, err := exec.Command("ip", "route", "add", network, "via", gateway, "metric", metric).Output()
	if err != nil {
		log.Println("ERROR || ip route add", network, "via", gateway, "metric", metric, " || err:", err, " || rawout: ", string(out))
		return err
	}
	return
}

func (i *Interface) DelRoute(network string, gateway string, metric string) (err error) {
	out, err := exec.Command("ip", "route", "delete", network, "via", gateway, "metric", metric).Output()
	if err != nil {
		log.Println("ERROR || ip route delete", network, "via", gateway, "metric", metric, " || err:", err, " || rawout: ", string(out))
		return err
	}
	return
}

type syscallCreateIF struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

type Config struct {
	Name       string
	Address    string
	User       uint
	Group      uint
	Multiqueue bool
	Persistent bool
	TunnelFile string
}

func NewTunnel(C *Config) (IF *Interface, err error) {
	IF = new(Interface)

	if C.TunnelFile == "" {
		C.TunnelFile = "/dev/net/tun"
	}

	IF.Address = C.Address

	fd, err := syscall.Open(C.TunnelFile, os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}

	// fdUintPtr := uintptr(fd)
	IF.FD = uintptr(fd)

	var flags uint16 = 0x1000
	flags |= 0x0001
	if C.Multiqueue {
		flags |= 0x0100 // MULTIQUEUE FLAG
	}

	var req syscallCreateIF
	req.Flags = flags
	copy(req.Name[:], []byte(C.Name))

	if err = ioctl(IF.FD, syscall.TUNSETIFF, uintptr(unsafe.Pointer(&req))); err != nil {
		return nil, err
	}
	IF.Name = strings.Trim(string(req.Name[:]), "\x00")

	if IF.User != 0 {
		if err = ioctl(IF.FD, syscall.TUNSETOWNER, uintptr(IF.User)); err != nil {
			return nil, err
		}
	}

	if IF.Group != 0 {
		if err = ioctl(IF.FD, syscall.TUNSETGROUP, uintptr(IF.Group)); err != nil {
			return nil, err
		}
	}

	if C.Persistent {
		if err = ioctl(IF.FD, syscall.TUNSETPERSIST, uintptr(1)); err != nil {
			return nil, err
		}
	}

	IF.RWC = os.NewFile(IF.FD, "tun_"+C.Name)
	return
}

func ioctl(fd uintptr, request uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}
