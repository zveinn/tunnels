package main

import (
	"fmt"
	"time"

	"github.com/zveinn/tunnels"
)

func main() {
	IF := &tunnels.Interface{
		Name:       "nvpn-new",
		Address:    "10.6.6.10",
		NetMask:    "255.255.255.0",
		TxQueuelen: 3000,
		MTU:        1501,
		Persistent: true,
	}

	// err := IF.Create()
	// if err != nil {
	// 	fmt.Println("CREATE FAIL")
	// 	panic(err)
	// }

	fmt.Println(IF.Name)
	fmt.Println(IF.Address)
	fmt.Println(IF.FD)
	fmt.Println(IF.Group)
	fmt.Println(IF.User)
	fmt.Println(IF.RWC)

	// err = IF.Syscall_Addr()
	// if err != nil {
	// 	fmt.Println("ADDR FAIL")
	// 	panic(err)
	// }
	// err = IF.Syscall_NetMask()
	// if err != nil {
	// 	fmt.Println("NETMASK FAIL")
	// 	panic(err)
	// }
	//
	// err = IF.Syscall_MTU()
	// if err != nil {
	// 	fmt.Println("MTU FAIL")
	// 	panic(err)
	// }
	//
	// err = IF.Syscall_TXQueuelen()
	// if err != nil {
	// 	fmt.Println("TXQUEUE FAIL")
	// 	panic(err)
	// }
	//
	// err = IF.Syscall_UP()
	// if err != nil {
	// 	fmt.Println("UP FAIL")
	// 	panic(err)
	// }
	//
	// err = IF.Syscall_Addr()
	// if err != nil {
	// 	fmt.Println("ADDR FAIL")
	// 	panic(err)
	// }

	err := IF.Syscall_AddRoute("9.9.9.9", IF.Address, "255.255.255.255", IF.Name)
	if err != nil {
		fmt.Println("ROUTE FAIL")
		panic(err)
	}

	// err := IF.Syscall_Delete()
	// if err != nil {
	// 	fmt.Println("ADDR FAIL")
	// 	panic(err)
	// }
	for {
		time.Sleep(1 * time.Second)
	}
	// err = IF.Syscall_MTU()
}
