package main

import (
	// "encoding/json"
	"flag"
	"fmt"
	"github.com/moovweb/gokogiri"
	"io/ioutil"
	"net/http"
	"os"
)

// type BetSeries struct {
// 	Player string
// 	Bets   []string
// }
//
// func (this *BetSeries) String() string {
// 	return fmt.Sprintf("Player: %s: %s", this.Player, this.Bets)
// }
//
// func (this *BetSeries) IsInvalidBet() bool {
// 	var badOutcomeCount = 0
// 	for _, outcome := range this.Bets {
// 		if outcome == "2:1" || outcome == "1:2" {
// 			badOutcomeCount++
// 		}
// 	}
// 	return badOutcomeCount > 4
// }
//
// type RuleCheckResult struct {
// 	InvalidBets []*BetSeries
//
// 	CheckedURL  string
// 	CheckedXPath string
//
// }
//
// func (this RuleCheckResult) MarshalJSON() ([]byte, error) {
// 	result := map[string]interface{}{}
// 	result["checked_url"] = this.CheckedURL
//
// 	if len(this.InvalidBets) > 0 {
// 		result["result"] = fmt.Sprintf("%d Spieler haben geträumt.", len(this.InvalidBets))
// 		result["players"] = this.InvalidBets
// 	} else {
// 		result["result"] = "Alle haben fein ordentlich getippt."
// 		result["players"] = nil
// 	}
//
// 	return json.Marshal(result)
// }
//
// func GetCheckResultForUrl(url string) RuleCheckResult {
// 	var doc *goquery.Document
// 	var e error
//
// 	if doc, e = goquery.NewDocument(url); e != nil {
// 		panic(e.Error())
// 	}
//
// 	invalidBets := []*BetSeries{}
// 	doc.Find("table.nw.kicktipp-tabs tr").Each(func(i int, s *goquery.Selection) {
// 		var betSeries = &BetSeries{}
// 		betSeries.Player = s.Find("td.mg_class").Text()
//
// 		betSeries.Bets = []string{}
// 		s.Find("td.t").Each(func(i int, selection *goquery.Selection) {
// 			outcomeWithTrailingGarbage := selection.Text()
// 			var outcome string = outcomeWithTrailingGarbage[0 : len(outcomeWithTrailingGarbage)-1]
// 			betSeries.Bets = append(betSeries.Bets, outcome)
// 		})
// 		s.Find("td.f").Each(func(i int, selection *goquery.Selection) {
// 			var outcome string = selection.Text()
// 			betSeries.Bets = append(betSeries.Bets, outcome)
// 		})
//
// 		fmt.Println(betSeries)
//
// 		if betSeries.IsInvalidBet() {
// 			invalidBets = append(invalidBets, betSeries)
// 		}
// 		//		fmt.Println()
// 	})
//
// 	//	b, _ := json.Marshal(invalidBets)
// 	//	fmt.Println(string(b))
// 	result := RuleCheckResult{}
// 	result.InvalidBets = invalidBets
// 	result.CheckedURL = url
//
// 	return result
// }

const DefaultUrl = "http://trakkor.better-idea.org/_status"
const DefaultXpath = "//rails_version"

func extractXpathFromUrl(xpath string, url string) string {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
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
		fmt.Printf("Xpath not found: %s", xpath)
		return ""
	}

	result := n[0].Content()
	return result
}

func handler(writer http.ResponseWriter, req *http.Request) {
	fmt.Printf("url.query: %s\n", req.URL.Query())

	// result = GetCheckResultForUrl(url)
	// b, _ = json.MarshalIndent(result, "", "  ")
	// fmt.Println(string(b))

	values := req.URL.Query()
	url := values.Get("url")
	xpath := values.Get("xpath")
	content := extractXpathFromUrl(xpath, url)

	b := []byte(fmt.Sprintf(`{"url": "%s", "xpath": "%s", "content": "%s"}`, url, xpath, content))
	writer.Header().Add("Content-Type", "application/json")
	writer.Write(b)
}

func main() {
	url := flag.String("url", DefaultUrl, "help message for url")
	xpath := flag.String("xpath", DefaultXpath, "help message for xpath")
	flag.Parse()

	content := extractXpathFromUrl(*xpath, *url)
	fmt.Printf("EXTRACTED: `%s`", content)

	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
