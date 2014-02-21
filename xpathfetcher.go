package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/moovweb/gokogiri"
)

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

func TimeFromUnixTimeStampString(str string) time.Time {
	n, _ := strconv.Atoi(str)
	loc, _ := time.LoadLocation("CET")

	return time.Unix(int64(n), 0).In(loc)
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

type Result struct {
	Url        string      `json:"url"`
	Xpath      string      `json:"xpath"`
	Content    string      `json:"content"`
	Error      interface{} `json:"error"`
	Version    string      `json:"version"`
	DeployedAt time.Time   `json:"deployed_at"`
}

func handler(writer http.ResponseWriter, req *http.Request) {
	log.Print(req)
	writer.Header().Add("Content-Type", "application/json")

	url := req.URL.Query().Get("url")
	xpath := req.URL.Query().Get("xpath")

	result := Result{
		Url:        url,
		Xpath:      xpath,
		Version:    os.Getenv("GIT_REVISION"),
		DeployedAt: TimeFromUnixTimeStampString(os.Getenv("DEPLOYED_AT")),
	}
	if len(url) == 0 || len(xpath) == 0 {
		writer.WriteHeader(400)
		result.Error = "Need both url and xpath query parameter."
	} else {
		content, err := ExtractXpathFromUrl(xpath, url)
		result.Content = content
		result.Error = ErrorMessageOrNil(err)
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
	url := flag.String("url", "", "help message for url")
	xpath := flag.String("xpath", "", "help message for xpath")
	flag.Parse()

	return *url, *xpath
}

func runTestUsingCommentLineArgs() {
	url, xpath := parseCommandLineArgs()
	if len(url) > 0 && len(xpath) > 0 {
		content, _ := ExtractXpathFromUrl(xpath, url)
		fmt.Printf("EXTRACTED: `%s`\n", content)
	}
}

func startServer() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		panic("PORT missing")
	}

	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	runTestUsingCommentLineArgs()
	startServer()
}
