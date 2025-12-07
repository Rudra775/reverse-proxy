package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type coreHandler struct {
	router  *Router
	metrics *Metrics
	timeout time.Duration
}

func (h *coreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.metrics.Inc()
	start := time.Now()

	be, err := h.router.Match(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	be.Inc()
	defer be.Dec()

	backendURL, err := url.Parse(be.URL)
	if err != nil {
		http.Error(w, "invalid backend url", http.StatusInternalServerError)
		return
	}

	target := backendURL.ResolveReference(r.URL) // merge host+path properly
	req, _ := http.NewRequest(r.Method, target.String(), r.Body)
	req.Header = r.Header.Clone()
	req.Host = backendURL.Host // important fix

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("backend error: %v", err)
		http.Error(w, "backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	log.Printf("%s -> %s (%s)", r.URL.Path, be.URL, time.Since(start))
}
