package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: web-title [url]")
        os.Exit(1)
    }
	
	url := os.Args[1]
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
    resp, err := http.Get(url)

    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    doc, err := html.Parse(bytes.NewReader(body))

    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    title := getTitle(doc)
    fmt.Println(title)
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
