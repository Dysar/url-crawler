package crawler

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
	logrus.Infof("Starting crawl for URL: %s", targetURL)
	startTime := time.Now()

	// Validate URL before making request
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		logrus.Errorf("Invalid URL %s: %v", targetURL, err)
		return Result{}, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		logrus.Errorf("Unsupported URL scheme %s for URL: %s", parsedURL.Scheme, targetURL)
		return Result{}, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		logrus.Errorf("Failed to create request for %s: %v", targetURL, err)
		return Result{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "url-crawler/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := c.client.Do(req)
	if err != nil {
		// Check if it's an EOF error - some servers close connection immediately
		errStr := err.Error()
		if strings.Contains(errStr, "EOF") {
			// EOF typically means connection closed before response - treat as network error
			logrus.Warnf("Connection closed (EOF) for %s: %v", targetURL, err)
			return Result{}, fmt.Errorf("connection closed (EOF): %w", err)
		}
		logrus.Errorf("HTTP request failed for %s: %v", targetURL, err)
		return Result{}, err
	}
	defer resp.Body.Close()

	logrus.Debugf("HTTP response received for %s: status=%d, content-type=%s", targetURL, resp.StatusCode, resp.Header.Get("Content-Type"))

	// Note: We parse HTML even for 4xx/5xx status codes, as error pages often contain HTML

	// Check content-type to ensure we're parsing HTML
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") && !strings.Contains(strings.ToLower(contentType), "application/xhtml+xml") {
		// Not HTML, but continue parsing anyway as some servers don't set content-type correctly
		logrus.Warnf("Content-Type header does not indicate HTML for URL %s: got %q", targetURL, contentType)
	}

	// Parse main document and collect metadata and links to check later
	z := html.NewTokenizer(resp.Body)
	res := Result{Headings: map[string]int{"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0}}
	collectedLinks := make([]string, 0, 32)
	var baseURL *url.URL
	var htmlVersion *string
	var titleBuilder strings.Builder
	var inTitleTag bool

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			// finalize inaccessible links count before returning
			if baseURL == nil {
				baseURL = parsedURL
			}
			logrus.Debugf("Parsed HTML for %s: found %d links, %d headings total", targetURL, len(collectedLinks), res.Headings["h1"]+res.Headings["h2"]+res.Headings["h3"]+res.Headings["h4"]+res.Headings["h5"]+res.Headings["h6"])
			if len(collectedLinks) > 0 {
				logrus.Infof("Checking accessibility of %d links for %s", len(collectedLinks), targetURL)
			}
			res.InaccessibleLinks = countInaccessibleLinks(ctx, baseURL, collectedLinks, c.client)
			// Set title if we collected one
			if titleBuilder.Len() > 0 {
				title := strings.TrimSpace(titleBuilder.String())
				res.Title = &title
				logrus.Debugf("Extracted title for %s: %s", targetURL, title)
			}
			// Set HTML version if we detected one, otherwise default to HTML5
			if htmlVersion == nil {
				defaultVersion := "HTML5"
				htmlVersion = &defaultVersion
			}
			res.HTMLVersion = htmlVersion
			duration := time.Since(startTime)
			logrus.Infof("Completed crawl for %s in %v: %d internal, %d external links, %d inaccessible, login form: %v",
				targetURL, duration, res.InternalLinks, res.ExternalLinks, res.InaccessibleLinks, res.HasLoginForm)
			return res, nil
		case html.DoctypeToken:
			// Try to detect HTML version from doctype
			doctype := z.Token().Data
			doctypeLower := strings.ToLower(doctype)
			if strings.Contains(doctypeLower, "html5") || strings.Contains(doctypeLower, "html 5") {
				version := "HTML5"
				htmlVersion = &version
			} else if strings.Contains(doctypeLower, "html 4.01 strict") {
				version := "HTML 4.01 Strict"
				htmlVersion = &version
			} else if strings.Contains(doctypeLower, "html 4.01 transitional") {
				version := "HTML 4.01 Transitional"
				htmlVersion = &version
			} else if strings.Contains(doctypeLower, "html 4.01") {
				version := "HTML 4.01"
				htmlVersion = &version
			} else if strings.Contains(doctypeLower, "xhtml 1.0") {
				version := "XHTML 1.0"
				htmlVersion = &version
			} else if strings.Contains(doctypeLower, "xhtml 1.1") {
				version := "XHTML 1.1"
				htmlVersion = &version
			} else if strings.Contains(doctypeLower, "html") {
				// Generic HTML, default to HTML5
				version := "HTML5"
				htmlVersion = &version
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			tn := strings.ToLower(t.Data)
			if tn == "base" {
				// Handle base tag for relative URL resolution
				for _, a := range t.Attr {
					if strings.ToLower(a.Key) == "href" {
						baseHref := strings.TrimSpace(a.Val)
						if baseHref != "" {
							parsedBase, err := url.Parse(baseHref)
							if err == nil {
								baseURL = parsedURL.ResolveReference(parsedBase)
								logrus.Debugf("Found base tag for %s: %s", targetURL, baseURL.String())
							}
						}
					}
				}
			}
			if tn == "title" {
				// If we encounter a new title tag, reset and start collecting
				inTitleTag = true
				titleBuilder.Reset()
				// Handle self-closing title tag (rare but possible)
				if tt == html.SelfClosingTagToken {
					inTitleTag = false
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
						// Skip non-HTTP/HTTPS links (mailto:, tel:, etc.)
						hrefLower := strings.ToLower(href)
						if strings.HasPrefix(hrefLower, "mailto:") ||
							strings.HasPrefix(hrefLower, "tel:") ||
							strings.HasPrefix(hrefLower, "sms:") ||
							strings.HasPrefix(hrefLower, "ftp:") ||
							strings.HasPrefix(hrefLower, "file:") {
							break
						}
						// Resolve relative URLs
						resolvedHref := href
						if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
							// Relative URL - resolve against base
							if baseURL == nil {
								baseURL = parsedURL
							}
							hrefURL, err := url.Parse(href)
							if err == nil {
								resolvedURL := baseURL.ResolveReference(hrefURL)
								resolvedHref = resolvedURL.String()
							}
						}
						// Only process HTTP/HTTPS links
						if strings.HasPrefix(resolvedHref, "http://") || strings.HasPrefix(resolvedHref, "https://") {
							if isExternal(parsedURL, resolvedHref) {
								res.ExternalLinks++
							} else {
								res.InternalLinks++
							}
							// Only collect HTTP/HTTPS links for accessibility checking
							collectedLinks = append(collectedLinks, resolvedHref)
						}
					}
				}
			}
			if tn == "input" {
				for _, a := range t.Attr {
					if strings.ToLower(a.Key) == "type" && strings.ToLower(a.Val) == "password" {
						res.HasLoginForm = true
						logrus.Debugf("Login form detected on %s (password input found)", targetURL)
					}
				}
			}
		case html.EndTagToken:
			t := z.Token()
			tn := strings.ToLower(t.Data)
			if tn == "title" {
				inTitleTag = false
			}
		case html.TextToken:
			if inTitleTag {
				titleBuilder.WriteString(z.Token().Data)
			}
		}
	}
}

