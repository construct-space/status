package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

//go:embed web/*
var webFS embed.FS

type Service struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
}

type ServiceStatus struct {
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	Status    string `json:"status"`
	Code      int    `json:"code,omitempty"`
	Latency   int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
	CheckedAt string `json:"checked_at"`
}

var services = []Service{
	{Name: "Accounts", URL: "https://accounts.construct.space/api/health", Domain: "accounts.construct.space"},
	{Name: "Blog", URL: "https://construct.blog/ghost/api/admin/site/", Domain: "construct.blog"},
	{Name: "Delivery", URL: "https://construct.delivery/health", Domain: "construct.delivery"},
	{Name: "Developer", URL: "https://developer.construct.space/api/health", Domain: "developer.construct.space"},
	{Name: "Domains", URL: "https://domains.construct.space/api/health", Domain: "domains.construct.space"},
	{Name: "Oracle", URL: "https://oracle.construct.space/api/health", Domain: "oracle.construct.space"},
	{Name: "PaaS", URL: "https://paas.construct.ninja/api/health", Domain: "paas.construct.ninja"},
	{Name: "Source", URL: "https://source.construct.space/api/health", Domain: "source.construct.space"},
	{Name: "Website", URL: "https://construct.space/api/health", Domain: "construct.space"},
}

var (
	cachedStatuses []ServiceStatus
	cacheMu        sync.RWMutex
	cacheTime      time.Time
)

func checkService(svc Service) ServiceStatus {
	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(svc.URL)
	latency := time.Since(start).Milliseconds()

	status := ServiceStatus{
		Name:      svc.Name,
		Domain:    svc.Domain,
		Latency:   latency,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		status.Status = "offline"
		status.Error = err.Error()
		return status
	}
	defer resp.Body.Close()

	status.Code = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		status.Status = "operational"
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		status.Status = "degraded"
	} else {
		status.Status = "offline"
	}

	return status
}

func checkAllServices() []ServiceStatus {
	var wg sync.WaitGroup
	results := make([]ServiceStatus, len(services))

	for i, svc := range services {
		wg.Add(1)
		go func(idx int, s Service) {
			defer wg.Done()
			results[idx] = checkService(s)
		}(i, svc)
	}

	wg.Wait()
	return results
}

func updateCache() {
	statuses := checkAllServices()
	cacheMu.Lock()
	cachedStatuses = statuses
	cacheTime = time.Now()
	cacheMu.Unlock()
}

func startBackgroundChecker() {
	updateCache()
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			updateCache()
		}
	}()
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	startBackgroundChecker()

	mux := http.NewServeMux()

	// API: get all service statuses
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		cacheMu.RLock()
		statuses := cachedStatuses
		checked := cacheTime
		cacheMu.RUnlock()

		operational := 0
		for _, s := range statuses {
			if s.Status == "operational" {
				operational++
			}
		}

		overall := "operational"
		if operational < len(statuses) && operational > 0 {
			overall = "degraded"
		} else if operational == 0 {
			overall = "outage"
		}

		writeJSON(w, 200, map[string]any{
			"overall":    overall,
			"services":   statuses,
			"checked_at": checked.UTC().Format(time.RFC3339),
			"total":      len(statuses),
			"operational": operational,
		})
	})

	// API: force refresh
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		updateCache()
		cacheMu.RLock()
		statuses := cachedStatuses
		cacheMu.RUnlock()
		writeJSON(w, 200, map[string]any{"services": statuses, "refreshed": true})
	})

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"status": "ok"})
	})

	// Serve frontend (SPA fallback)
	webContent, _ := fs.Sub(webFS, "web")
	fileServer := http.FileServer(http.FS(webContent))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Try static file first
		if r.URL.Path != "/" {
			f, err := fs.Stat(webContent, r.URL.Path[1:])
			if err == nil && !f.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// SPA fallback: serve index.html
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	log.Printf("[status] Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
