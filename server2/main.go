package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sync/semaphore"
)

const (
	availableMemoryParameter   = "1"
	freeMemoryPercentParameter = "2"
)

type connCount struct {
	cnt int
	m   sync.Mutex
}

func (c *connCount) Increase() {
	c.m.Lock()
	defer c.m.Unlock()
	c.cnt++
}

func (c *connCount) GetCnt() int {
	c.m.Lock()
	defer c.m.Unlock()
	return c.cnt
}

var (
	maxConnnections int64 = 20
	cnt             connCount
)

func main() {

	ch := make(chan string, 1)
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sem := semaphore.NewWeighted(maxConnnections)

	pipe, err := os.OpenFile("/tmp/log_pipe", os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		logger.Error("os.OpenFile failed", "err", err)
		return
	}
	defer pipe.Close()

	logger = slog.New(NewPipeHandler(os.Stdout, ch))
	go func(ch chan string) {
		for {
			select {
			case <-ctx.Done():
				if len(ch) != 0 {
					line := <-ch
					io.WriteString(pipe, line)
				}
				return
			default:
				line := <-ch
				if _, err := io.WriteString(pipe, line); err != nil {
					logger.Error("write failed", "err", err)
					return
				}
			}
		}
	}(ch)

	l, err := net.Listen("tcp", ":5002")
	if err != nil {
		logger.Error("error while starting listening", "err", err)
	}

	logger.Info("Server up")
	go func() {
		for {
			if err := sem.Acquire(ctx, 1); err != nil {
				logger.Error("failed to acquire semaphore", slog.String("err", err.Error()))
				break
			}

			conn, err := l.Accept()
			if err != nil {
				logger.Error("error while accepting", slog.String("err", err.Error()))
				conn.Close()
				continue
			}
			conn.SetDeadline(time.Now().Add(10 * time.Minute))
			go serve(conn)
			logger.Info("start serving")
		}
	}()
	<-quitChan
	logger.Info("Server down")
	cancel()
	time.Sleep(5 * time.Second)
}

// serve entry, check simple or long option
func serve(conn net.Conn) {
	cnt.Increase()
	sendMessage(conn, "choose option: (1 - simple, 2 - long)")
	buf1 := make([]byte, 2)
	if n, err := conn.Read(buf1); err != nil || n == 0 {
		return
	}

	sendMessage(conn, "choose parameter: (1 - available memory in Mb, 2 - free memory percent)")
	buf2 := make([]byte, 2)
	n, err := conn.Read(buf2)
	if err != nil || n == 0 {
		return
	}

	if buf1[0] == 49 {
		serveSimple(conn, string(buf2[:n]))
	} else if buf1[0] == 50 {
		serveLong(conn, string(buf2[:n]))
	} else {
		sendMessage(conn, "Not valid option")
	}
	logger.Info("Stop serving")
}

func serveSimple(conn net.Conn, parameter string) {
	defer conn.Close()
	availableMemory, freeMemoryPercent, err := getMemoryInfo()
	if err != nil {
		sendMessage(conn, fmt.Sprintf("error occured on server - %s", err.Error()))
	} else {
		sendMessage(conn, getInfo(parameter, availableMemory, freeMemoryPercent))
	}
}

func serveLong(conn net.Conn, parameter string) {
	defer conn.Close()

	var (
		currentAvailableMemory   uint64
		currentFreeMemoryPercent float64
	)

	for {
		availableMemory, freeMemoryPercent, err := getMemoryInfo()
		if err != nil {
			sendMessage(conn, fmt.Sprintf("error occured while getting info - %s", err.Error()))
			break
		}
		if (parameter == availableMemoryParameter && availableMemory != currentAvailableMemory) ||
			(parameter == freeMemoryPercentParameter && freeMemoryPercent != currentFreeMemoryPercent) {
			currentAvailableMemory = availableMemory
			currentFreeMemoryPercent = freeMemoryPercent
			err := sendMessage(conn, getInfo(parameter, availableMemory, freeMemoryPercent))
			if err != nil {
				logger.Info("Stop serving - client disconnected")
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func getInfo(parameter string, avaliable uint64, percents float64) string {
	var res string
	switch parameter {
	case availableMemoryParameter:
		res = fmt.Sprintf("available memory in Mb - %d", avaliable)
	case freeMemoryPercentParameter:
		res = fmt.Sprintf("free memory percents - %.2f", percents)
	default:
		res = "not valid parameter"
	}
	return res
}

func getMemoryInfo() (uint64, float64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Error("Ошибка при получении информации о памяти", slog.String("err", err.Error()))
		return 0, 0, err
	}

	// Объем доступной физической памяти в Mb
	availableMemory := v.Available / (1048576.0)
	// Процент свободной памяти
	freeMemoryPercent := 100.0 - v.UsedPercent
	return availableMemory, freeMemoryPercent, nil
}

func sendMessage(conn net.Conn, message string) error {
	finalm := fmt.Sprintf("time=%s, message=%s", time.Now().Format(time.RFC822), message)
	if n, err := conn.Write([]byte(finalm)); err != nil {
		return err
	} else if n == 0 {
		return fmt.Errorf("nothing writed")
	}
	return nil
}
