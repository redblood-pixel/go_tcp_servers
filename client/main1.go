//go:build interactive
// +build interactive

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

var (
	in *bufio.Reader
)

func ReadString() string {
	s, err := in.ReadString('\n')
	if err != nil {
		fmt.Println("error while reading string - ", err.Error())
	}
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func main() {

	in = bufio.NewReader(os.Stdin)
	for {
		clientEntryPoint()
	}
}

func clientEntryPoint() {
	var addr string
	fmt.Println("Выберете номер сервера: ")
	// Choose on of servers
	v := ReadString()

	if v == "1" {
		addr = "localhost:5001"
	} else if v == "2" {
		addr = "localhost:5002"
	} else {
		fmt.Printf("No such server -%s", v)
		return
	}

	conn, err := net.Dial("tcp", addr)

	if err != nil {
		fmt.Println("Error while connecting")
		return
	}
	defer conn.Close()

	if v == "1" {
		firstServer(conn)
	} else if v == "2" {
		// Choose parameter type
		secondServer(conn)
	}
}

func firstServer(conn net.Conn) {
	fmt.Println("Choose color of text in server:")
	color := ReadString()
	message := []byte(color)

	if n, err := conn.Write(message); err != nil || n == 0 {
		fmt.Println("Error while sending message")
	}

	input := make([]byte, 128)
	n, err := conn.Read(input)
	if n == 0 || err != nil {
		fmt.Println("Error while reading")
	}
	fmt.Println(string(input[:n]))
}

func secondServer(conn net.Conn) {
	input := make([]byte, 128)
	if n, err := conn.Read(input); n == 0 || err != nil {
		fmt.Println("Error while reading")
	}
	fmt.Println(string(input))
	v := ReadString()
	buf := make([]byte, 2)
	switch v {
	case "1":
		buf = []byte("1")
	case "2":
		buf = []byte("2")
	default:
		fmt.Println("not valid option")
		return
	}
	if n, err := conn.Write(buf); err != nil || n == 0 {
		fmt.Println("Error while writing option")
	}

	if n, err := conn.Read(input); n == 0 || err != nil {
		fmt.Println("Error while reading")
	}
	fmt.Println(string(input))
	parameter := ReadString()
	switch parameter {
	case "1":
		buf = []byte("1")
	case "2":
		buf = []byte("2")
	default:
		fmt.Println("not valid option")
		return
	}

	if n, err := conn.Write(buf); err != nil || n == 0 {
		fmt.Println("Error while writing option")
	}

	buf = make([]byte, 1024)
	ch := make(chan struct{}, 1)

	fmt.Println(v, parameter)
	if v == "2" {
		go func() {
			ReadString()
			ch <- struct{}{}
		}()
	}

loop:
	for {
		select {
		case <-ch:
			break loop
		default:
		}
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Соединение закрыто сервером")
			} else {
				fmt.Printf("Ошибка чтения - %s\n", err.Error())
			}
			break
		}
		fmt.Println(string(buf[:n]))
	}
}
