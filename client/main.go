//go:build docker
// +build docker

package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

var (
	server1Addr string
	server2Addr string
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Получаем адреса серверов из переменных окружения
	server1Addr = getEnv("SERVER1_ADDR", "localhost:5001")
	server2Addr = getEnv("SERVER2_ADDR", "localhost:5002")

	clientEntryPoint()
}

func clientEntryPoint() {
	var addr string

	// Получаем выбор сервера из переменной окружения
	serverChoice := getEnv("SERVER_CHOICE", "")
	if serverChoice == "" {
		fmt.Println("SERVER_CHOICE environment variable not set")
		return
	}

	if serverChoice == "1" {
		addr = server1Addr
	} else if serverChoice == "2" {
		addr = server2Addr
	} else {
		fmt.Printf("No such server - %s", serverChoice)
		return
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error while connecting:", err)
		return
	}
	defer conn.Close()

	if serverChoice == "1" {
		firstServer(conn)
	} else if serverChoice == "2" {
		secondServer(conn)
	}
}

func firstServer(conn net.Conn) {
	// Получаем цвет из переменной окружения
	color := getEnv("TEXT_COLOR", "")
	if color == "" {
		fmt.Println("TEXT_COLOR environment variable not set")
		return
	}

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

	// Получаем тип параметра из переменной окружения
	paramType := getEnv("PARAM_TYPE", "")
	if paramType == "" {
		fmt.Println("PARAM_TYPE environment variable not set")
		return
	}

	buf := make([]byte, 2)
	switch paramType {
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

	// Получаем параметр из переменной окружения
	parameter := getEnv("PARAMETER", "")
	if parameter == "" {
		fmt.Println("PARAMETER environment variable not set")
		return
	}

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

	for {
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
