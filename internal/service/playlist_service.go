package service

import (
	"cloud-go-testtask/internal/repository"
	"log/slog"
	"sync"
	"time"
)

/*
PlaylistService is an implementation of structure that can control repository layer access and playlist
*/
type PlaylistService struct {
	repo     *repository.PlaylistRepository
	mu       sync.Mutex
	stopChan chan struct{}
	paused   bool
	position time.Duration
	logger   *slog.Logger
}

func NewPlaylistService(repo *repository.PlaylistRepository, logger *slog.Logger) *PlaylistService {
	return &PlaylistService{
		repo:     repo,
		stopChan: make(chan struct{}),
		logger:   logger,
	}
}

/*
playCurrentSong is a method that emulates song playback
*/
func (s *PlaylistService) playCurrentSong() {
	const op = "service.playlist_service.playCurrentSong"

	s.logger = s.logger.With( // Add details to logger about performed operation
		slog.String("op", op),
	)

	for {
		s.mu.Lock()
		playlist := s.repo.GetPlaylist()
		current := playlist.Current
		position := s.position
		s.mu.Unlock()

		if current == nil || current.Song == nil {
			s.logger.Info("No song to play. Playback stopped.")
			return // End playback if there is no next song
		}

		s.logger.Info(
			"Playing song",
			slog.String("title", current.Song.Title),
			slog.Duration("remaining_duration", current.Song.Duration-position),
		)
		// Calculate duration of song
		duration := current.Song.Duration - position
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		startTime := time.Now()

		// Playback current song
		for progress := time.Duration(0); progress < duration; progress = time.Since(startTime) {
			select {
			case <-ticker.C:
				s.mu.Lock()
				s.position = progress
				s.mu.Unlock()
			case <-s.stopChan:
				s.logger.Info(
					"Playback stopped for song",
					slog.String("title", current.Song.Title),
					slog.String("op", op),
				)
				return // Stop playback
			}
		}

		// Swith to next song
		s.mu.Lock()
		if playlist.Current.Next != nil {
			playlist.Current = playlist.Current.Next
			s.position = 0 // We start playback at 0s of next song
		} else {
			// There is no next song, stop
			s.logger.Info("No next song. Playback completed.")
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
	}
}
