package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// server reads from pipe

	logChan := make(chan []byte, 200)

	quiteChan := make(chan os.Signal, 1)
	signal.Notify(quiteChan, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pipeRead(ctx, "/tmp/log_pipe", logChan)
	go logFileWriter(ctx, "/tmp/logs.txt", logChan)

	<-quiteChan
	cancel()
	close(logChan)
	fmt.Println("Log server is down")
}

func pipeRead(ctx context.Context, pipePath string, logChan chan<- []byte) {
	pipe, err := os.OpenFile(pipePath, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		fmt.Printf("os.OpenFile on path %s failed - %s", pipePath, err.Error())
		return
	}
	defer pipe.Close()

	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := pipe.Read(buf)
			if err != nil {
				if err == syscall.EAGAIN {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				fmt.Printf("Read error: %v", err)
				break
			}
			if n > 0 {
				data := append([]byte{}, buf[:n]...)

				fmt.Print(string(data))
				logChan <- data
			}
		}
	}
}

func logFileWriter(ctx context.Context, filePath string, logChan <-chan []byte) {
	logFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("os.OpenFile on path %s failed - %s", filePath, err.Error())
	}
	defer logFile.Close()

	var syncCount int
	for {
		select {
		case <-ctx.Done():
			for len(logChan) > 0 {
				data := <-logChan
				logFile.Write(data)
			}
			logFile.Sync()
			return
		case log, ok := <-logChan:
			if !ok {
				return
			}
			if _, err := logFile.Write(log); err != nil {
				fmt.Printf("error while opening - %s", err)
			}
			syncCount++
			if syncCount%10 == 0 {
				logFile.Sync()
			}
		}
	}
}
