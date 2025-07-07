//nolint:gosec
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apple/pkl-go/pkl"
)

func main() {
	evaluator, err := pkl.NewEvaluator(context.Background(), pkl.PreconfiguredOptions)
	if err != nil {
		panic(err)
	}
	configFiles, err := filepath.Glob("configs/*.pkl")
	if err != nil {
		panic(err)
	}
	for i := range configFiles {
		filename := configFiles[i]
		bytes, err := evaluator.EvaluateExpressionRaw(context.Background(), pkl.FileSource(filename), "module")
		if err != nil {
			panic(err)
		}
		path := "internal/build/" + strings.Replace(filename, ".pkl", ".msgpack", 1)
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			panic(err)
		}
		if err = os.WriteFile(path, bytes, 0o644); err != nil {
			panic(err)
		}
		fmt.Printf("Generated %s", path)
	}
}
