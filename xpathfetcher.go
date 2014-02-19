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

func ReadBodyFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func ExtractXpathFromUrl(xpath string, url string) (string, error) {
	body, err := ReadBodyFromUrl(url)
	if err != nil {
		return "", err
	}

	doc, err := gokogiri.ParseHtml(body)
	defer doc.Free()
	if err != nil {
		return "", err
	}
	// fmt.Printf("Parsed HTML: %s", doc)

	nodes, err := doc.Root().Search(xpath)
	if err != nil {
		return "", err
	}
	if len(nodes) < 1 {
		return "", errors.New(fmt.Sprintf("Xpath not found: %s", xpath))
	}

	result := nodes[0].Content()
	return result, nil
}

func handler(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	url := req.URL.Query().Get("url")
	xpath := req.URL.Query().Get("xpath")

	var result map[string]interface{}
	if len(url) == 0 || len(xpath) == 0 {
		writer.WriteHeader(400)
		result = map[string]interface{}{
			"error": "Need both url and xpath query parameter.",
		}
	} else {
		content, err := ExtractXpathFromUrl(xpath, url)
		result = map[string]interface{}{
			"url":     url,
			"xpath":   xpath,
			"content": content,
			"error":   ErrorMessageOrNil(err),
			"version": os.Getenv("GIT_REVISION"),
		}
	}

	responseBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	writer.Write(responseBytes)
}

func ErrorMessageOrNil(err error) interface{} {
	if err != nil {
		return err.Error()
	} else {
		return nil
	}
}

func parseCommandLineArgs() (string, string) {
	url := flag.String("url", DefaultUrl, "help message for url")
	xpath := flag.String("xpath", DefaultXpath, "help message for xpath")
	flag.Parse()

	return *url, *xpath
}

func runTestUsingCommentLineArgs() {
	url, xpath := parseCommandLineArgs()
	content, _ := ExtractXpathFromUrl(xpath, url)
	fmt.Printf("EXTRACTED: `%s`", content)
}

func startServer() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	runTestUsingCommentLineArgs()
	startServer()
}
