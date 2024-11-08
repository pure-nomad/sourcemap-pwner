package internals

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"sourcemap-pwner/internals/utils"
	"strings"
	"sync"
	"time"
	// "log"
	"golang.org/x/net/html"
)


// TODO: basic auth support in urls.txt


func CheckUrl(url string, ch chan utils.SourceMap, wg *sync.WaitGroup) {

	defer wg.Done()

	tr := &http.Transport{
		MaxIdleConns:        100,              // The maximum number of idle (keep-alive) connections
		MaxIdleConnsPerHost: 20,               // The maximum number of idle connections per host
		IdleConnTimeout:     90 * time.Second, // Idle connection timeout
		// Set connection timeout parameters
		DisableCompression:    true,            // Disable compression to avoid overhead for large amounts of requests
		TLSHandshakeTimeout:   5 * time.Second, // TLS handshake timeout
		ExpectContinueTimeout: 1 * time.Second, // Timeout for HTTP 100-Continue resp
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get(url)

	if err != nil {
		log.Printf("Couldn't make GET request to %q. Error: %v.\n", url, err)
		os.Exit(0)
	}

	// 403 check to prevent empty structs being returned

	if resp.StatusCode == 403 {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Couldn't read the body of: %q. Error: %v\n", url, err)
		os.Exit(0)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	var valid utils.SourceMap
	var once sync.Once

	// so to find js files we will just parse the html and look for javascript file references. ex: <script src="/main.js"/>

	r := strings.NewReader(string(body))

	doc, err := html.Parse(r)
	if err != nil {
		log.Printf("Couldn't read the body of: %q. Error: %v.",url,err)
		os.Exit(0)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, a := range n.Attr {
				if a.Key == "src" {

					// this works great! but what if the javascript files are loaded externally?
					// we're going to add some methods to our sourcemap struct to do some checks on this :D
					// we should probably implement a manual mode that will show the user the urls so they can decide if they want to dump sourcemap from that js file.

					valid.Url = resp.Request.URL.String() // we set the url that we actually made the request to so that we detect any redirects.
					valid.JsFiles = append(valid.JsFiles, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	valid.CheckJsFileHostname()

	check := valid.CheckSourcemap()

	if check == true {
		log.Printf("Found sourcemap! %q",url)
		log.Println("Parsing...")
		valid.ParseSourcemap()
	} else {
		return
	}

	/* this call makes sure we only get one sourcemap struct back on the channel that has ALL the js files found :)
	AND NOW THEY ARE CLEANED READY FOR REQUESTS TO BE MADE! (CheckJsFileHostname) ... but I know I will have to parse the body of this as well to find sourceMappingURL=
	*/

	once.Do(func() { ch <- valid })
}
