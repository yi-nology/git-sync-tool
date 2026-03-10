//go:build ignore

package main

import (
	"embed"
	"fmt"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	data, err := assets.ReadFile("frontend/dist/index.html")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success! Read %d bytes\n", len(data))
	fmt.Println(string(data[:100]))
}
