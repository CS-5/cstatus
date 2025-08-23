package main

import (
	"fmt"
	"io"
	"os"

	"github.com/CS-5/statusline/builder"
	"github.com/CS-5/statusline/render"
)

func main() {
	// Read JSON from stdin
	jsonData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	ctx, err := render.NewRenderContext(string(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	builder := builder.New()
	fmt.Print(builder.Build(ctx))
}
