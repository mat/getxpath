package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/moovweb/gokogiri"
	"io/ioutil"
	"net/http"
	"os"
)

const DefaultUrl = "http://trakkor.better-idea.org/_status"
const DefaultXpath = "//rails_version"

func extractXpathFromUrl(xpath string, url string) (string, error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// bodyText := string(body)
	// fmt.Printf("Body is: %s", bodyText)

	doc, err := gokogiri.ParseHtml(body)
	defer doc.Free()
	if err != nil {
		panic(err)
	}
	// fmt.Printf("Parsed HTML: %s", doc)
	n, _ := doc.Root().Search(xpath)
	if len(n) < 1 {
		return "", errors.New(fmt.Sprintf("Xpath not found: %s", xpath))
	}

	result := n[0].Content()
	return result, nil
}

func handler(writer http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	url := values.Get("url")
	xpath := values.Get("xpath")
	content, err := extractXpathFromUrl(xpath, url)

	writer.Header().Add("Content-Type", "application/json")

	result := map[string]interface{}{
		"url":     url,
		"xpath":   xpath,
		"content": content,
		"error":   err,
	}

	responseBytes, err := json.MarshalIndent(result, "", "  ")
	writer.Write(responseBytes)
}

func main() {
	url := flag.String("url", DefaultUrl, "help message for url")
	xpath := flag.String("xpath", DefaultXpath, "help message for xpath")
	flag.Parse()

	content, err := extractXpathFromUrl(*xpath, *url)
	fmt.Printf("EXTRACTED: `%s`", content)

	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
