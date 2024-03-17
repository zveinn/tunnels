//go:build windows

package tunnels

import (
	"fmt"
	"log"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modwintun               = newLazyDLL("./wintun.dll", setupLogger)
	procWintunCreateAdapter = modwintun.NewProc("WintunCreateAdapter")
	procWintunOpenAdapter   = modwintun.NewProc("WintunOpenAdapter")
	procWintunCloseAdapter  = modwintun.NewProc("WintunCloseAdapter")
	procWintunDeleteDriver  = modwintun.NewProc("WintunDeleteDriver")
	// procWintunGetAdapterLUID          = modwintun.NewProc("WintunGetAdapterLUID")
	procWintunGetRunningDriverVersion = modwintun.NewProc("WintunGetRunningDriverVersion")
	procWintunAllocateSendPacket      = modwintun.NewProc("WintunAllocateSendPacket")
	procWintunEndSession              = modwintun.NewProc("WintunEndSession")
	// procWintunGetReadWaitEvent        = modwintun.NewProc("WintunGetReadWaitEvent")
	procWintunReceivePacket        = modwintun.NewProc("WintunReceivePacket")
	procWintunReleaseReceivePacket = modwintun.NewProc("WintunReleaseReceivePacket")
	procWintunSendPacket           = modwintun.NewProc("WintunSendPacket")
	procWintunStartSession         = modwintun.NewProc("WintunStartSession")
	procWintunGetLastError         = modwintun.NewProc("WintunGetLastError")
	// procWintunSetLogger               = modwintun.NewProc("WintunSetLogger")

	GUID *windows.GUID

	// WINDOWS DLL
	IPHLPApi = syscall.NewLazyDLL("iphlpapi.dll")
	GetTCP   = IPHLPApi.NewProc("GetExtendedTcpTable")
	GetUDP   = IPHLPApi.NewProc("GetExtendedUdpTable")
	SetTCP   = IPHLPApi.NewProc("SetTcpEntry")
)

type loggerLevel int

const (
	PacketSizeMax   = 0xffff    // Maximum packet size
	RingCapacityMin = 0x20000   // Minimum ring capacity (128 kiB)
	RingCapacityMax = 0x4000000 // Maximum ring capacity (64 MiB)
	AdapterNameMax  = 128

	// WINDOWS DLL

	MIB_TCP_TABLE_OWNER_PID_ALL = 5
	MIB_TCP_STATE_DELETE_TCB    = 12
)

// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// LOGGING
// LOGGING
// LOGGING
// LOGGING
// LOGGING
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
type TimestampedWriter interface {
	WriteWithTimestamp(p []byte, ts int64) (n int, err error)
}

func logMessage(level loggerLevel, timestamp uint64, msg *uint16) int {
	log.Println(windows.UTF16PtrToString(msg))
	return 0
}

func setupLogger(dll *lazyDLL) {
	var callback uintptr
	// log.Println("SETTING UP", runtime.GOARCH)

	if runtime.GOARCH == "386" {
		callback = windows.NewCallback(func(level loggerLevel, timestampLow, timestampHigh uint32, msg *uint16) int {
			return logMessage(level, uint64(timestampHigh)<<32|uint64(timestampLow), msg)
		})
	} else if runtime.GOARCH == "arm" {
		callback = windows.NewCallback(func(level loggerLevel, _, timestampLow, timestampHigh uint32, msg *uint16) int {
			return logMessage(level, uint64(timestampHigh)<<32|uint64(timestampLow), msg)
		})
	} else if runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" {
		callback = windows.NewCallback(logMessage)
	} else {
		callback = windows.NewCallback(logMessage)
	}

	syscall.SyscallN(dll.NewProc("WintunSetLogger").Addr(), callback)
}

// RunningVersion returns the version of the loaded driver.
func RunningVersion() (version uint32, err error) {
	r0, _, e1 := syscall.SyscallN(procWintunGetRunningDriverVersion.Addr())
	version = uint32(r0)
	if version == 0 {
		err = e1
	}
	return
}

