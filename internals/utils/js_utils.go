package utils

import (
	"log"
	"net/url"
	"regexp"
	"strings"
)

type SourceMap struct {
	Url     string
	JsFiles []string
}

// TODO: check if the website was entered with http:// and if so we'll use that one instead of https.

var goodPrefix = [2]string{"http://", "https://"}

func ExtractHostname(sm *SourceMap) *url.URL {
	hostname, _ := url.Parse(sm.Url)
	//log.Printf("Hostname: %v", hostname.Hostname())
	return hostname
}

func urlCleaner(uri string) []string {

	regexArr := [2]string{"^(http|https)://", "^//"}
	filters := make([]*regexp.Regexp, len(regexArr))

	for idx, pattern := range regexArr {
		filters[idx] = regexp.MustCompile(pattern)
	}

	for idx, filter := range filters {

		if filter.MatchString(uri) {
			// log.Printf("Url was matched: %q", uri)
			switch idx {
			case 0:
				//log.Printf("http / https: %v", uri)
				return []string{uri}
			case 1:
				// log.Printf("protocol relative: %v", uri)
				uri = strings.Trim(uri, "//")
				return []string{goodPrefix[1] + uri}
			default:
				log.Printf("Wtf is this? %v", uri)
			}
		}

	}
	return []string{}
}

func (sm *SourceMap) CheckJsFileHostname() {

	h := ExtractHostname(sm)
	var updatedJsFiles []string

	for _, jsFile := range sm.JsFiles {

		// this regex is for absolute path: /static/example.js

		if regexp.MustCompile("^/[^/]").MatchString(jsFile) {

			// log.Println("matched")

			updatedJsFiles = append(updatedJsFiles, goodPrefix[1]+h.Hostname()+jsFile)

		} else if strings.Contains(jsFile, h.Hostname()) {

			// log.Printf("The url: %v appears to be associated with your website: %v", h, sm.Url)
			jsSlice := urlCleaner(jsFile)
			updatedJsFiles = append(updatedJsFiles, jsSlice...)

		}
	}

	sm.JsFiles = updatedJsFiles

}
