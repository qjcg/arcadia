package main

import (
	"context"

	"github.com/warpstreamlabs/bento/public/service"

	// Import all standard Benthos components
	_ "github.com/warpstreamlabs/bento/public/components/all"

	// Add your plugin packages here
	_ "github.com/qjcg/arcadia/examples/bento/99-a-benthos-plugin"
)

func main() {
	service.RunCLI(context.Background())
}