func isExternal(baseURL *url.URL, href string) bool {
	// Parse the href URL
	hrefURL, err := url.Parse(href)
	if err != nil {
		// If we can't parse it, treat as external to be safe
		return true
	}

	// Protocol-relative URLs (//example.com) are considered external if host differs
	if hrefURL.Scheme == "" && strings.HasPrefix(href, "//") {
		// Compare hosts
		return !strings.EqualFold(hrefURL.Host, baseURL.Host)
	}

	// If no scheme, it's relative (internal)
	if hrefURL.Scheme == "" {
		return false
	}

	// Compare schemes and hosts
	if hrefURL.Scheme != baseURL.Scheme {
		return true
	}

	// Compare hosts (case-insensitive)
	return !strings.EqualFold(hrefURL.Host, baseURL.Host)
}

// countInaccessibleLinks visits the collected links and counts those returning HTTP 4xx/5xx.
// Only HTTP status codes are considered; network errors are ignored for this metric.
// Non-HTTP/HTTPS links (mailto:, tel:, etc.) are skipped.
// Uses parallel processing with a worker pool to improve performance.
func countInaccessibleLinks(ctx context.Context, baseURL *url.URL, hrefs []string, client Fetcher) int {
	if len(hrefs) == 0 {
		return 0
	}

	// Filter to only HTTP/HTTPS links
	httpLinks := make([]string, 0, len(hrefs))
	for _, href := range hrefs {
		hrefLower := strings.ToLower(href)
		if strings.HasPrefix(hrefLower, "http://") || strings.HasPrefix(hrefLower, "https://") {
			u, err := url.Parse(href)
			if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
				httpLinks = append(httpLinks, href)
			}
		}
	}

	if len(httpLinks) == 0 {
		return 0
	}

	logrus.Debugf("Checking %d HTTP/HTTPS links for accessibility (filtered from %d total links)", len(httpLinks), len(hrefs))

	// Use parallel processing with a worker pool (max 10 concurrent requests)
	// to avoid overwhelming servers and improve performance
	const maxWorkers = 10
	workerLimit := min(maxWorkers, len(httpLinks))

	var wg sync.WaitGroup
	var mu sync.Mutex
	inaccessible := 0
	linkChan := make(chan string, len(httpLinks))

	// Send all links to channel
	for _, href := range httpLinks {
		linkChan <- href
	}
	close(linkChan)

	// Start workers
	for range workerLimit {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for href := range linkChan {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Create a shorter timeout per link (5 seconds) to avoid blocking
				linkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				isInaccessible := checkLinkAccessibility(linkCtx, baseURL, href, client)
				cancel()

				if isInaccessible {
					mu.Lock()
					inaccessible++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	if inaccessible > 0 {
		logrus.Infof("Found %d inaccessible links (4xx/5xx) out of %d checked", inaccessible, len(httpLinks))
	}
	return inaccessible
}

// checkLinkAccessibility checks if a single link is inaccessible (returns 4xx/5xx).
// Returns true if the link is inaccessible, false otherwise.
// Note: href is expected to be HTTP/HTTPS (already filtered by caller).
func checkLinkAccessibility(ctx context.Context, baseURL *url.URL, href string, client Fetcher) bool {
	u, err := url.Parse(href)
	if err != nil {
		return false
	}
	abs := baseURL.ResolveReference(u)

	// Prefer HEAD to save bandwidth; fall back to GET if method not allowed
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, abs.String(), nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		// Ignore network errors for this metric

		return false
	}
	defer resp.Body.Close()

	// Some servers do not support HEAD properly
	if resp.StatusCode == http.StatusMethodNotAllowed {
		resp.Body.Close()
		reqGet, err := http.NewRequestWithContext(ctx, http.MethodGet, abs.String(), nil)
		if err != nil {
			return false
		}
		resp, err = client.Do(reqGet)
		if err != nil {
			return false
		}
	}

	// Check status code and ensure we close the body
	isInaccessible := resp.StatusCode >= 400 && resp.StatusCode <= 599
	return isInaccessible
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

	// Client.Timeout: Overall timeout for entire request
	return &http.Client{Timeout: timeout, Transport: tr}
}
