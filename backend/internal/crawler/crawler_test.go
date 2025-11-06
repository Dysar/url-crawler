package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCrawl_BasicExtraction(t *testing.T) {
	// Build HTML after test server is created so we can include an absolute same-host link
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	html := `<!doctype html>
    <html>
      <head>
        <title>Test Page</title>
      </head>
      <body>
        <h1>Main</h1>
        <h2>Sub</h2>
        <a href="/about">About</a>
        <a href="` + ts.URL + `/ext">External</a>
        <a href="#section">Skip</a>
        <a href="javascript:void(0)">SkipJS</a>
      </body>
    </html>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ext", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	c := New(HTTPClient(5 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.Crawl(ctx, ts.URL)
	if err != nil {
		t.Fatalf("crawl error: %v", err)
	}

	if res.HTMLVersion == nil || *res.HTMLVersion != "HTML5" {
		t.Fatalf("expected HTML5, got %#v", res.HTMLVersion)
	}
	if res.Title == nil || *res.Title != "Test Page" {
		t.Fatalf("expected title 'Test Page', got %#v", res.Title)
	}
	if res.Headings["h1"] != 1 || res.Headings["h2"] != 1 {
		t.Fatalf("unexpected headings: %+v", res.Headings)
	}
	if res.InternalLinks != 1 {
		t.Fatalf("expected 1 internal link, got %d", res.InternalLinks)
	}
	if res.ExternalLinks != 1 {
		t.Fatalf("expected 1 external link, got %d", res.ExternalLinks)
	}
	if res.HasLoginForm {
		t.Fatalf("expected no login form, got true")
	}
	if res.InaccessibleLinks != 0 {
		t.Fatalf("expected inaccessible links = 0, got %d", res.InaccessibleLinks)
	}
}

func TestCrawl_LoginFormDetection(t *testing.T) {
	html := `<!doctype html>
    <html><body>
      <form>
        <input type="text" name="u" />
        <input type="password" name="p" />
      </form>
    </body></html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer ts.Close()

	c := New(HTTPClient(5 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.Crawl(ctx, ts.URL)
	if err != nil {
		t.Fatalf("crawl error: %v", err)
	}
	if !res.HasLoginForm {
		t.Fatalf("expected HasLoginForm=true")
	}
}

func TestCrawl_HeadingsAllLevels(t *testing.T) {
	html := `<!doctype html>
    <html><body>
      <h1>a</h1><h2>b</h2><h3>c</h3><h4>d</h4><h5>e</h5><h6>f</h6>
    </body></html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer ts.Close()

	c := New(HTTPClient(5 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.Crawl(ctx, ts.URL)
	if err != nil {
		t.Fatalf("crawl error: %v", err)
	}
	for _, level := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		if res.Headings[level] != 1 {
			t.Fatalf("expected %s=1, got %d", level, res.Headings[level])
		}
	}
}

// TestCrawl_InaccessibleLinks_LocalE2E provides a deterministic e2e-style check
// for inaccessible link counting using a local HTTP server.
func TestCrawl_InaccessibleLinks_LocalE2E(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!doctype html><html><body>
          <a href="/ok">OK</a>
          <a href="/nf">NotFound</a>
          <a href="/err">Error</a>
        </body></html>`
		_, _ = w.Write([]byte(html))
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/nf", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("nf"))
	})
	mux.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("err"))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := New(HTTPClient(5 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.Crawl(ctx, ts.URL)
	if err != nil {
		t.Fatalf("crawl error: %v", err)
	}
	if res.InaccessibleLinks != 2 {
		t.Fatalf("expected inaccessible links = 2, got %d", res.InaccessibleLinks)
	}
}

func TestCrawl_AbsoluteSameHostCountedAsExternalByHeuristic(t *testing.T) {
	// The current heuristic treats all absolute http(s) links as external, even if same host.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!doctype html><html><body>
          <a href="` + tsURL(r) + `/abs">AbsSameHost</a>
          <a href="/rel">RelLink</a>
        </body></html>`
		_, _ = w.Write([]byte(html))
	}))
	defer ts.Close()

	c := New(HTTPClient(5 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.Crawl(ctx, ts.URL)
	if err != nil {
		t.Fatalf("crawl error: %v", err)
	}
	if res.InternalLinks != 1 {
		t.Fatalf("expected 1 internal (relative) link, got %d", res.InternalLinks)
	}
	if res.ExternalLinks != 1 {
		t.Fatalf("expected 1 external (absolute) link, got %d", res.ExternalLinks)
	}
}

func TestCrawl_InaccessibleLinksCount(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!doctype html><html><body>
          <a href="/ok">OK</a>
          <a href="/nf">NotFound</a>
          <a href="/err">Error</a>
        </body></html>`
		_, _ = w.Write([]byte(html))
	})
	// HEAD and GET for /ok -> 200
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// /nf -> 404
	mux.HandleFunc("/nf", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("nf"))
	})
	// /err -> 500
	mux.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("err"))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := New(HTTPClient(5 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.Crawl(ctx, ts.URL)
	if err != nil {
		t.Fatalf("crawl error: %v", err)
	}

	if res.InaccessibleLinks != 2 {
		t.Fatalf("expected inaccessible links = 2, got %d", res.InaccessibleLinks)
	}
}

// tsURL constructs the absolute URL for the current httptest server from the request.
func tsURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
