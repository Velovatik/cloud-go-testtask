package delivery

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// TODO: Implement all CRUD methods

func NewRouter(h *PlaylistHandler) http.Handler {
	r := chi.NewRouter()

	r.Post("/songs", h.AddSongHandler)

	r.Get("/playlist", h.GetPlaylistHandler)
	r.Get("/current", h.GetCurrentSongHandler)

	r.Post("/play", h.PlayHandler)
	r.Post("/pause", h.PauseHandler)
	r.Post("/next", h.NextHandler)
	r.Post("/prev", h.PrevHandler)

	return r
}
