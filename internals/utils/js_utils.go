package utils

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"crypto/tls"
	"time"
	"github.com/MathieuTurcotte/sourcemap"
)

type SourceMap struct {
	Url     string
	JsFiles []string
	JsBody  string
}

var tr = &http.Transport{
	MaxIdleConns:        100,              // The maximum number of idle (keep-alive) connections
	MaxIdleConnsPerHost: 20,               // The maximum number of idle connections per host
	IdleConnTimeout:     90 * time.Second, // Idle connection timeout
	// Set connection timeout parameters
	DisableCompression:    true,            // Disable compression to avoid overhead for large amounts of requests
	TLSHandshakeTimeout:   5 * time.Second, // TLS handshake timeout
	ExpectContinueTimeout: 1 * time.Second, // Timeout for HTTP 100-Continue resp
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
}

var client = &http.Client{
	Transport: tr,
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

func (sm *SourceMap) CheckSourcemap() bool {

	var body []byte

	for idx := range sm.JsFiles {

		// log.Print(idx)

		uri := sm.JsFiles[idx]



		resp,err := client.Get(uri)

		if err != nil {
			log.Println("error making js req")
			return false
		}
		
		defer resp.Body.Close()

		body,err = io.ReadAll(resp.Body)

		if err != nil {
			log.Println("eror rreading js body")
			return false
		}


		match := strings.Contains(string(body), "//# sourceMappingURL=")

		if match {
			return true
		}
	}

	return false
}

func (sm *SourceMap) ParseSourcemap() {
	for idx := range sm.JsFiles {
		log.Printf("Sourcemap url: %s",sm.JsFiles[idx])
		
		resp,err := client.Get(sm.JsFiles[idx]+".map")
		if err != nil {
			log.Printf("error requesting sourcemap js: %s",err)
		}

		defer resp.Body.Close()

		body,err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error reading sourcemap js body: %s",err)
		}

		reader := strings.NewReader(string(body))

		for {
			// n, err := reader.Read(body)
			
			sm,err := sourcemap.Read(reader)
			if err != nil {
				log.Printf("error reading sourcemap: %s",err)
			}

			// gtg -> https://pkg.go.dev/github.com/MathieuTurcotte/sourcemap#pkg-functions
			log.Print(sm.SourcesContent)

			if err == io.EOF {
			break
		}
		}

	}
}