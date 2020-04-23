package run

import (
	"encoding/json"
	"net/http"
)

type Healthz struct {
	DiscordReady  bool
	WebhooksReady bool
}

func NewHealthz() *Healthz {
	return &Healthz{
		DiscordReady:  false,
		WebhooksReady: false,
	}
}

func (h *Healthz) Start(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleFunc)
	http.ListenAndServe(addr, mux)
}

func (h *Healthz) handleFunc(rw http.ResponseWriter, r *http.Request) {
	if h.DiscordReady && h.WebhooksReady {
		rw.WriteHeader(200)
		status := map[string]interface{}{
			"status": "ok",
		}
		json.NewEncoder(rw).Encode(status)
	} else {
		rw.WriteHeader(503)
		status := map[string]interface{}{
			"status": "not ready",
			"states": map[string]bool{
				"discord":  h.DiscordReady,
				"webhooks": h.WebhooksReady,
			},
		}
		json.NewEncoder(rw).Encode(status)
	}
}
