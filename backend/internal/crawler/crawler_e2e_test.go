//go:build e2e
// +build e2e

package crawler

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestCrawl_RealURLs tests the crawler against real URLs from the README.
// Run with: go test -tags=e2e ./internal/crawler
// Or set E2E_TEST=1 environment variable
func TestCrawl_RealURLs(t *testing.T) {
	if os.Getenv("E2E_TEST") == "" {
		t.Skip("Skipping e2e tests. Set E2E_TEST=1 to run")
	}

	testCases := []struct {
		name     string
		url      string
		validate func(t *testing.T, res Result)
	}{
		{
			name: "Minimal page should have a title, an H1 heading, and no login form",
			url:  "http://example.com",
			validate: func(t *testing.T, res Result) {
				if res.HTMLVersion == nil {
					t.Error("expected HTMLVersion to be set")
				}
				if res.Title == nil || *res.Title == "" {
					t.Error("expected title to be set")
				}
				if res.Headings["h1"] < 1 {
					t.Errorf("expected at least 1 h1, got %d", res.Headings["h1"])
				}
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
		{
			name: "Plain HTML should have a title, an H1 heading, and no login form",
			url:  "http://httpbin.org/html",
			validate: func(t *testing.T, res Result) {
				if res.Title == nil || *res.Title == "" {
					t.Error("expected title to be set")
				}
				if res.Headings["h1"] < 1 {
					t.Errorf("expected at least 1 h1, got %d", res.Headings["h1"])
				}
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
		{
			name: "Legacy page should have a title and at least one external link",
			url:  "https://info.cern.ch/hypertext/WWW/TheProject.html",
			validate: func(t *testing.T, res Result) {
				if res.Title == nil || *res.Title == "" {
					t.Error("expected title to be set")
				}
				if res.ExternalLinks < 1 {
					t.Errorf("expected at least 1 external link, got %d", res.ExternalLinks)
				}
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
		{
			name: "HTTP-only page should have a title and no login form",
			url:  "http://neverssl.com",
			validate: func(t *testing.T, res Result) {
				if res.Title == nil || *res.Title == "" {
					t.Error("expected title to be set")
				}
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
		{
			name: "Example page should have title, H1 and H2, and both link types",
			url:  "https://www.w3.org/Style/Examples/011/firstcss.en.html",
			validate: func(t *testing.T, res Result) {
				if res.Title == nil || *res.Title == "" {
					t.Error("expected title to be set")
				}
				if res.Headings["h1"] < 1 {
					t.Errorf("expected at least 1 h1, got %d", res.Headings["h1"])
				}
				if res.Headings["h2"] < 1 {
					t.Errorf("expected at least 1 h2, got %d", res.Headings["h2"])
				}
				if res.InternalLinks < 1 || res.ExternalLinks < 1 {
					t.Errorf("expected both internal and external links, got internal=%d external=%d",
						res.InternalLinks, res.ExternalLinks)
				}
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
		{
			name: "Form page should have a title and no login form",
			url:  "http://httpbin.org/forms/post",
			validate: func(t *testing.T, res Result) {
				if res.Title == nil || *res.Title == "" {
					t.Error("expected title to be set")
				}
				// httpbin forms/post doesn't have password field, so no login form
				if res.HasLoginForm {
					t.Error("expected no login form (no password input)")
				}
			},
		},
		{
			name: "Login page should have a password input field detected as login form",
			url:  "https://github.com/login",
			validate: func(t *testing.T, res Result) {
				if res.HTMLVersion == nil {
					t.Error("expected HTMLVersion to be set")
				}
				if !res.HasLoginForm {
					t.Error("expected login form to be detected (password input should be present)")
				}
			},
		},
		{
			name: "Spec index should have title, H1–H3, and many links",
			url:  "https://www.w3.org/TR/PNG/",
			validate: func(t *testing.T, res Result) {
				if res.Title == nil || *res.Title == "" {
					t.Error("expected title to be set")
				}
				if res.Headings["h1"] < 1 {
					t.Errorf("expected at least 1 h1, got %d", res.Headings["h1"])
				}
				if res.Headings["h2"] < 1 {
					t.Errorf("expected at least 1 h2, got %d", res.Headings["h2"])
				}
				if res.Headings["h3"] < 1 {
					t.Errorf("expected at least 1 h3, got %d", res.Headings["h3"])
				}
				if (res.InternalLinks + res.ExternalLinks) < 20 {
					t.Errorf("expected many links overall, got total=%d (internal=%d external=%d)",
						res.InternalLinks+res.ExternalLinks, res.InternalLinks, res.ExternalLinks)
				}
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
		{
			name: "404 page should be crawlable with HTML version and no login form",
			url:  "https://httpstat.us/404",
			validate: func(t *testing.T, res Result) {
				// Error pages should still be crawlable
				if res.HTMLVersion == nil {
					t.Error("expected HTMLVersion to be set")
				}
				// May or may not have title depending on implementation
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
		{
			name: "500 page should be crawlable with HTML version and no login form",
			url:  "https://httpstat.us/500",
			validate: func(t *testing.T, res Result) {
				// Error pages should still be crawlable
				if res.HTMLVersion == nil {
					t.Error("expected HTMLVersion to be set")
				}
				if res.HasLoginForm {
					t.Error("expected no login form")
				}
			},
		},
	}

	c := New(HTTPClient(30 * time.Second))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			res, err := c.Crawl(ctx, tc.url)
			if err != nil {
				t.Fatalf("crawl error for %s: %v", tc.url, err)
			}

			t.Logf("Results for %s:", tc.url)
			t.Logf("  HTMLVersion: %v", res.HTMLVersion)
			t.Logf("  Title: %v", res.Title)
			t.Logf("  Headings: h1=%d h2=%d h3=%d h4=%d h5=%d h6=%d",
				res.Headings["h1"], res.Headings["h2"], res.Headings["h3"],
				res.Headings["h4"], res.Headings["h5"], res.Headings["h6"])
			t.Logf("  InternalLinks: %d, ExternalLinks: %d", res.InternalLinks, res.ExternalLinks)
			t.Logf("  InaccessibleLinks: %d", res.InaccessibleLinks)
			t.Logf("  HasLoginForm: %v", res.HasLoginForm)

			tc.validate(t, res)
		})
	}
}
