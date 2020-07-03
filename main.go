package main

import (
	"github.com/fuskovic/screen-recorder/cmd"
	"go.coder.com/cli"
)

func main() {
	cli.RunRoot(&cmd.Root{})
}
