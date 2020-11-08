package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"os"
	"strings"
)

type ConsoleInputReader struct {
	outputChan      chan string
	reader          *readline.Instance
	done            chan struct{}
	cancelableStdin *readline.CancelableStdin
}

// Move cursor to the first character of the line and erase the entire line
func (c *ConsoleInputReader) Clean() {
	fmt.Print("\r\u001b[2K\u001b[3D")
}

// Re-print the unfinished line to STDOUT with prompt.
func (c *ConsoleInputReader) Refresh() {
	c.reader.Refresh()
}

func (c *ConsoleInputReader) Close() {
	// hacked around for a bug in chzyer/readline
	// it was tried to fix by https://github.com/chzyer/readline/issues/52 but failed
	// because the Close method of the inner cancelable Stdin was not called by (*Instance).Close()
	// so we should call it by ourselves
	if err := c.cancelableStdin.Close(); err != nil {
		wsdogLogger.Debugf("close cancelable stdin failed: %s", err.Error())
	}

	if err := c.reader.Close(); err != nil {
		wsdogLogger.Debugf("close input reader failed: %s", err.Error())
	}
}

func NewConsoleInputReader() *ConsoleInputReader {
	cancelableStdin := readline.NewCancelableStdin(os.Stdin)
	reader, err := readline.NewEx(&readline.Config{Prompt: "> ",
		Stdin: cancelableStdin,
	})

	if err != nil {
		wsdogLogger.Fatalf("setup read from console failed: %s", err)
	}

	outputChan := make(chan string)
	done := make(chan struct{})
	r := ConsoleInputReader{outputChan, reader, done, cancelableStdin}

	go func() {
		defer close(outputChan)
		defer close(done)
		for {
			text, err := reader.Readline()
			if err != nil {
				wsdogLogger.Debugf("receive error when read from console %s", err.Error())
				return
			}

			outputChan <- strings.TrimSuffix(text, "\n")
		}
	}()

	return &r
}
