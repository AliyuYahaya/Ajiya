// Command ajiya keeps a project's phases and tickets in plain markdown.
package main

import (
	"os"

	"github.com/AliyuYahaya/ajiya/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
