package service

import (
	"cloud-go-testtask/internal/entity"
	"log/slog"
	"sync"
	"time"
)

/*
PlaylistRepositoryInterface is a contract for repository
*/
type PlaylistRepositoryInterface interface {
	AddSong(song *entity.Song) error
	GetPlaylist() (*entity.Playlist, error)
}

/*
PlaylistService is an implementation of structure that can control repository layer access and playlist
*/
type PlaylistService struct {
	repo     PlaylistRepositoryInterface
	mu       sync.Mutex
	stopChan chan struct{}
	paused   bool
	playing  bool
	position time.Duration
	logger   *slog.Logger
}

func NewPlaylistService(repo PlaylistRepositoryInterface, logger *slog.Logger) *PlaylistService {
	return &PlaylistService{
		repo:     repo,
		stopChan: make(chan struct{}, 1),
		logger:   logger,
	}
}

func (s *PlaylistService) Play() error {
	const op = "service.playlist_service.Play"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.playing {
		operationLogger.Warn("Already playing")
		return nil
	}

	if s.paused {
		s.paused = false
		s.playing = true
		operationLogger.Info("Resuming playback")
		go s.playCurrentSong()
		return nil
	}

	playlist, err := s.repo.GetPlaylist()
	if err != nil {
		operationLogger.Error("Failed to get playlist", slog.String("error", err.Error()))
		return err
	}

	if playlist.GetCurrent() == nil {
		operationLogger.Warn("No song to play")
		return nil
	}

	s.playing = true
	operationLogger.Info("Starting playback")
	go s.playCurrentSong()

	return nil
}

func (s *PlaylistService) Pause() error {
	const op = "service.playlist_service.Pause"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.playing || s.paused {
		operationLogger.Warn("Already paused or not playing")
		return nil
	}

	s.paused = true
	s.playing = false
	operationLogger.Info("Playback paused...")

	select {
	case s.stopChan <- struct{}{}:
	default:
	}

	return nil
}

func (s *PlaylistService) AddSong(song *entity.Song) error {
	const op = "service.playlist_service.AddSong"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.repo.AddSong(song); err != nil {
		operationLogger.Error("Failed to add song", slog.String("error", err.Error()))
		return err
	}

	operationLogger.Info("Song added to playlist", slog.String("title", song.Title))

	return nil
}

func (s *PlaylistService) Next() error {
	const op = "service.playlist_service.Next"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	playlist, err := s.repo.GetPlaylist()
	if err != nil {
		operationLogger.Error("Failed to get playlist", slog.String("error", err.Error()))
		return err
	}

	current := playlist.GetCurrent()
	if current == nil || current.Next == nil {
		operationLogger.Warn("No next song to play")
		return nil
	}

	select {
	case s.stopChan <- struct{}{}:
	default:
	}

	if err := playlist.SetCurrent(current.Next); err != nil {
		operationLogger.Error("Failed to set current song", slog.String("error", err.Error()))
		return err
	}

	s.position = 0
	s.paused = false
	s.playing = true

	operationLogger.Info("Switched to next song", slog.String("title", current.Next.Song.Title))

	go s.playCurrentSong()

	return nil
}

func (s *PlaylistService) Prev() error {
	const op = "service.playlist_service.Prev"
	operationLogger := s.logger.With(slog.String("op", op))

	s.mu.Lock()
	defer s.mu.Unlock()

	playlist, err := s.repo.GetPlaylist()
	if err != nil {
		operationLogger.Error("Failed to get playlist", slog.String("error", err.Error()))
		return err
	}

	current := playlist.GetCurrent()
	if current == nil || current.Prev == nil {
		operationLogger.Warn("No previous song to play")
		return nil
	}

	select { // Fix: send signal to channel only if it is not full
	case s.stopChan <- struct{}{}:
	default:
	}

	//playlist.Current = playlist.Current.Prev

	if err := playlist.SetCurrent(current.Prev); err != nil {
		operationLogger.Error("Failed to set current song", slog.String("error", err.Error()))
		return err
	}

	s.position = 0
	s.paused = false
	s.playing = true

	operationLogger.Info("Switched to previous song", slog.String("title", current.Prev.Song.Title))

	go s.playCurrentSong()

	return nil
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

		playlist, err := s.repo.GetPlaylist()
		if err != nil {
			operationLogger.Error("Failed to get playlist", slog.String("error", err.Error()))
			s.mu.Unlock()
			return
		}

		current := playlist.GetCurrent()
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

		if current.Next != nil {
			if err := playlist.SetCurrent(current.Next); err != nil {
				operationLogger.Error("Failed to set current song", slog.String("error", err.Error()))
				s.mu.Unlock()
				return
			}
			s.position = 0
		} else {
			operationLogger.Info("No next song. Playback completed.")
			s.mu.Unlock()
			return
		}

		s.mu.Unlock()
	}
}
