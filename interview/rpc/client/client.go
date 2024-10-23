package main

import (
	"fmt"
	"net/rpc"
)

type Person struct {
	Name string
}

func main() {
	p := &Person{Name: "hxia043"}
	client, err := rpc.DialHTTP("tcp4", "127.0.0.1:6445")
	if err != nil {
		fmt.Println(err)
	}

	reply := new(string)
	if err := client.Call("Mock.SayHello", p, reply); err != nil {
		fmt.Println(err)
	}

	fmt.Println(*reply)
}
