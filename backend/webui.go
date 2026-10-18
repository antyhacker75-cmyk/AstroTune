package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed web/templates/parties.html web/static/*
var webAssets embed.FS

var partiesTmpl = template.Must(
	template.ParseFS(webAssets, "web/templates/parties.html"),
)

// registerWebUI mounts /parties, /api/parties, and /static/.
// Call from setupTestServer() and main() after the other mux.HandleFunc calls.
func registerWebUI(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, webAssets, "web/templates/index.html")
	})
	mux.HandleFunc("GET /parties", handlePartiesPage)
	mux.HandleFunc("GET /api/parties", handlePartiesAPI)

	staticContent, _ := fs.Sub(webAssets, "web/static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))
}

type partyView struct {
	Code string `json:"code"`
	Host string `json:"host"`
	Members    int    `json:"users"`
	NowPlaying string `json:"nowPlaying"`
}

func handlePartiesPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = partiesTmpl.Execute(w, nil)
}

func handlePartiesAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	views := make([]partyView, 0, len(store.All()))
	for _, p := range store.All() {
		p.Lock()
		nowPlaying := ""
		if p.Playback != nil && p.Playback.Track != nil {
			nowPlaying = p.Playback.Track.Title
			if p.Playback.Track.Artist != "" {
				nowPlaying += " — " + p.Playback.Track.Artist
			}
		}
		views = append(views, partyView{
			Code:       p.Code,
			Members:    len(p.Members),
			Host: func() string { if h := p.Host(); h != nil { return h.DisplayName }; return "Unknown" }(),
			NowPlaying: nowPlaying,
		})
		p.Unlock()
	}
	_ = json.NewEncoder(w).Encode(views)
}
