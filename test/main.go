package main

import (
	"fmt"
	"time"

	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/zveinn/tunnels"
	"golang.org/x/sys/windows"
)

func main() {
	GUID, err := guid.NewV4()
	if err != nil {
		fmt.Println("unable to create windows GUID", err)
		return
	}

	IF := &tunnels.Interface{
		Name:          "nvpn",
		IPv4Address:   "10.4.4.4",
		IPv6Address:   "fe80::1",
		NetMask:       "255.255.255.0",
		TxQueuelen:    3000,
		MTU:           1501,
		Persistent:    true,
		GUID:          windows.GUID(GUID),
		RingCap:       0x4000000,
		RetransmitMS:  "500",
		Gateway:       "10.4.4.4",
		GatewayMetric: "1000",
		// User:       1000,
		// Group:      1000,
	}

	// fmt.Println(IF.ame)
	// fmt.Println(IF.IPv4Address)
	// fmt.Println(IF.FD)
	// fmt.Println(IF.Group)
	// fmt.Println(IF.User)
	// fmt.Println(IF.RWC)

	err = IF.CreateOrOpen()
	if err != nil {
		fmt.Println("CREATE OUT:", err)
	}
	// err = IF.Syscall_Addr()
	// if err != nil {
	// 	fmt.Println("UP OUT:", err)
	// }
	err = IF.Syscall_UP()
	if err != nil {
		fmt.Println("UP OUT:", err)
	}
	for {
		time.Sleep(20 * time.Millisecond)
		packet, size, _ := IF.ReceivePacket()
		if size == 0 {
			continue
		}
		fmt.Println(size, len(packet), err)
		fmt.Println(packet)
		fmt.Printf("%p\n", &packet)
		fmt.Printf("%p\n", &packet[0])
		fmt.Println(packet[9])
	}
	// copy(packetNew, packet)

	// IF2 := &tunnels.Interface{
	// 	Name:        "nvpn",
	// 	IPv4Address: "10.6.6.10",
	// 	IPv6Address: "fe80::1",
	// 	NetMask:     "255.255.255.0",
	// 	TxQueuelen:  3000,
	// 	MTU:         1501,
	// 	Persistent:  true,
	// 	GUID:        windows.GUID(GUID),
	// 	// User:       1000,
	// 	// Group:      1000,
	// }
	// err = IF2.CreateOrOpen()

	// if err != nil {
	// 	fmt.Println("ERR OUT:", err)
	// 	// fmt.Println("CREATE FAIL")
	// 	// panic(err)
	// }

	// err = IF.Syscall_Addr()
	// if err != nil {
	// 	fmt.Println("ADDR FAIL")
	// 	panic(err)
	// }

	// err = IF.Syscall_Addrv6()
	// if err != nil {
	// 	fmt.Println("ADDRv6 FAIL")
	// 	panic(err)
	// }

	// err = IF.Syscall_NetMask()
	// if err != nil {
	// 	fmt.Println("NETMASK FAIL")
	// 	panic(err)
	// }

	// err = IF.Syscall_MTU()
	// if err != nil {
	// 	fmt.Println("MTU FAIL")
	// 	panic(err)
	// }

	// err = IF.Syscall_TXQueuelen()
	// if err != nil {
	// 	fmt.Println("TXQUEUE FAIL")
	// 	panic(err)
	// }

	// err = IF.Syscall_UP()
	// if err != nil {
	// 	fmt.Println("UP FAIL")
	// 	panic(err)
	// }

	// err = tunnels.IP_AddRoute("9.9.9.9", IF.IPv4Address, "0")
	// if err != nil {
	// 	fmt.Println("ROUTE FAIL", err)
	// 	panic(1)
	// }

	// inBuff := make([]byte, 1000)

	// for {
	// 	n, err := IF.RWC.Read(inBuff)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	fmt.Println(inBuff[:n])
	// }
	for {
		time.Sleep(1 * time.Second)
	}
	// err = IF.Syscall_MTU()
}
