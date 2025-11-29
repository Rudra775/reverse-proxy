package proxy

import (
	"io"
	"log"
	"net/http"
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

	req, _ := http.NewRequest(r.Method, be.URL+r.URL.Path, r.Body)
	req.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("backend error:", err)
		http.Error(w, "backend error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	log.Printf("%s -> %s (%s)", r.URL.Path, be.URL, time.Since(start))
}