// Version returns the version of the Wintun DLL.
func Version() string {
	if modwintun.Load() != nil {
		return "unknown"
	}
	resInfo, err := windows.FindResource(modwintun.module, windows.ResourceID(1), windows.RT_VERSION)
	if err != nil {
		return "unknown"
	}
	data, err := windows.LoadResourceData(modwintun.module, resInfo)
	if err != nil {
		return "unknown"
	}

	var fixedInfo *windows.VS_FIXEDFILEINFO
	fixedInfoLen := uint32(unsafe.Sizeof(*fixedInfo))
	err = windows.VerQueryValue(unsafe.Pointer(&data[0]), `\`, unsafe.Pointer(&fixedInfo), &fixedInfoLen)
	if err != nil {
		return "unknown"
	}
	version := fmt.Sprintf("%d.%d", (fixedInfo.FileVersionMS>>16)&0xff, (fixedInfo.FileVersionMS>>0)&0xff)
	if nextNibble := (fixedInfo.FileVersionLS >> 16) & 0xff; nextNibble != 0 {
		version += fmt.Sprintf(".%d", nextNibble)
	}
	if nextNibble := (fixedInfo.FileVersionLS >> 0) & 0xff; nextNibble != 0 {
		version += fmt.Sprintf(".%d", nextNibble)
	}
	return version
}

// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// INTERFACE
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
// =========================================
func (IF *Interface) CreateOrOpen() (err error) {
	IF.NamePtr, err = windows.UTF16PtrFromString(IF.Name)
	if err != nil {
		return
	}
	IF.UNamePtr = uintptr(unsafe.Pointer(IF.NamePtr))

	IF.TypePtr, err = windows.UTF16PtrFromString("Niceland")
	if err != nil {
		return
	}
	IF.UTypePtr = uintptr(unsafe.Pointer(IF.TypePtr))

	// TRY GENERATING A STATIC UID
	// A.GUID = new(windows.GUID)

	// https://github.com/microsoft/go-winio/blob/main/pkg/guid/guid.go
	// https://github.com/WireGuard/wintun/pull/7

	// https://github.com/WireGuard/wintun/blob/master/README.md#wintuncreateadapter
	IF.UGUIDPtr = uintptr(unsafe.Pointer(&IF.GUID))

	fmt.Println("opening adapter", IF.Name)
	var msg error

	IF.Handle, _, msg = syscall.SyscallN(
		procWintunOpenAdapter.Addr(),
		IF.UNamePtr,
	)

	fmt.Println("open adapter: "+IF.Name+" msg:", msg)

	if IF.Handle == 0 {
		fmt.Println("creating adapter:", IF.Name)
		fmt.Println("creating adapter:", IF.UNamePtr, IF.UTypePtr, IF.UGUIDPtr)
		IF.Handle, _, msg = syscall.SyscallN(
			procWintunCreateAdapter.Addr(),
			IF.UNamePtr,
			IF.UTypePtr,
			IF.UGUIDPtr,
		)
		fmt.Println("create adapter: "+IF.Name+" msg:", msg)
		if IF.Handle == 0 {
			err = msg
			fmt.Println("error creating adapter:", IF.Handle, err)
			return
		}

	}

	fmt.Println("adapter created: ", IF.Name)
	// runtime.SetFinalizer(IF.Handle, AdapterCleanup)
	return
}

func (IF *Interface) ReceivePacket() (packet []byte, size uint16, err error) {
	// var packetSize uint32
	r0, _, msg := syscall.SyscallN(
		procWintunReceivePacket.Addr(),
		IF.SessionHandle,
		uintptr(unsafe.Pointer(&size)),
	)

	if r0 == 0 {
		err = msg
		return
	}

	packet = unsafe.Slice((*byte)(unsafe.Pointer(r0)), size)
	return
}

func (IF *Interface) ReleaseReceivePacket(packet []byte) (err error) {
	r0, _, msg := syscall.SyscallN(
		procWintunReleaseReceivePacket.Addr(),
		IF.SessionHandle,
		uintptr(unsafe.Pointer(&packet[0])),
	)
	if r0 == 0 {
		err = msg
		return
	}

	return
}

func (IF *Interface) AllocateSendPacket(packetSize int) (packet []byte, err error) {
	r0, _, msg := syscall.SyscallN(
		procWintunAllocateSendPacket.Addr(),
		IF.SessionHandle,
		uintptr(packetSize),
	)
	if r0 == 0 {
		err = msg
		return
	}
	packet = unsafe.Slice((*byte)(unsafe.Pointer(r0)), packetSize)
	return
}

func (IF *Interface) SendPacket(packet []byte) (err error) {
	_, _, _ = syscall.SyscallN(
		procWintunSendPacket.Addr(),
		IF.SessionHandle,
		uintptr(unsafe.Pointer(&packet[0])),
	)
	// if r0 == 0 {
	// 	err = msg
	// 	return
	// }
	return
}
