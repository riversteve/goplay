package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: web-title.exe [url]")
		os.Exit(1)
	}

	url := os.Args[1]
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	resp, err := http.Get(url)
	handleError(err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	handleError(err)

	doc, err := html.Parse(bytes.NewReader(body))
	handleError(err)

	title := getTitle(doc)
	fmt.Println(title)
}

func handleError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func getTitle(doc *html.Node) string {
	if doc.Type == html.ElementNode && doc.Data == "title" {
		return doc.FirstChild.Data
	}

	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		title := getTitle(c)

		if title != "" {
			return title
		}
	}

	return ""
}
