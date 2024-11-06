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

func ExtractHostname(sm *SourceMap) *url.URL {
	hostname, _ := url.Parse(sm.Url)
	//log.Printf("Hostname: %v", hostname.Hostname())
	return hostname
}

func urlCleaner(uri string) []string {

	goodPrefix := [3]string{"http://", "https://"}
	regexArr := [3]string{`^(http|https)://`, `^//`, `^/[^/]`}
	filters := make([]*regexp.Regexp, len(regexArr))

	for idx, pattern := range regexArr {
		filters[idx] = regexp.MustCompile(pattern)
	}

	for idx, filter := range filters {
		if filter.MatchString(uri) {
			switch idx {
			case 0:
				//log.Printf("http / https: %v", uri)
				return []string{uri}
			case 1:
				//log.Printf("protocol relative: %v", uri)
				uri = strings.Trim(uri, "//")
				return []string{goodPrefix[1] + uri}
			case 2:
				//log.Printf("absolute path: %v", uri)
				uri = strings.Trim(uri, "/")
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
	for _, jsFile := range sm.JsFiles {
		if strings.Contains(jsFile, h.Hostname()) {

			//log.Printf("The url: %v appears to be associated with your website: %v", h, sm.Url)

			jsSlice := urlCleaner(jsFile)
			sm.JsFiles = jsSlice
		}
	}

}
