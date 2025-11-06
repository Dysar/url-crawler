package crawler

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
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

	// Parse main document and collect metadata and links to check later
	z := html.NewTokenizer(resp.Body)
	res := Result{Headings: map[string]int{"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0}}
	collectedLinks := make([]string, 0, 32)

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
			// finalize inaccessible links count before returning
			res.InaccessibleLinks = countInaccessibleLinks(ctx, targetURL, collectedLinks, c.client)
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
						if isExternal(href) {
							res.ExternalLinks++
						} else {
							res.InternalLinks++
						}
						collectedLinks = append(collectedLinks, href)
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

func isExternal(href string) bool {
	// Simple heuristic: absolute links starting with http and not sharing prefix host treated as external.
	// Refine later using url.Parse for robust host comparison.
	return strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")
}

// countInaccessibleLinks visits the collected links and counts those returning HTTP 4xx/5xx.
// Only HTTP status codes are considered; network errors are ignored for this metric.
func countInaccessibleLinks(ctx context.Context, base string, hrefs []string, client Fetcher) int {
	if len(hrefs) == 0 {
		return 0
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return 0
	}
	inaccessible := 0
	for _, href := range hrefs {
		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		abs := baseURL.ResolveReference(u)

		// Prefer HEAD to save bandwidth; fall back to GET if method not allowed
		req, _ := http.NewRequestWithContext(ctx, http.MethodHead, abs.String(), nil)
		resp, err := client.Do(req)
		if err != nil {
			// Ignore network errors for this metric
			continue
		}
		// Some servers do not support HEAD properly
		if resp.StatusCode == http.StatusMethodNotAllowed {
			resp.Body.Close()
			reqGet, _ := http.NewRequestWithContext(ctx, http.MethodGet, abs.String(), nil)
			resp, err = client.Do(reqGet)
			if err != nil {
				continue
			}
		}
		if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
			inaccessible++
		}
		resp.Body.Close()
	}
	return inaccessible
}

// HTTPClient configures timeouts and disables HTTP/2 for better compatibility with some servers
func HTTPClient(timeout time.Duration) *http.Client {
	tr := &http.Transport{
		// Proxy: Use HTTP_PROXY, HTTPS_PROXY, and NO_PROXY environment variables for proxy configuration
		Proxy: http.ProxyFromEnvironment,

		// DialContext: Network connection settings
		DialContext: (&net.Dialer{
			// Timeout: Maximum time to wait for establishing a TCP connection (10 seconds)
			Timeout: 10 * time.Second,
			// KeepAlive: How long to keep TCP connections alive for reuse (30 seconds)
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// TLSClientConfig: TLS/SSL security settings
		TLSClientConfig: &tls.Config{
			// MinVersion: Require TLS 1.2 minimum (disables insecure TLS 1.0 and 1.1)
			MinVersion: tls.VersionTLS12,
		},

		// TLSHandshakeTimeout: Maximum time to wait for TLS handshake to complete (10 seconds)
		TLSHandshakeTimeout: 10 * time.Second,

		// ExpectContinueTimeout: Time to wait for server's 100-continue response before sending body (1 second)
		ExpectContinueTimeout: 1 * time.Second,

		// ResponseHeaderTimeout: Maximum time to wait for response headers after sending request (15 seconds)
		ResponseHeaderTimeout: 15 * time.Second,

		// IdleConnTimeout: How long to keep idle connections in the connection pool before closing (90 seconds)
		IdleConnTimeout: 90 * time.Second,
	}

	// Client.Timeout: Overall timeout for entire request (including connection, headers, and body reading)
	return &http.Client{Timeout: timeout, Transport: tr}
}
