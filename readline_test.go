package cli

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadLine(t *testing.T) {
	defer func() { stdin = os.Stdin }()
	ctx := context.Background()

	stdin = strings.NewReader("This is a test line\n")
	line := ReadLine(ctx)
	assert.Equal(t, "This is a test line", line)
}

func TestReadLine_BlockContext(t *testing.T) {
	defer func() { stdin = os.Stdin }()
	ctx, cancel := context.WithCancel(context.Background())

	sync := make(chan bool)
	r := blockingReader{input: make(chan string, 1), omitNewLine: true}
	stdin = r

	r.input <- "wait until the user hits enter"

	var line string
	go func() {
		<-sync
		line = ReadLine(ctx)
		<-sync
	}()

	sync <- true // start the goroutine

	select {
	case sync <- true:
		t.Fatal("ReadLine should block until the first \\n but it returned")
	case <-time.After(5 * time.Millisecond):
		// ok good, time to move on
	}

	cancel()
	select {
	case sync <- true:
		assert.Empty(t, line)
	case <-time.After(5 * time.Millisecond):
		t.Fatal("ReadLine should return if the given context is canceled")
	}
}

func TestReadLines(t *testing.T) {
	defer func() { stdin = os.Stdin }()
	ctx := context.Background()
	cases := map[string][]string{
		"empty input": nil,
		"one line": {
			"line 1\n",
		},
		"three lines": {
			"line 1\n",
			"line 2\n",
			"line 3\n",
		},
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			t.Log(name)
			stdin = strings.NewReader(strings.Join(input, ""))
			linesChan := ReadLines(ctx)
			lines := extract(linesChan)
			for i, expected := range input {
				expected = expected[:len(expected)-1] // expect string without trailing newline
				assert.Equal(t, expected, lines[i])
			}
		})
	}
}

func TestReadLinesCancel(t *testing.T) {
	defer func() { stdin = os.Stdin }()
	ctx, cancel := context.WithCancel(context.Background())

	r := blockingReader{input: make(chan string, 1)}
	r.input <- "line1"
	stdin = r

	linesChan := ReadLines(ctx)
	<-linesChan
	cancel()

	select {
	case _, ok := <-linesChan:
		assert.False(t, ok, "channel should have been closed when context is canceled")
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout: seems like the channel was not closed")
	}
}

func extract(c <-chan string) []string {
	result := make(chan []string)
	go func() {
		var lines []string
		for s := range c {
			lines = append(lines, s)
		}
		result <- lines
	}()

	select {
	case r := <-result:
		return r
	case <-time.After(time.Second):
		panic("timeout")
	}
}

type blockingReader struct {
	input       chan string
	omitNewLine bool
}

func (r blockingReader) Read(p []byte) (int, error) {
	s := <-r.input
	if !r.omitNewLine {
		s = s + "\n"
	}

	return strings.NewReader(s).Read(p)
}
