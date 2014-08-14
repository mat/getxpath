package main

import "testing"

func TestBasic(t *testing.T) {
	xpath := "//title"
	uri := "http://example.com"
	expected := "Example Domain"
	runTest(t, expected, uri, xpath)
}

func TestFollowRedirect(t *testing.T) {
	xpath := "//title"
	uri := "http://google.com"
	expected := "Google"
	runTest(t, expected, uri, xpath)
}

func TestHttps(t *testing.T) {
	xpath := "//title"
	uri := "https://news.ycombinator.com"
	expected := "Hacker News"
	runTest(t, expected, uri, xpath)
}

func TestUmlautsInContent(t *testing.T) {
	xpath := "//title"
	uri := "http://www.better-idea.org"
	expected := "matthias lüdtke"
	runTest(t, expected, uri, xpath)
}

func TestUmlautsOnAmazonDe(t *testing.T) {
	xpath := "//title"
	uri := "http://www.amazon.de"
	expected := "Amazon.de: Günstige Preise für Elektronik & Foto, Filme, Musik, Bücher, Games, Spielzeug & mehr"
	runTest(t, expected, uri, xpath)
}

func TestErrorForNonExistentHost(t *testing.T) {
	xpath := "//title"
	uri := "http://www.does-not-exist-domain.de"
	expectError(t, uri, xpath)
}

func TestErrorForNonExistentXpathElement(t *testing.T) {
	xpath := "//does_not_exist"
	uri := "http://www.better-idea.org"
	expectError(t, uri, xpath)
}

func expectError(t *testing.T, uri string, xpath string) {
	actual, e := ExtractXpathFromUrl(xpath, uri)
	if e == nil {
		t.Errorf("Did expect an eror but got none.")
	}
	if actual != "" {
		t.Errorf("Expected empty result, but got ExtractXpathFromUrl(%v, %v) = '%v'", xpath, uri, actual)
	}
}

func runTest(t *testing.T, expected string, uri string, xpath string) {
	actual, e := ExtractXpathFromUrl(xpath, uri)
	if e != nil {
		t.Errorf("Did not expect an eror but got: %v", e)
	}
	if actual != expected {
		t.Errorf("Got ExtractXpathFromUrl(%v, %v) = '%v', wanted '%v'", xpath, uri, actual, expected)
	}
}
