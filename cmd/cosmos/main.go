package main

import (
	"fmt"
	"os"

	"github.com/cosmos-toolkit/cosmos-cli/internal/catalog"
	"github.com/cosmos-toolkit/cosmos-cli/internal/cli"
)

func main() {
	subFS, err := getTemplatesFS()
	if err == nil {
		catalog.SetTemplatesFS(subFS)
	}

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
