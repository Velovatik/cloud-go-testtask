package delivery

import (
	"cloud-go-testtask/internal/usecase"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// There will be handlers for API implementation based on usecases

type PlaylistHandler struct {
	uc     *usecase.PlaylistUseCase
	logger *slog.Logger
}

func NewPlaylistHandler(uc *usecase.PlaylistUseCase, logger *slog.Logger) *PlaylistHandler {
	return &PlaylistHandler{
		uc:     uc,
		logger: logger,
	}
}

type addSongRequest struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Duration int    `json:"duration"`
}

func (h *PlaylistHandler) AddSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "delivery.PlaylistHandler.AddSongHandler"
	operationLogger := h.logger.With(slog.String("op", op))

	operationLogger.Info("Received AddSong request")

	var req addSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		operationLogger.Error("Failed to decode request body", slog.String("error", err.Error()))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Artist == "" || req.Duration <= 0 {
		operationLogger.Warn("Invalid song parameters",
			slog.String("title", req.Title),
			slog.String("artist", req.Artist),
			slog.Int("duration", req.Duration),
		)
		http.Error(w, "invalid song parameters", http.StatusBadRequest)
		return
	}

	err := h.uc.AddSong(req.Title, req.Artist, time.Duration(req.Duration)*time.Second)
	if err != nil {

		switch {
		case errors.Is(err, usecase.ErrAddSongToDB):
			operationLogger.Error("Failed to add song to database", slog.String("error", err.Error()))
			http.Error(w, "failed to add song to database", http.StatusInternalServerError)
		case errors.Is(err, usecase.ErrAddSongToCache):
			operationLogger.Error("Failed to add song to cache", slog.String("error", err.Error()))
			http.Error(w, "failed to add song to cache", http.StatusInternalServerError)
		default:
			operationLogger.Error("Failed to add song", slog.String("error", err.Error()))
			http.Error(w, "failed to add song", http.StatusInternalServerError)
		}
		return

		//http.Error(w, "failed to add song: "+err.Error(), http.StatusInternalServerError)
		//return
	}

	operationLogger.Info("Song added successfully",
		slog.String("title", req.Title),
		slog.String("artist", req.Artist),
	)

	w.WriteHeader(http.StatusCreated)
}

func (h *PlaylistHandler) PlayHandler(w http.ResponseWriter, r *http.Request) {
	const op = "delivery.PlaylistHandler.PlayHandler"
	operationLogger := h.logger.With(slog.String("op", op))

	operationLogger.Info("Received Play request")

	if err := h.uc.Play(); err != nil {
		operationLogger.Error("Failed to start playback", slog.String("error", err.Error()))
		http.Error(w, "failed to play: "+err.Error(), http.StatusInternalServerError)
		return
	}

	operationLogger.Info("Playback started successfully")
	w.WriteHeader(http.StatusOK)
}

func (h *PlaylistHandler) PauseHandler(w http.ResponseWriter, r *http.Request) {
	const op = "delivery.PlaylistHandler.PauseHandler"
	operationLogger := h.logger.With(slog.String("op", op))

	operationLogger.Info("Received Pause request")

	if err := h.uc.Pause(); err != nil {
		operationLogger.Warn("Failed to pause playback", slog.String("error", err.Error()))
		http.Error(w, "failed to pause: "+err.Error(), http.StatusConflict)
		return
	}

	operationLogger.Info("Playback paused successfully")
	w.WriteHeader(http.StatusOK)
}

func (h *PlaylistHandler) NextHandler(w http.ResponseWriter, r *http.Request) {
	const op = "delivery.PlaylistHandler.NextHandler"
	operationLogger := h.logger.With(slog.String("op", op))

	operationLogger.Info("Received Next request")

	if err := h.uc.Next(); err != nil {
		operationLogger.Warn("Failed to move to next song", slog.String("error", err.Error()))
		http.Error(w, "failed to next: "+err.Error(), http.StatusNotFound)
		return
	}

	operationLogger.Info("Moved to next song successfully")
	w.WriteHeader(http.StatusOK)
}

func (h *PlaylistHandler) PrevHandler(w http.ResponseWriter, r *http.Request) {
	const op = "delivery.PlaylistHandler.PrevHandler"
	operationLogger := h.logger.With(slog.String("op", op))

	operationLogger.Info("Received Prev request")

	if err := h.uc.Prev(); err != nil {
		operationLogger.Warn("Failed to move to previous song", slog.String("error", err.Error()))
		http.Error(w, "failed to prev: "+err.Error(), http.StatusNotFound)
		return
	}

	operationLogger.Info("Moved to previous song successfully")
	w.WriteHeader(http.StatusOK)
}

func (h *PlaylistHandler) GetCurrentSongHandler(w http.ResponseWriter, r *http.Request) {
	const op = "delivery.PlaylistHandler.GetCurrentSongHandler"
	operationLogger := h.logger.With(slog.String("op", op))

	operationLogger.Info("Received GetCurrentSong request")

	song, err := h.uc.GetCurrentSong()
	if err != nil {
		operationLogger.Warn("Failed to get current song", slog.String("error", err.Error()))
		http.Error(w, "no current song: "+err.Error(), http.StatusNotFound)
		return
	}
	operationLogger.Info("Current song retrieved successfully",
		slog.String("title", song.Title),
		slog.String("artist", song.Artist),
	)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(song); err != nil {
		operationLogger.Error("Failed to encode current song to JSON", slog.String("error", err.Error()))
		http.Error(w, "failed to encode song", http.StatusInternalServerError)
		return
	}

}

func (h *PlaylistHandler) GetPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	const op = "delivery.PlaylistHandler.GetPlaylistHandler"
	operationLogger := h.logger.With(slog.String("op", op))

	operationLogger.Info("Received GetPlaylist request")

	playlist, err := h.uc.GetPlaylist()
	if err != nil {
		operationLogger.Error("Failed to get playlist", slog.String("error", err.Error()))
		http.Error(w, "failed to get playlist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var sogs []map[string]interface{}
	current := playlist.GetHead()
	for current != nil {
		if current.Song != nil {
			sogs = append(sogs, map[string]interface{}{
				"id":       current.Song.ID,
				"title":    current.Song.Title,
				"artist":   current.Song.Artist,
				"duration": int(current.Song.Duration.Seconds()),
			})
		}
		current = current.Next
	}

	resp := map[string]interface{}{
		"sogs": sogs,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		operationLogger.Error("Failed to encode playlist to JSON", slog.String("error", err.Error()))
		http.Error(w, "failed to encode playlist", http.StatusInternalServerError)
		return
	}
}
