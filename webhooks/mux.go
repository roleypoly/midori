package webhooks

import (
	"net/http"

	"github.com/lampjaw/discordgobot"
)

type WebhookMux struct {
	Mux *http.ServeMux
	Bot *discordgobot.Gobot
}

func NewWebhookMux(bot *discordgobot.Gobot) *WebhookMux {
	mux := http.NewServeMux()

	m := &WebhookMux{
		Mux: mux,
		Bot: bot,
	}

	mux.HandleFunc("/", m.index)

	return m
}

func (mux *WebhookMux) index(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		mux.unimplemented(rw, r)
		return
	}

	rw.WriteHeader(303)
	rw.Header().Add("Location", "https://roleypoly.com")
}

func (mux *WebhookMux) unimplemented(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(404)
}
