package main

import (
	"fmt"

	"github.com/zveinn/tunnels"
)

func main() {
	IF := &tunnels.Interface{
		Name:        "nvpn-new",
		IPv4Address: "10.6.6.10",
		IPv6Address: "fe80::1",
		NetMask:     "255.255.255.0",
		TxQueuelen:  3000,
		MTU:         1501,
		Persistent:  true,
		// User:       1000,
		// Group:      1000,
	}

	fmt.Println(IF.Name)
	fmt.Println(IF.IPv4Address)
	fmt.Println(IF.FD)
	fmt.Println(IF.Group)
	fmt.Println(IF.User)
	fmt.Println(IF.RWC)
	// time.Sleep(2 * time.Second)
	// err := IF.Syscall_Delete()
	// if err != nil {
	// 	fmt.Println("ADDR FAIL")
	// 	panic(err)
	// }
	tunnels.FindGateway()

	err := IF.Create()
	if err != nil {
		fmt.Println("CREATE FAIL")
		panic(err)
	}

	err = IF.Syscall_Addr()
	if err != nil {
		fmt.Println("ADDR FAIL")
		panic(err)
	}

	// err = IF.Syscall_Addrv6()
	// if err != nil {
	// 	fmt.Println("ADDRv6 FAIL")
	// 	panic(err)
	// }

	err = IF.Syscall_NetMask()
	if err != nil {
		fmt.Println("NETMASK FAIL")
		panic(err)
	}

	err = IF.Syscall_MTU()
	if err != nil {
		fmt.Println("MTU FAIL")
		panic(err)
	}

	err = IF.Syscall_TXQueuelen()
	if err != nil {
		fmt.Println("TXQUEUE FAIL")
		panic(err)
	}

	err = IF.Syscall_UP()
	if err != nil {
		fmt.Println("UP FAIL")
		panic(err)
	}

	err = tunnels.IP_AddRoute("9.9.9.9", IF.IPv4Address, "0")
	if err != nil {
		fmt.Println("ROUTE FAIL", err)
		panic(1)
	}

	inBuff := make([]byte, 1000)

	for {
		n, err := IF.RWC.Read(inBuff)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(inBuff[:n])
	}

	// err = IF.Syscall_MTU()
}
