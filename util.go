// Collection of small, useful utility functions
package util

import (
	"bufio"
	"flag"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/davidnarayan/go-logging"
)

// TimeoutDialer creates a Dialer that can timeout after a connection timeout
// or a read timeout
func TimeoutDialer(connTimeout, rwTimeout time.Duration) func(netw, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, connTimeout)

		if err != nil {
			return nil, err
		}

		conn.SetDeadline(time.Now().Add(rwTimeout))

		return conn, nil
	}
}

// ExitWhenOrphaned checks if the current process becomes orphaned and
// immediately kills the script. This prevents multiple instances of the
// script running indefinitely when it's being launched by some external
// process (e.g. Splunk).
func ExitWhenOrphaned() {
	tick1s := time.NewTicker(time.Second * 1).C

	// Check for orphaned process
	go func() {
		for {
			select {
			case <-tick1s:
				if os.Getppid() == 1 {
					logging.Fatal("Orphaned process! Exiting!")
				}
			}
		}
	}()
}

// Collect args from files or stdin
func GetArgs() (args []string) {
	var inputs []io.Reader

	// Check for files first and use stdin if no files are specified
	files := flag.Args()

	if len(files) > 0 {
		for _, file := range files {
			input, err := os.Open(file)

			if err != nil {
				logging.Warn(err.Error())
				continue
			}

			inputs = append(inputs, input)
		}
	} else {
		inputs = append(inputs, os.Stdin)
	}

	// Scan all inputs for arguments
	for _, input := range inputs {
		scanner := bufio.NewScanner(input)

		for scanner.Scan() {
			line := scanner.Text()
			// Skip empty lines and comments
			arg := strings.TrimSpace(line)

			if arg == "" || strings.HasPrefix(arg, "#") {
				continue
			}

			args = append(args, arg)
			logging.Trace("Adding arg: %s", arg)
		}

		if err := scanner.Err(); err != nil {
			logging.Error("Unable to add line: %s", err)
		}
	}

	return args
}
