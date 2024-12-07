package service

import (
	"cloud-go-testtask/internal/entity"
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
	playing  bool
	position time.Duration
	logger   *slog.Logger
}

func NewPlaylistService(repo *repository.PlaylistRepository, logger *slog.Logger) *PlaylistService {
	return &PlaylistService{
		repo:     repo,
		stopChan: make(chan struct{}, 1),
		logger:   logger,
	}
}

func (s *PlaylistService) Play() {
	const op = "service.playlist_service.Play"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	//TODO: add "playing" flag handling

	if s.playing {
		operationLogger.Warn("Already playing")
		return
	}

	if s.paused {
		s.paused = false
		operationLogger.Info("Resuming playback")
		return
	}

	s.playing = true
	operationLogger.Info("Starting playback")
	go s.playCurrentSong()
}

func (s *PlaylistService) Pause() {
	const op = "service.playlist_service.Pause"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.playing || s.paused {
		operationLogger.Warn("Already paused or not playing")
		return
	}

	s.paused = true
	operationLogger.Info("Playback paused...")

	select {
	case s.stopChan <- struct{}{}:
	default:
	}
}

func (s *PlaylistService) AddSong(song *entity.Song) {
	const op = "service.playlist_service.AddSong"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	s.repo.AddSong(song) // TODO: add err := s.repo ... for error handling

	operationLogger.Info("Song added to playlist", slog.String("title", song.Title))
}

func (s *PlaylistService) Next() {
	const op = "service.playlist_service.Next"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	playlist := s.repo.GetPlaylist()
	if playlist.Current == nil || playlist.Current.Next == nil {
		operationLogger.Warn("No next song to play")
		return
	}

	select {
	case s.stopChan <- struct{}{}:
	default:
	}

	playlist.Current = playlist.Current.Next
	s.position = 0
	s.playing = false // fix

	operationLogger.Info("Switched to next song", slog.String("title", playlist.Current.Song.Title))

	s.playing = true // fix
	go s.playCurrentSong()
}

func (s *PlaylistService) Prev() {
	const op = "service.playlist_service.Prev"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	playlist := s.repo.GetPlaylist()
	if playlist.Current == nil || playlist.Current.Prev == nil {
		s.logger.Warn("No previous song to play")
		return
	}

	select { // Fix: send signal to channel only if it is not full
	case s.stopChan <- struct{}{}:
	default:
	}

	playlist.Current = playlist.Current.Prev
	s.position = 0
	s.playing = false

	operationLogger.Info("Switched to previous song", slog.String("title", playlist.Current.Song.Title))

	s.playing = true
	go s.playCurrentSong()
}

/*
playCurrentSong is a method that emulates song playback
*/
func (s *PlaylistService) playCurrentSong() {
	const op = "service.playlist_service.playCurrentSong"

	operationLogger := s.logger.With( // Add details to logger about performed operation
		slog.String("op", op),
	)

	defer func() {
		s.mu.Lock()
		s.stopChan = make(chan struct{}, 1) // Reset the stop channel after playback ends
		s.mu.Unlock()
	}()

	for {
		s.mu.Lock()
		playlist := s.repo.GetPlaylist()
		current := playlist.Current
		position := s.position
		s.mu.Unlock()

		if current == nil || current.Song == nil {
			operationLogger.Info("No song to play. Playback stopped.")
			return // End playback if there is no next song
		}

		operationLogger.Info(
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
				operationLogger.Info(
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
			operationLogger.Info("No next song. Playback completed.")
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
	}
}
