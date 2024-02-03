//go:build freebsd || linux || openbsd

package tunnels

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

type syscallAddAddrV4 struct {
	Name [16]byte
	syscall.RawSockaddrInet4
}
type syscallAddAddrV6 struct {
	Name [16]byte
	syscall.RawSockaddrInet6
}

type syscallChangeMTU struct {
	Name [16]byte
	MTU  int32
}

type syscallChangeTXQueueLen struct {
	Name       [16]byte
	TxQueueLen int32
}

type syscallSetFlags struct {
	Name  [16]byte
	Flags int16
}

// TODO
// 1. Find default gateway
// 2. Find Default DNS ( only on windows )

type Interface struct {
	Name        string
	IPv4Address string
	IPv6Address string
	NetMask     string
	TxQueuelen  int32
	MTU         int32
	User        uint
	Group       uint
	Multiqueue  bool
	Persistent  bool
	TunnelFile  string

	RWC io.ReadWriteCloser
	FD  uintptr
}

func (IF *Interface) Syscall_TXQueuelen() (err error) {
	var ifr syscallChangeTXQueueLen
	copy(ifr.Name[:], []byte(IF.Name))
	ifr.TxQueueLen = IF.TxQueuelen

	if err = socketCtl(
		syscall.SIOCSIFTXQLEN,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

func (IF *Interface) Syscall_MTU() (err error) {
	var ifr syscallChangeMTU
	copy(ifr.Name[:], []byte(IF.Name))
	ifr.MTU = IF.MTU

	if err = socketCtl(
		syscall.SIOCSIFMTU,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

func (IF *Interface) Syscall_NetMask() (err error) {
	var ifr syscallAddAddrV4
	ifr.Port = 0
	ifr.Family = syscall.AF_INET

	copy(ifr.Name[:], []byte(IF.Name))
	copy(ifr.Addr[:], net.ParseIP(IF.NetMask).To4())

	if err = socketCtl(
		syscall.SIOCSIFNETMASK,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

// func (IF *Interface) Syscall_Addrv6() (err error) {
// 	iface, err := net.InterfaceByName(IF.Name)
// 	if err != nil {
// 		return err
// 	}
//
// 	var ifr syscallAddAddrV6
// 	ifr.Port = 0
// 	ifr.Family = syscall.AF_INET6
// 	ifr.Flowinfo = uint32(0)
// 	ifr.Scope_id = uint32(iface.Index)
//
// 	copy(ifr.Name[:], []byte(IF.Name))
// 	copy(ifr.Addr[:], net.ParseIP(IF.IPv6Address).To16())
// 	fmt.Println(ifr)
//
// 	if err = socketCtlv6(
// 		syscall.SIOCSIFADDR,
// 		uintptr(unsafe.Pointer(&ifr)),
// 	); err != nil {
// 		return
// 	}
//
// 	return
// }

// func (IF *Interface) IPv6_ADDR() (err error) {
// 	out, err := exec.Command(
// 		"ip",
// 		"-6",
// 		"addr",
// 		"add",
// 		IF.IPv6Address,
// 		"dev",
// 		IF.Name,
// 	).CombinedOutput()
// 	if err != nil {
// 		return errors.New(err.Error() + "---" + string(out))
// 	}
// 	return
// }

func (IF *Interface) Syscall_Addr() (err error) {
	var ifr syscallAddAddrV4
	ifr.Port = 0
	ifr.Family = syscall.AF_INET

	copy(ifr.Name[:], []byte(IF.Name))
	copy(ifr.Addr[:], net.ParseIP(IF.IPv4Address).To4())

	if err = socketCtl(
		syscall.SIOCSIFADDR,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

func (IF *Interface) Syscall_DOWN() (err error) {
	var ifr syscallSetFlags

	copy(ifr.Name[:], []byte(IF.Name))
	ifr.Flags |= 0x0

	if err = socketCtl(
		syscall.SIOCSIFFLAGS,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

func (IF *Interface) Syscall_Delete() (err error) {
	var ifr syscallSetFlags
	DOR := 1 << 17

	copy(ifr.Name[:], []byte(IF.Name))
	ifr.Flags |= 0x0
	ifr.Flags = int16(DOR)

	if err = socketCtl(
		syscall.SIOCSIFFLAGS,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	_ = exec.Command("ip", "link", "delete", IF.Name).Run()

	return
}

func (IF *Interface) Syscall_UP() (err error) {
	var ifr syscallSetFlags

	copy(ifr.Name[:], []byte(IF.Name))
	ifr.Flags |= 0x1

	if err = socketCtl(
		syscall.SIOCSIFFLAGS,
		uintptr(unsafe.Pointer(&ifr)),
	); err != nil {
		return
	}

	return
}

//
// func (IF *Interface) IP_ADDR(ip string) (err error) {
// 	out, err := exec.Command(
// 		"ip",
// 		"addr",
// 		"add",
// 		ip,
// 		"dev",
// 		IF.Name,
// 	).Output()
// 	if err != nil {
// 		log.Println("ERROR || ip addr add", ip, "dev", IF.Name, " || err:", err, " || rawout: ", string(out))
// 		return err
// 	}
// 	return
// }
//
// func (IF *Interface) IP_DOWN() (err error) {
// 	out, err := exec.Command(
// 		"ip",
// 		"link",
// 		"set",
// 		"dev",
// 		IF.Name,
// 		"down",
// 	).Output()
// 	if err != nil {
// 		log.Println("ERROR || ip link set dev", IF.Name, "down || err:", err, " || rawout: ", string(out))
// 		return err
// 	}
// 	return
// }
//
// func (IF *Interface) IP_UP() (err error) {
// 	out, err := exec.Command("ip", "link", "set", "dev", IF.Name, "up").Output()
// 	if err != nil {
// 		log.Println("ERROR || ip link set dev", IF.Name, "up || err:", err, " || rawout: ", string(out))
// 		return err
// 	}
// 	return
// }
//
// func (i *Interface) IP_TXQueueLen(ql string) (err error) {
// 	out, err := exec.Command("ip", "link", "set", i.Name, "txqueuelen", ql).Output()
// 	if err != nil {
// 		log.Println("ERROR || ip link set", i.Name, "txqueuelen", ql, " || err:", err, " || rawout: ", string(out))
// 		return err
// 	}
// 	return
// }
//
// func (IF *Interface) IP_MTU() (err error) {
// 	out, err := exec.Command("ip", "link", "set", IF.Name, "mtu", fmt.Sprint(IF.MTU)).Output()
// 	if err != nil {
// 		log.Println("ERROR || ip link set", IF.Name, "mtu", IF.MTU, " || err:", err, " || rawout: ", string(out))
// 		return err
// 	}
// 	return
// }

type syscallCreateIF struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func (IF *Interface) Create() (err error) {
	if IF.TunnelFile == "" {
		IF.TunnelFile = "/dev/net/tun"
		// IF.TunnelFile = "/dev/net/tun"
	}

	fd, err := syscall.Open(IF.TunnelFile, os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}

	IF.FD = uintptr(fd)

	var flags uint16 = 0x1000
	flags |= 0x0001
	if IF.Multiqueue {
		flags |= 0x0100 // MULTIQUEUE FLAG
	}

	var req syscallCreateIF
	req.Flags = flags
	copy(req.Name[:], []byte(IF.Name))

	if err = tunnelCtl(IF.FD, syscall.TUNSETIFF, uintptr(unsafe.Pointer(&req))); err != nil {
		return err
	}

	if IF.User != 0 {
		if err = tunnelCtl(IF.FD, syscall.TUNSETOWNER, uintptr(IF.User)); err != nil {
			return err
		}
	}

	if IF.Group != 0 {
		if err = tunnelCtl(IF.FD, syscall.TUNSETGROUP, uintptr(IF.Group)); err != nil {
			return err
		}
	}

	if IF.Persistent {
		if err = tunnelCtl(IF.FD, syscall.TUNSETPERSIST, uintptr(1)); err != nil {
			return err
		}
	}

	IF.RWC = os.NewFile(IF.FD, "tun_"+IF.Name)
	return
}

func socketCtlv6(request uintptr, argp uintptr) error {
	fd, err := syscall.Socket(
		syscall.AF_INET6,
		syscall.SOCK_DGRAM,
		syscall.IPPROTO_IP,
	)
	defer syscall.Close(fd)
	if err != nil {
		return err
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

func socketCtl(request uintptr, argp uintptr) error {
	fd, err := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_DGRAM,
		syscall.IPPROTO_IP,
	)
	defer syscall.Close(fd)
	if err != nil {
		return err
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

func tunnelCtl(fd uintptr, request uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

func IP_AddRoute(network string, gateway string, metric string) (err error) {
	_ = IP_DelRoute(network, gateway, metric)
	out, err := exec.Command("ip", "route", "add", network, "via", gateway, "metric", metric).CombinedOutput()
	if err != nil {
		return errors.New("err :" + err.Error() + " || out: " + string(out))
	}
	return
}

func IP_DelRoute(network string, gateway string, metric string) (err error) {
	out, err := exec.Command("ip", "route", "delete", network, "via", gateway, "metric", metric).CombinedOutput()
	if err != nil {
		return errors.New("err :" + err.Error() + " || out: " + string(out))
	}
	return
}

type syscallAddRoute struct {
	Name [16]byte

	rt_dst     syscall.RawSockaddrInet4 // Destination address
	rt_gateway syscall.RawSockaddrInet4 // Gateway address
	rt_genmask syscall.RawSockaddrInet4 // Netmask

	rt_flags  uint16
	Metric    int32
	RefCount  int
	Use       int
	Priority  int
	Device    [16]byte // Interface name
	MTU       int
	Window    int
	IRTT      int
	Reserved1 int
	Reserved2 int
}

func (IF *Interface) Syscall_AddRoute_INPROGRESS(destination, gateway, netmask, name string) error {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(sock)

	var dst, gw, mask syscall.RawSockaddrInet4
	var route syscallAddRoute
	// copy(route.Device[:], []byte(name))

	copy(dst.Addr[:], net.ParseIP(destination).To4())
	dst.Family = syscall.AF_INET
	// dst.Port = 0

	copy(gw.Addr[:], net.ParseIP(gateway).To4())
	gw.Family = syscall.AF_INET
	// gw.Port = 0

	copy(mask.Addr[:], net.ParseIP(netmask).To4())
	mask.Family = syscall.AF_INET
	// mask.Port = 0

	route.rt_dst = dst
	route.rt_gateway = gw
	route.rt_genmask = mask
	// route.Metric = 0
	route.rt_flags = syscall.RTF_GATEWAY | syscall.RTF_UP
	// route.Device = [16]byte{}
	// copy(route.Device[:], ifaceName)

	fmt.Println("ROUTE:", route)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(sock), syscall.SIOCADDRT, uintptr(unsafe.Pointer(&route)))
	if errno != 0 {
		return fmt.Errorf("Failed to add route: %v", errno)
	}

	return nil
}
