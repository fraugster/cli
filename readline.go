package cli

import (
	"bufio"
	"context"
	"io"
	"os"
)

// stdin is the io.Reader that lines are read from.
// This is variable so we can mock it in tests.
// This is not injected into ReadLines for pragmatic convenience reasons.
var stdin io.Reader = os.Stdin

// ReadLine reads a single line from stdin and returns it without the trailing
// newline. This function blocks until the first newline is read or the context
// is canceled. In the later case the empty string is returned.
func ReadLine(ctx context.Context) string {
	r := bufio.NewReader(stdin)

	input := make(chan string)
	go func() {
		line, err := r.ReadString('\n')
		if err != nil || len(line) == 0 {
			input <- ""
			return
		}
		input <- line[:len(line)-1]
	}()

	select {
	case <-ctx.Done():
		return ""
	case s := <-input:
		return s
	}
}

// ReadLines reads lines from stdin and returns them in a channel.
// All strings in the returned channel will not include the trailing newline.
// The channel is closed automatically if there are no more lines or if the
// context is closed.
//
// This function panics if there was any error other than io.EOF when reading
// from os.Stdin.
func ReadLines(ctx context.Context) <-chan string {
	r := bufio.NewReader(stdin)
	c := make(chan string)
	go func() {
		for {
			line, err := r.ReadString('\n')
			switch {
			case err == io.EOF:
				close(c)
				return
			case err != nil:
				panic(err)
			}

			c <- line[:len(line)-1]
		}
	}()

	lines := make(chan string)
	go func() {
		for {
			select {
			case l, ok := <-c:
				if !ok {
					close(lines)
					return
				}
				lines <- l
			case <-ctx.Done():
				close(lines)
				return
			}
		}
	}()

	return lines
}
