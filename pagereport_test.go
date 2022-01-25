package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
)

const (
	testURL = "https://example.com/test-page"
)

func TestNewPageReport(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	contentType := "text/html"
	statusCode := 200
	body := []byte("<html>")

	headers := http.Header{
		"Content-Type": []string{contentType},
	}

	pageReport := NewPageReport(u, statusCode, &headers, body)

	if pageReport.URL != testURL {
		t.Error("NewPageReport URL != testURL")
	}

	if pageReport.parsedURL != u {
		t.Error("NewPageReport parsedURL != u")
	}

	if pageReport.StatusCode != statusCode {
		t.Error("NewPageReport StatusCode != statusCode")
	}

	if pageReport.ContentType != "text/html" {
		t.Error("NewPageReport ContentType != contentType")
	}

	if string(pageReport.Body) != string(body) {
		t.Error("NewPageReport Body != body")
	}
}

func TestNewRedirectPageReport(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte("<html>")
	statusCode := 301
	redirectURL := "https://example.com/redirect"

	headers := http.Header{
		"Location":     []string{redirectURL},
		"Content-Type": []string{"text/html"},
	}

	pageReport := NewPageReport(u, statusCode, &headers, body)

	if string(pageReport.RedirectURL) != redirectURL {
		t.Error("NewPageReport RedirectURL != redirectURL")
	}

	if pageReport.StatusCode != statusCode {
		t.Error("NewPageReport StatusCode != statusCode")
	}
}

func TestPageReportHTML(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	contentType := "text/html"
	statusCode := 200

	body, err := ioutil.ReadFile("test/test.html")
	if err != nil {
		log.Fatal(err)
	}

	headers := &http.Header{
		"Content-Type": []string{contentType},
	}

	pageReport := NewPageReport(u, statusCode, headers, body)

	if pageReport.Lang != "en" {
		t.Error("Lang != en")
	}

	if pageReport.Title != "Test Page Title" {
		t.Error("Title != Test Page Title")
	}

	if pageReport.Description != "Test Page Description" {
		t.Error("Description != Test Page Description")
	}

	if len(pageReport.Links) != 4 {
		t.Error("len(Links) != 4")
	}

	if len(pageReport.Links) > 0 {
		if pageReport.Links[0].URL != "https://example.com/link1" {
			t.Error("pageReport.Links[0].URL != https://example.com/link1")
		}
		if pageReport.Links[1].URL != "https://example.com/test-page/link2" {
			t.Error("pageReport.Links[1].URL != https://example.com/test-page/link2")
		}
		if pageReport.Links[0].Text != "link1" {
			t.Error("pageReport.Links[0].Text != link1")
		}
		if pageReport.Links[0].Rel != "nofollow" {
			t.Error("pageReport.Links[0].Rel != nofollow")
		}
		if pageReport.Links[0].External != false {
			t.Error("pageReport.Links[0].External != false")
		}
		if pageReport.Links[3].Text != "" {
			t.Error("pageReport.Links[3].Text != \"\"")
		}
	}

	if len(pageReport.ExternalLinks) != 1 {
		t.Error("len(pageReport.ExternalLinks) != 1")
	}

	if pageReport.Refresh != "0;URL='https://example.com/'" {
		t.Error("Refresh != 0;URL='https://example.com/'")
	}

	if pageReport.RedirectURL != "https://example.com/" {
		t.Error("RedirectURL != https://example.com/")
	}

	if pageReport.Robots != "noindex, nofollow" {
		t.Error("Robots != noindex, nofollow")
	}

	if pageReport.Canonical != "http://example.com/canonical/" {
		t.Error("Canonical != http://example.com/canonical/")
	}

	if pageReport.H1 != "H1 Title" {
		t.Error("H1 != H1 Title")
	}

	if pageReport.H2 != "H2 Title" {
		t.Error("H2 != H2 Title")
	}

	if pageReport.Words != 10 {
		t.Error("Words != 10")
	}

	if len(pageReport.Hreflangs) != 1 {
		t.Error("Hreflang != 1")
	}

	if len(pageReport.Hreflangs) == 1 && pageReport.Hreflangs[0].URL != "https://example.com/fr" {
		t.Error("Hreglangs[0].URL != https://example.com/fr")
	}

	if len(pageReport.Hreflangs) == 1 && pageReport.Hreflangs[0].Lang != "fr" {
		t.Error("Hreglangs[0].URL != fr")
	}

	if len(pageReport.Images) != 2 {
		t.Error("Images != 2")
	}

	if pageReport.Images[0].URL != "https://example.com/img/logo.png" {
		t.Error("pageReport.Images[0].URL != https://example.com/img/logo.png")
	}

	if len(pageReport.Scripts) != 1 {
		t.Error("Scripts != 1")
	}

	if len(pageReport.Scripts) == 1 && pageReport.Scripts[0] != "https://example.com/js/app.js" {
		t.Error("Scripts[0] != https://example.com/js/app.js")
	}

	if len(pageReport.Styles) != 1 {
		t.Error("Styles != 1")
	}

	if len(pageReport.Styles) == 1 && pageReport.Styles[0] != "https://example.com/css/style.css" {
		t.Error("Styles[0] != https://example.com/css/style.css")
	}
}
