package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defTimeout             = 10 * time.Second
	checkTimeToChannelDone = 100 * time.Millisecond
)

func main() {
	timeOutFlag := flag.Duration("timeout", defTimeout, "Connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 || len(args) > 2 {
		fmt.Print("Usage: go run main.go [-timeout N] host port\n")
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	connect, err := connectWithTimeOut(address, *timeOutFlag)
	if err != nil {
		fmt.Printf("Error connecting to %s: %v\n", address, err)
		os.Exit(1)
	}
	defer func() {
		if err = connect.Close(); err != nil {
			fmt.Printf("Error: close connect %v", err)
		}
	}()

	fmt.Printf("Connect successfully to %s\n", address)

	done := make(chan struct{})
	var wg sync.WaitGroup

	setupSignal(done)

	wg.Add(2)
	go readFromSocketAndWriteToStdout(connect, done, &wg)
	go readFromStdinAndWriteToSocket(connect, done, &wg)

	wg.Wait()
}

func connectWithTimeOut(address string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func readFromStdinAndWriteToSocket(conn net.Conn, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-done:
			return
		default:
			input, err := reader.ReadString('\n')
			if strings.Contains(input, "\x04") { // CTRL+D
				fmt.Println("Ctrl+D detected. Closing connection.")
				os.Exit(0)
			}
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Printf("Error reading from stdin: %v\n", err)
				close(done)
				return
			}
			_, err = conn.Write([]byte(input))
			if err != nil {
				fmt.Printf("Error writing to socket: %v\n", err)
				close(done)
				return
			}
		}
	}
}

func readFromSocketAndWriteToStdout(conn net.Conn, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)

	for {
		select {
		case <-done:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(checkTimeToChannelDone))
			n, err := reader.Read(buf)
			if err != nil {

				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if err == io.EOF {
					fmt.Println("\nConnection closed by server.")
					close(done)
					return
				}
				fmt.Printf("Error reading from socket: %v\n", err)
				close(done)
				return
			}
			if n > 0 {
				fmt.Print(string(buf[:n]))
			}
		}
	}
}

func setupSignal(done chan struct{}) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range signalChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("\nReceived interrupt signal. Closing connection.")
				close(done)
				return
			}
		}
	}()
}
