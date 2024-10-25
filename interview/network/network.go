package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

// 发送消息
func Send(conn net.Conn, message string) error {
	data := []byte(message)
	// 首先发送消息长度
	length := uint32(len(data))
	fmt.Println(length)

	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, length)

	// 发送消息头
	if _, err := conn.Write(header); err != nil {
		return err
	}

	// 发送消息体
	_, err := conn.Write(data)
	return err
}

// 接收消息
func Receive(conn net.Conn) (string, error) {
	header := make([]byte, 4)
	// 读取消息头
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", err
	}

	// 解析消息长度
	length := binary.BigEndian.Uint32(header)
	data := make([]byte, length+13)

	// 读取消息体
	if _, err := io.ReadFull(conn, data); err != nil {
		return "", err
	}

	return string(data), nil
}

func main() {
	var wg sync.WaitGroup

	listener, _ := net.Listen("tcp", ":12345")
	defer listener.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, _ := listener.Accept()
		defer conn.Close()

		for {
			message, err := Receive(conn)
			if err != nil {
				fmt.Println("Receive error:", err)
				break
			}
			fmt.Println("Received:", message)
		}
	}()

	conn, _ := net.Dial("tcp", "localhost:12345")
	defer conn.Close()

	Send(conn, "Hello, world!")
	Send(conn, "Another message")

	wg.Wait()
}
