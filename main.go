package main

import (
	"fmt"

	"github.com/riversteve/goplay/api"
)

func main() {
	for i := 0; i < 10; i++ {
		apiKey, err := api.GenerateAPIKey()
		if err != nil {
			fmt.Println("Error generating API key:", err)
			return
		}

		fmt.Println("Generated API key:", apiKey)
	}
}
