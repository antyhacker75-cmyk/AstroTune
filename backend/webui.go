package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"github.com/antyhacker75-cmyk/BitChord/backend/clock"
)

//go:embed web/templates/* web/static/*
var webAssets embed.FS

var inviteTmpl = template.Must(template.ParseFS(webAssets, "web/templates/invite.html"))

var partiesTmpl = template.Must(
	template.ParseFS(webAssets, "web/templates/parties.html"),
)

func registerWebUI(mux *http.ServeMux) {
	mux.HandleFunc("GET /parties", handlePartiesPage)
	mux.HandleFunc("GET /api/parties", handlePartiesAPI)

	staticContent, _ := fs.Sub(webAssets, "web/static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))
}

type partyView struct {
	Code       string `json:"code"`
	Host       string `json:"host"`
	Members    int    `json:"users"`
	NowPlaying string `json:"nowPlaying"`
	Artwork    string `json:"artwork"`
	Progress   int64  `json:"progress"`
	Duration   int64  `json:"duration"`
	Messages   int    `json:"messages"`
	HostOnly   bool   `json:"hostOnly"`
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
		msgCount := p.MessageCount
		hostOnly := p.HostOnlyControl
		nowPlaying := ""
		art := ""
		var duration int64 = 0
		var progress int64 = 0
		
		host := "Unknown"
		if h := p.Host(); h != nil {
			host = h.DisplayName
		}

		if p.Playback != nil && p.Playback.Track != nil {
			nowPlaying = p.Playback.Track.Title
			if p.Playback.Track.Artist != "" {
				nowPlaying += " — " + p.Playback.Track.Artist
			}
			if p.Playback.Track.ThumbnailUrl != nil {
				art = *p.Playback.Track.ThumbnailUrl
			}
			if p.Playback.Track.DurationMs != nil {
				duration = *p.Playback.Track.DurationMs / 1000
			}
			progress = p.Playback.PositionMs / 1000
			if p.Playback.IsPlaying && p.Playback.AnchorMs > 0 {
				progress += (clock.NowMs() - p.Playback.AnchorMs) / 1000
			}
			if progress > duration && duration > 0 {
				progress = duration
			}
		}
		
		views = append(views, partyView{
			Code:       p.Code,
			Host:       host,
			Members:    len(p.Members),
			NowPlaying: nowPlaying,
			Artwork:    art,
			Progress:   progress,
			Duration:   duration,
			Messages:   msgCount,
			HostOnly:   hostOnly,
		})
		p.Unlock()
	}
	_ = json.NewEncoder(w).Encode(views)
}
