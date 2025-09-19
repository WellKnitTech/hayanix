package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/wellknittech/hayanix/internal/cli"
)

var version = "0.1.0"

func main() {
	var cliArgs cli.Args
	ctx := kong.Parse(&cliArgs,
		kong.Name("hayanix"),
		kong.Description("Sigma-based threat hunting and fast forensics timeline generator for *nix logs"),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

	// Run the specific command
	if err := ctx.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
