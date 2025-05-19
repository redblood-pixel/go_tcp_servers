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

	"github.com/fatih/color"
	"golang.org/x/sync/semaphore"
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
		slog.Error("os.OpenFile failed", "err", err)
		return
	}
	defer pipe.Close()
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
					slog.Error("write failed", "err", err)
					return
				}
			}
		}
	}(ch)
	currentColor = color.New(color.FgWhite)
	logger = slog.New(NewColorHandler(os.Stdout, ch))

	l, err := net.Listen("tcp", ":5001")
	if err != nil {
		logger.Error("error while starting", slog.String("err", err.Error()))
		cancel()
		time.Sleep(1 * time.Second)
		return
	}
	logger.Info("Server up")
	go func() {
		for {
			select {
			case <-ctx.Done():
				if len(ch) != 0 {
					line := <-ch
					io.WriteString(pipe, line)
				}
				return
			default:
				if err := sem.Acquire(ctx, 1); err != nil {
					logger.Error("failed to acquire semaphore")
					break
				}

				conn, err := l.Accept()
				if err != nil {
					logger.Error("error while accepting", slog.String("err", err.Error()))
					if conn != nil {
						conn.Close()
					}
					continue
				}
				conn.SetDeadline(time.Now().Add(10 * time.Minute))
				go serve(conn)
			}
		}
	}()
	<-quitChan
	logger.Info("Server down")
	cancel()
	time.Sleep(5 * time.Second)
}

func serve(conn net.Conn) {
	cnt.Increase()
	defer conn.Close()
	logger.Info("start serving")

	input := make([]byte, 128)

	n, err := conn.Read(input)
	if err != nil {
		logger.Error(
			"error while reading from conn",
			slog.String("err", err.Error()),
			slog.Int("conn", cnt.GetCnt()),
		)
		return
	}
	color := string(input[:n])
	ok := changeColor(color)
	var message string
	if ok {
		logger.Info(
			"color changed",
			slog.String("color", color),
			slog.Int("conn", cnt.GetCnt()),
		)
		message = "color changed"
	} else {
		logger.Info(
			"unsupported color",
			slog.String("color", color),
			slog.Int("conn", cnt.GetCnt()),
		)
		message = "unsupported color"
	}
	err = sendMessage(conn, message)
	if err != nil {
		logger.Error(
			"error while writing to conn",
			slog.String("err", err.Error()),
			slog.Int("conn", cnt.GetCnt()),
		)
		return
	}

	logger.Info("stop serving")
}

func sendMessage(conn net.Conn, message string) error {
	finalm := fmt.Sprintf("time=%s, message=%s", time.Now(), message)
	if n, err := conn.Write([]byte(finalm)); err != nil {
		logger.Error(
			"error while sending message",
			slog.String("err", err.Error()),
			slog.Int("conn", cnt.GetCnt()),
		)
		return err
	} else if n == 0 {
		return fmt.Errorf("nothing writed")
	}
	return nil
}
