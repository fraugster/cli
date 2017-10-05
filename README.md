## github.com/fraugster/cli

[![Build Status](https://secure.travis-ci.org/fraugster/cli.png?branch=master)](http://travis-ci.org/fraugster/cli)
[![GoDoc](https://godoc.org/github.com/fraugster/cli?status.svg)](https://godoc.org/github.com/fraugster/cli)
[![License](https://img.shields.io/github/license/fraugster/cli.svg)](https://github.com/fraugster/cli/blob/master/LICENSE)

Package `cli` implements utility functions for running command line applications.

The following simple features are currently implemented:
- creating a `context.Context` which is closed when `SIGINT`, `SIGQUIT` or `SIGTERM` is received.
- context aware reading lines from stdin into a channel
- printing values using user a defined format (e.g. `json`, `yml` or `table`)

## Motivation

CLI applications often perform similar tasks such as handling stop signals, reading
from stdin and printing results to stdout. This package provides these functions
in a single and coherent repository. Applications built with `github.com/fraugster/cli`
treat `context.Context` as first class citizen to remain responsive and implement
graceful shutdown. Naturally this is visible when consuming (multi-line) input
from stdin.  The `cli.Print` function is especially useful since it encourages
developers to support multiple machine readable output formats which makes it easy
to pipe the results of one application into another (e.g. `my-app -o=json | jq …`).

## Installation 

```sh
$ go get github.com/fraugster/cli
```

## Usage

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/fraugster/cli"
)

func main() {
	// Create a context that is done when SIGINT, SIGQUIT or SIGTERM is received
	ctx := cli.Context()

	// Let the user decide what output format she prefers.
	format := flag.String("output", "json", "Output format. One of json|yaml|table|raw")
	flag.Parse()

	// Reading a single line from stdin (returns "" if context is canceled).
	fmt.Print("Please insert your name: ")
	name := cli.ReadLine(ctx)
	fmt.Println("Hello", name)

	// Continuously read lines from stdin and return them in a channel.
	fmt.Println("You have 10 seconds to talk to me")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var inputs []string
	for line := range cli.ReadLines(ctx) {
		inputs = append(inputs, line)
		fmt.Println("OK, go on")
	}
	fmt.Println("Time is up")
    
	// Print something to stdout in a format specified by the user.
	fmt.Println("Your inputs were:")
	cli.Print(*format, inputs)
}
```

## Dependencies

- `gopkg.in/yaml.v2` for YAML output
- `github.com/stretchr/testify` to run unit tests

### License

This package is licensed under the the BSD 3-Clause License.
Please see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome but must be discussed first in a new github issue before
Pull Requests will be accepted.
