//go:build windows

package tunnels

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

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
	Gateway     string

	RWC io.ReadWriteCloser
	FD  uintptr

	// WINDOWS SPECIFIC
	GUID          windows.GUID
	NamePtr       *uint16
	TypePtr       *uint16
	UNamePtr      uintptr
	UTypePtr      uintptr
	UGUIDPtr      uintptr
	Handle        uintptr
	SessionHandle uintptr
	RingCap       uint32
	GatewayMetric string
	RetransmitMS  string
	IFIndex       int

	// ReceiveBuffer []byte
	// SendBuffer    []byte
	// TunHandle windows.InvalidHandle
}

func (IF *Interface) Syscall_NetMask() (err error) {
	return fmt.Errorf("netmask changes are not implemented on windows")
}

func (IF *Interface) Syscall_TXQueuelen() (err error) {
	return fmt.Errorf("txqueuelen changes are not implemented on windows")
}

func (IF *Interface) Syscall_UP() (err error) {
	var msg error
	IF.SessionHandle, _, msg = syscall.SyscallN(
		procWintunStartSession.Addr(),
		IF.Handle,
		uintptr(IF.RingCap))
	if IF.SessionHandle == 0 {
		err = msg
		return
	}

	return nil
}

// type AdapterHandle struct {
// 	Handle uintptr
// }

//	func AdapterCleanup(AH *AdapterHandle) {
//		// syscall.SyscallN(procWintunCloseAdapter.Addr(), 1, AH.handle, 0, 0)
//	}
func (IF *Interface) Syscall_StopReader() (err error) {
	// AH := new(AdapterHandle)
	// AH.Handle = IF.Handle
	// runtime.SetFinalizer(AH, AdapterCleanup)

	r1, _, msg := syscall.SyscallN(
		procWintunEndSession.Addr(),
		IF.SessionHandle)
	if r1 == 0 {
		err = msg
	}

	return
}

func (IF *Interface) Syscall_DOWN() (err error) {
	cmd := exec.Command(
		"netsh",
		"interface",
		"ipv4",
		"delete",
		"address",
		`name="`+IF.Name+`"`,
		"addr=",
		IF.IPv4Address,
		"gateway=",
		"All",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	ob, cerr := cmd.Output()
	if cerr != nil {
		return fmt.Errorf("%s - out: %s", cerr, ob)
	}

	return
}

func (IF *Interface) Syscall_Delete() (err error) {
	r1, _, msg := syscall.SyscallN(
		procWintunCloseAdapter.Addr(),
		IF.Handle)
	if r1 == 0 {
		err = msg
	}
	return
}

// netsh interface ipv4 set interface interface="Ethernet 2" retransmit=1
func (IF *Interface) Syscall_Retransmit() (err error) {
	cmd := exec.Command(
		"netsh",
		"interface",
		"ipv4",
		"set",
		"interface",
		`interface="`+IF.Name+`"`,
		"retransmittime",
		IF.RetransmitMS,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	ob, cerr := cmd.Output()

	if cerr != nil {
		return fmt.Errorf("%s - out: %s ", ob, cerr)
	}
	return
}

// netsh interface ipv4 set interface interface="Ethernet 2" mtu=1500
func (IF *Interface) Syscall_MTU() (err error) {
	cmd := exec.Command(
		"netsh",
		"interface",
		"ipv4",
		"set",
		"interface",
		`interface="`+IF.Name+`"`,
		"mtu",
		string(IF.MTU),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	ob, cerr := cmd.Output()
	if cerr != nil {
		return fmt.Errorf("%s - out: %s ", ob, cerr)
	}
	return
}

func (IF *Interface) Syscall_Addr() (err error) {
	fmt.Println(IF.Name, IF.IPv4Address, IF.Gateway)
	cmd := exec.Command(
		"netsh",
		"interface",
		"ipv4",
		"set",
		"address",
		`name="`+IF.Name+`"`,
		"static",
		IF.IPv4Address,
		IF.NetMask,
		IF.Gateway,
		"gwmetric="+IF.GatewayMetric,
	)

	fmt.Println(
		"netsh",
		"interface",
		"ipv4",
		"set",
		"address",
		`name="`+IF.Name+`"`,
		"static",
		IF.IPv4Address,
		IF.NetMask,
		IF.Gateway,
		"gwmetric="+IF.GatewayMetric,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	ob, cerr := cmd.Output()
	if cerr != nil {
		return fmt.Errorf("%s - out: %s ", ob, cerr)
	}

	return
}

func DNS_Set(IFNameOrIndex, DNSIP, Index string) (err error) {
	cmd := exec.Command(
		"netsh",
		"interface",
		"ipv4",
		"add",
		"dnsservers",
		`name=`+IFNameOrIndex,
		"address="+DNSIP,
		"index="+Index,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	ob, cerr := cmd.Output()
	if cerr != nil {
		return fmt.Errorf("%s - out: %s", cerr, ob)
	}

	return nil
}

func IP_AddRoute(network string, gateway string, mask string, metric string, ifindex string) (err error) {
	if metric == "0" {
		metric = "1"
	}

	_ = IP_DelRoute(network, gateway, metric)

	var cmd *exec.Cmd
	if ifindex == "0" || ifindex == "" {
		cmd = exec.Command(
			"route",
			"add",
			network,
			"mask",
			mask,
			gateway,
			"metric",
			metric,
		)
	} else {
		cmd = exec.Command(
			"route",
			"add",
			network,
			"mask",
			mask,
			gateway,
			"metric",
			metric,
			"IF",
			ifindex,
		)
	}

	fmt.Println(
		"route",
		"add",
		network,
		"mask",
		mask,
		gateway,
		"metric",
		metric,
		"IF",
		ifindex,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	ob, cerr := cmd.Output()

	fmt.Println("ADD OUT:", string(ob), cerr)

	if cerr != nil {
		return fmt.Errorf("%s - out: %s", cerr, ob)
	}

	return
}

func IP_DelRoute(network string, _ string, _ string) (err error) {
	cmd := exec.Command(
		"route",
		"DELETE",
		network,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	ob, cerr := cmd.Output()
	if cerr != nil {
		return fmt.Errorf("%s - out: %s", cerr, ob)
	}

	return
}
