package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"code.google.com/p/go.net/html/charset"

	"github.com/moovweb/gokogiri"
)

var logger = log.New(os.Stdout, "getxpath: ", log.LstdFlags|log.Lmicroseconds)

func ReadBodyFromUrl(url string) ([]byte, error) {
	resp, e := http.Get(url)
	for retries := 1; e != nil && retries <= 3; retries += 1 {
		logger.Printf("Retrying to fetch %s (%d)\n", url, retries)
		time.Sleep(time.Duration(retries) * time.Second)
		resp, e = http.Get(url)
	}
	if e != nil {
		logger.Printf("Fetching %s failed too many times.", url)
		return nil, e
	}

	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
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
	bodyBytes, e := ReadBodyFromUrl(url)
	if e != nil {
		return "", e
	}
	status.BytesProcessed += int64(len(bodyBytes))

	utf8bytes, e := convertToUtf8(bodyBytes)
	if e != nil {
		return "", e
	}

	doc, e := gokogiri.ParseHtml(utf8bytes)
	if e != nil {
		return "", e
	}
	if doc == nil {
		return "", errors.New(fmt.Sprintf("Could not ParseHtml"))
	}
	defer doc.Free()

	root := doc.Root()
	if root == nil {
		return "", errors.New(fmt.Sprintf("Could not ParseHtml: Doc has no root"))
	}

	nodes, e := root.Search(xpath)
	if e != nil {
		return "", e
	}
	if len(nodes) < 1 {
		return "", errors.New(fmt.Sprintf("Xpath not found"))
	}

	res := nodes[0].Content()
	return res, nil
}

func convertToUtf8(bytez []byte) ([]byte, error) {
	reader := bytes.NewReader(bytez)
	utf8reader, e := charset.NewReader(reader, "")
	utf8bytes, e := ioutil.ReadAll(utf8reader)
	return utf8bytes, e
}

type Query struct {
	Url   string `json:"url"`
	Xpath string `json:"xpath"`
}

type Result struct {
	Query  interface{} `json:"query"`
	Result string      `json:"result"`
	Error  interface{} `json:"error"`
}

var status = &Status{}

type Status struct {
	Version      string
	GoVersion    string
	DeployedAt   time.Time
	FirstRequest time.Time

	OkCount        int64
	ErrorCount     int64
	LastOk         time.Time
	LastError      time.Time
	BytesProcessed int64
}

func requestHandler(writer http.ResponseWriter, req *http.Request) {
	if (status.FirstRequest == time.Time{}) {
		status.FirstRequest = time.Now()
	}
	logger.Print(req)
	writer.Header().Add("Content-Type", "application/json; charset=utf-8")

	q := Query{
		Url:   req.URL.Query().Get("url"),
		Xpath: req.URL.Query().Get("xpath"),
	}

	res := Result{
		Query: q,
	}
	if len(q.Url) == 0 || len(q.Xpath) == 0 {
		writer.WriteHeader(400)
		res.Error = "Need both url and xpath query parameter."
	} else {
		content, e := ExtractXpathFromUrl(q.Xpath, q.Url)
		res.Result = content
		res.Error = ErrorMessageOrNil(e)
		if e != nil {
			logger.Printf("ERROR: Could not get xpath for query %v because: %v", q, e)
		}
	}

	if res.Error != nil {
		status.LastError = time.Now()
		status.ErrorCount += 1
	} else {
		status.LastOk = time.Now()
		status.OkCount += 1
	}

	bytes, e := json.MarshalIndent(res, "", "  ")
	if e != nil {
		panic(e)
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
	logger.Print(req)
	writer.Header().Add("Content-Type", "application/json")

	bytes, e := json.MarshalIndent(status, "", "  ")
	if e != nil {
		logger.Panic(e)
	}
	writer.Write(bytes)
}

func startServer() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		panic("PORT missing")
	}

	http.HandleFunc("/_status", statusHandler)
	http.HandleFunc("/get", requestHandler)

	e := http.ListenAndServe(":"+port, nil)
	if e != nil {
		panic(e)
	}
}

func main() {
	runTestUsingCommentLineArgs()
	startServer()
}

func init() {
	status = &Status{
		Version:    os.Getenv("GIT_REVISION"),
		GoVersion:  runtime.Version(),
		DeployedAt: TimeFromUnixTimeStampString(os.Getenv("DEPLOYED_AT")),
	}
}
