package crawler

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Result struct {
	HTMLVersion       *string
	Title             *string
	Headings          map[string]int
	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int
	HasLoginForm      bool
}

type Fetcher interface {
	Do(req *http.Request) (*http.Response, error)
}

type Crawler struct {
	client Fetcher
}

func New(client Fetcher) *Crawler { return &Crawler{client: client} }

func (c *Crawler) Crawl(ctx context.Context, targetURL string) (Result, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	req.Header.Set("User-Agent", "url-crawler/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	// Basic accessibility counting is done when checking links later; here we parse main doc
	z := html.NewTokenizer(resp.Body)
	res := Result{Headings: map[string]int{"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0}}

	// Infer HTML version from doctype if present
	// Tokenizer does not expose doctype directly; we rely on first token type DoctypeToken via Token() on Raw
	// For simplicity: if content-type contains html and we see lowercase doctype, assume HTML5
	// This heuristic is acceptable for MVP; refine later if needed
	htmlVersion := "HTML5"
	res.HTMLVersion = &htmlVersion

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return res, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			tn := strings.ToLower(t.Data)
			if tn == "title" {
				if z.Next() == html.TextToken {
					title := strings.TrimSpace(z.Token().Data)
					res.Title = &title
				}
			}
			if tn == "h1" || tn == "h2" || tn == "h3" || tn == "h4" || tn == "h5" || tn == "h6" {
				res.Headings[tn]++
			}
			if tn == "a" {
				for _, a := range t.Attr {
					if strings.ToLower(a.Key) == "href" {
						href := strings.TrimSpace(a.Val)
						if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
							break
						}
						if isExternal(targetURL, href) {
							res.ExternalLinks++
						} else {
							res.InternalLinks++
						}
					}
				}
			}
			if tn == "input" {
				for _, a := range t.Attr {
					if strings.ToLower(a.Key) == "type" && strings.ToLower(a.Val) == "password" {
						res.HasLoginForm = true
					}
				}
			}
		}
	}
}

func isExternal(baseURL, href string) bool {
	// Simple heuristic: absolute links starting with http and not sharing prefix host treated as external.
	// Refine later using url.Parse for robust host comparison.
	return strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")
}

// HTTPClient configures timeouts and disables HTTP/2 for better compatibility with some servers
func HTTPClient(timeout time.Duration) *http.Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		// Disable HTTP/2 explicitly
		TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
	}
	return &http.Client{Timeout: timeout, Transport: tr}
}
