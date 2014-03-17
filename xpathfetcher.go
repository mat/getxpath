package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/moovweb/gokogiri"
)

var log = logrus.New()

func ReadBodyFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	for retries := 1; err != nil && retries <= 3; retries += 1 {
		log.Printf("Retrying to fetch %s (%d)\n", url, retries)
		time.Sleep(time.Duration(retries) * time.Second)
		resp, err = http.Get(url)
	}
	if err != nil {
		log.Warn("Fetching %s failed too many times.", url)
		return nil, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return bytes, nil
}

func TimeFromUnixTimeStampString(str string) time.Time {
	n, _ := strconv.Atoi(str)
	loc, _ := time.LoadLocation("CET")

	return time.Unix(int64(n), 0).In(loc)
}

func ExtractXpathFromUrl(xpath string, url string) (string, error) {
	bytes, err := ReadBodyFromUrl(url)
	if err != nil {
		return "", err
	}
	status.BytesProcessed += int64(len(bytes))

	doc, err := gokogiri.ParseHtml(bytes)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", errors.New(fmt.Sprintf("Could not ParseHtml"))
	}
	defer doc.Free()

	root := doc.Root()
	if root == nil {
		return "", errors.New(fmt.Sprintf("Could not ParseHtml: Doc has no root"))
	}

	nodes, err := root.Search(xpath)
	if err != nil {
		return "", err
	}
	if len(nodes) < 1 {
		return "", errors.New(fmt.Sprintf("Xpath not found"))
	}

	result := nodes[0].Content()
	return result, nil
}

type Result struct {
	Url     string      `json:"url"`
	Xpath   string      `json:"xpath"`
	Content string      `json:"content"`
	Error   interface{} `json:"error"`
}

var status = &Status{}

type Status struct {
	Version    string
	DeployedAt time.Time

	OkCount        int64
	ErrorCount     int64
	LastOk         time.Time
	LastError      time.Time
	BytesProcessed int64
}

func requestHandler(writer http.ResponseWriter, req *http.Request) {
	log.Print(req)
	writer.Header().Add("Content-Type", "application/json")

	url := req.URL.Query().Get("url")
	xpath := req.URL.Query().Get("xpath")

	result := Result{
		Url:   url,
		Xpath: xpath,
	}
	if len(url) == 0 || len(xpath) == 0 {
		writer.WriteHeader(400)
		result.Error = "Need both url and xpath query parameter."
	} else {
		content, err := ExtractXpathFromUrl(xpath, url)
		result.Content = content
		result.Error = ErrorMessageOrNil(err)
		if err != nil {
			log.WithFields(logrus.Fields{"url": url, "xpath": xpath}).Error(ErrorMessageOrNil(err))
		}
	}

	if result.Error != nil {
		status.LastError = time.Now()
		status.ErrorCount += 1
	} else {
		status.LastOk = time.Now()
		status.OkCount += 1
	}

	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	writer.Write(bytes)
}

func ErrorMessageOrNil(e error) interface{} {
	if e != nil {
		return e.Error()
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

func statusHandler(writer http.ResponseWriter, req *http.Request) {
	log.Print(req)
	writer.Header().Add("Content-Type", "application/json")

	bytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	writer.Write(bytes)
}

func startServer() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		panic("PORT missing")
	}

	http.HandleFunc("/_status", statusHandler)
	http.HandleFunc("/", requestHandler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	runTestUsingCommentLineArgs()
	startServer()
}

func init() {
	status = &Status{
		Version:    os.Getenv("GIT_REVISION"),
		DeployedAt: TimeFromUnixTimeStampString(os.Getenv("DEPLOYED_AT")),
	}
}
