package main

import _ "expvar"

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/moovweb/gokogiri"
	"golang.org/x/net/html/charset"
)

var logger = log.New(os.Stdout, "getxpath: ", log.LstdFlags|log.Lmicroseconds)

func readBodyFromURL(url string) ([]byte, string, error) {
	resp, e := http.Get(url)
	for retries := 1; e != nil && retries <= 3; retries++ {
		logger.Printf("Retrying to fetch %s (%d)\n", url, retries)
		time.Sleep(time.Duration(retries) * time.Second)
		resp, e = http.Get(url)
	}
	if e != nil {
		logger.Printf("Fetching %s failed too many times.", url)
		return nil, "", e
	}

	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, "", e
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	return bytes, contentType, nil
}

func timeFromUnixTimeStampString(str string) time.Time {
	n, _ := strconv.Atoi(str)
	loc, _ := time.LoadLocation("CET")

	return time.Unix(int64(n), 0).In(loc)
}

func extractXpathFromURL(url string, xpath string) (string, error) {
	bodyBytes, contentType, e := readBodyFromURL(url)
	if e != nil {
		return "", e
	}
	status.BytesProcessed += int64(len(bodyBytes))

	utf8bytes, e := convertToUtf8(bodyBytes, contentType)
	if e != nil {
		return "", e
	}

	doc, e := gokogiri.ParseHtml(utf8bytes)
	if e != nil {
		return "", e
	}
	if doc == nil {
		return "", fmt.Errorf("Could not ParseHtml")
	}
	defer doc.Free()

	root := doc.Root()
	if root == nil {
		return "", fmt.Errorf("Could not ParseHtml: Doc has no root")
	}

	nodes, e := root.Search(xpath)
	if e != nil {
		return "", e
	}
	if len(nodes) < 1 {
		return "", fmt.Errorf("Xpath not found")
	}

	res := nodes[0].Content()
	return res, nil
}

func convertToUtf8(bytez []byte, contentType string) ([]byte, error) {
	reader := bytes.NewReader(bytez)
	utf8reader, e := charset.NewReader(reader, contentType)
	if e != nil {
		return nil, e
	}
	utf8bytes, e := ioutil.ReadAll(utf8reader)
	return utf8bytes, e
}

type query struct {
	URL   string `json:"url"`
	Xpath string `json:"xpath"`
}

type result struct {
	Query  interface{} `json:"query"`
	Result string      `json:"result"`
	Error  interface{} `json:"error"`
}

var status = &statusData{}

type statusData struct {
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

	q := query{
		URL:   req.FormValue("url"),
		Xpath: req.FormValue("xpath"),
	}

	res := result{
		Query: q,
	}
	if len(q.URL) > 0 && len(q.Xpath) > 0 {
		content, e := extractXpathFromURL(q.URL, q.Xpath)
		res.Result = content
		res.Error = errorMessageOrNil(e)
		if e != nil {
			logger.Printf("ERROR: Could not get xpath for query %v because: %v", q, e)
		}
	} else {
		writer.WriteHeader(400)
		res.Error = "Need both url and xpath query parameter."
	}

	if res.Error != nil {
		status.LastError = time.Now()
		status.ErrorCount++
	} else {
		status.LastOk = time.Now()
		status.OkCount++
	}

	bytes, e := json.Marshal(res)
	if e != nil {
		panic(e)
	}
	writer.Write(bytes)
}

func errorMessageOrNil(e error) interface{} {
	if e != nil {
		return e.Error()
	}
	return nil
}

func parseCommandLineArgs() (string, string, int) {
	url := flag.String("url", "", "URL to fetch")
	xpath := flag.String("xpath", "", "XPath to extract from the document at <url>")
	port := flag.Int("port", 0, "Port in server mode")
	flag.Parse()

	return *url, *xpath, *port
}

func runTestUsingCommentLineArgs(url string, xpath string) {
	if len(url) > 0 && len(xpath) > 0 {
		content, _ := extractXpathFromURL(url, xpath)
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

func startServer(port int) {
	http.HandleFunc("/_status", statusHandler)
	http.HandleFunc("/get", requestHandler)

	e := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if e != nil {
		panic(e)
	}
}

func main() {
	url, xpath, port := parseCommandLineArgs()

	if port > 0 {
		startServer(port)
	} else if url != "" && xpath != "" {
		runTestUsingCommentLineArgs(url, xpath)
	} else {
		flag.PrintDefaults()
	}

}

func init() {
	status = &statusData{
		Version:    os.Getenv("GIT_REVISION"),
		GoVersion:  runtime.Version(),
		DeployedAt: timeFromUnixTimeStampString(os.Getenv("DEPLOYED_AT")),
	}
}
