package usecase

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

/*
PlaylistUseCase performs orchestration logic
*/
type PlaylistUseCase struct {
	mu        sync.Mutex
	rdbmsRepo repository.PlaylistRepository
	cacheRepo repository.PlaylistRepository

	playing  bool
	paused   bool
	position time.Duration

	stopChan chan struct{}
	logger   *slog.Logger
}

func NewPlaylistUseCase(rdbmsRepo, cacheRepo repository.PlaylistRepository, logger *slog.Logger) *PlaylistUseCase {
	return &PlaylistUseCase{
		rdbmsRepo: rdbmsRepo,
		cacheRepo: cacheRepo,
		logger:    logger,
	}
}

func (uc *PlaylistUseCase) InitCache() error {
	const op = "usecase.PlaylistUseCase.InitCache"
	operationLogger := uc.logger.With(slog.String("op", op))

	operationLogger.Debug("Initializing cache")

	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.rdbmsRepo.GetPlaylist()
	if err != nil {
		operationLogger.Error("Failed to get playlist from DB",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("InitCache: %w", err)
	}

	current := playlist.GetHead()
	for current != nil {
		if err := uc.cacheRepo.AddSong(current.Song); err != nil {
			operationLogger.Error("Failed to add song to Cache",
				slog.String("song_title", current.Song.Title),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("InitCache: %w", err)
		}
		operationLogger.Debug("Added song to Cache",
			slog.String("song_title", current.Song.Title),
		)

		current = current.Next
	}

	currentNode := playlist.GetCurrent()
	if currentNode != nil {
		if err := uc.cacheRepo.SetCurrent(currentNode); err != nil {
			operationLogger.Error("Failed to set current song in Cache",
				slog.String("song_title", currentNode.Song.Title),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("InitCache: %w", err)
		}
		operationLogger.Debug("Set current song in Cache",
			slog.String("song_title", currentNode.Song.Title),
		)
	}

	operationLogger.Debug("Cache initialized successfully")
	return nil
}

func (uc *PlaylistUseCase) Play() error {
	const op = "usecase.PlaylistUseCase.Play"
	operationLogger := uc.logger.With(slog.String("op", op))

	operationLogger.Debug("Play called")

	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.playing && !uc.paused {
		operationLogger.Warn("Play called, but already playing")
		return nil
	}

	uc.paused = false
	if !uc.playing {
		uc.playing = true
		uc.stopChan = make(chan struct{}, 1)
		go uc.playCurrentSong() // Playback emulation
		operationLogger.Debug("Playback started")
	} else {
		go uc.playCurrentSong()
		operationLogger.Debug("Resumed playback")
	}

	return nil
}

func (uc *PlaylistUseCase) Pause() error {
	const op = "usecase.PlaylistUseCase.Pause"
	operationLogger := uc.logger.With(slog.String("op", op))

	operationLogger.Debug("Pause called")

	uc.mu.Lock()
	defer uc.mu.Unlock()

	if !uc.playing {
		operationLogger.Warn("Pause called, but not playing")
		return ErrNotPlaying
	}
	if uc.paused {
		operationLogger.Warn("Pause called, but already paused")
		return ErrAlreadyPaused
	}

	uc.paused = true
	uc.playing = false

	select {
	case uc.stopChan <- struct{}{}:
	default:
	}

	operationLogger.Debug("Playback paused")

	return nil
}

func (uc *PlaylistUseCase) Next() error {
	const op = "usecase.PlaylistUseCase.Next"
	operationLogger := uc.logger.With(slog.String("op", op))

	operationLogger.Debug("Next called")

	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.cacheRepo.GetPlaylist()
	if err != nil {
		operationLogger.Error("Failed to get playlist from Cache",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", ErrGetPlaylistFromCache, err)
	}

	current := playlist.GetCurrent()
	if current == nil || current.Next == nil {
		operationLogger.Warn("No next song available in playlist")
		return fmt.Errorf("%w: %v", ErrNoNextSong, err)
	}

	uc.position = 0
	uc.playing = false
	uc.paused = false

	select {
	case uc.stopChan <- struct{}{}:
	default:
	}

	if err := uc.cacheRepo.SetCurrent(current.Next); err != nil {
		return fmt.Errorf("%w: %v", ErrSetCurrentInCache, err)
	}

	if err := uc.rdbmsRepo.SetCurrent(current.Next); err != nil {
		return fmt.Errorf("%w: %v", ErrSetCurrentInDB, err)
	}

	uc.playing = true
	uc.stopChan = make(chan struct{}, 1)
	go uc.playCurrentSong()
	operationLogger.Info("Moved to next song and started playback")

	return nil
}

func (uc *PlaylistUseCase) Prev() error {
	const op = "usecase.PlaylistUseCase.Prev"
	operationLogger := uc.logger.With(slog.String("op", op))

	operationLogger.Debug("Prev called")

	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.cacheRepo.GetPlaylist()
	if err != nil {
		operationLogger.Error("Failed to get playlist from Cache",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", ErrGetPlaylistFromCache, err)
	}

	current := playlist.GetCurrent()
	if current == nil || current.Prev == nil {
		operationLogger.Warn("No previous song available in playlist")
		return fmt.Errorf("%w: %v", ErrNoPrevSong, err)
	}

	uc.position = 0
	uc.playing = false
	uc.paused = false

	select {
	case uc.stopChan <- struct{}{}:
	default:
	}

	if err := uc.cacheRepo.SetCurrent(current.Prev); err != nil {
		return fmt.Errorf("%w: %v", ErrSetCurrentInCache, err)
	}
	if err := uc.rdbmsRepo.SetCurrent(current.Prev); err != nil {
		return fmt.Errorf("%w: %v", ErrSetCurrentInDB, err)
	}

	uc.playing = true
	uc.stopChan = make(chan struct{}, 1)
	go uc.playCurrentSong()
	operationLogger.Info("Moved to previous song and started playback")

	return nil
}

func (uc *PlaylistUseCase) GetCurrentSong() (*entity.Song, error) {
	const op = "usecase.PlaylistUseCase.GetCurrentSong"
	operationLogger := uc.logger.With(slog.String("op", op))

	operationLogger.Debug("GetCurrentSong called")

	uc.mu.Lock()
	defer uc.mu.Unlock()

	node, err := uc.cacheRepo.GetCurrent()
	if err != nil {
		operationLogger.Error("Failed to get current song from Cache",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("%w: %v", ErrGetCurrentNode, err)
	}
	if node == nil || node.Song == nil {
		operationLogger.Warn("No current song set in playlist")
		return nil, fmt.Errorf("%w: %v", ErrNoCurrentSong, err)
	}
	operationLogger.Debug("Retrieved current song",
		slog.String("title", node.Song.Title),
		slog.String("artist", node.Song.Artist),
	)

	return node.Song, nil
}

func (uc *PlaylistUseCase) GetPlaylist() (*entity.Playlist, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.cacheRepo.GetPlaylist()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetPlaylistFromCache, err)
	}
	return playlist, nil
}

/*
playCurrentSong is a method that emulates song playback
*/
func (uc *PlaylistUseCase) playCurrentSong() {
	const op = "usecase.playlist_usecase.playCurrentSong"
	operationLogger := uc.logger.With(slog.String("op", op))

	for {
		uc.mu.Lock()
		playlist, err := uc.cacheRepo.GetPlaylist()
		if err != nil {
			operationLogger.Error("Failed to get playlist", slog.String("error", err.Error()))
			uc.playing = false
			uc.mu.Unlock()
			return
		}

		current := playlist.GetCurrent()
		position := uc.position
		playing := uc.playing
		stopChan := uc.stopChan
		uc.mu.Unlock()

		if !playing {

			return // Exit if not play TODO: add logging and err handling
		}

		if current == nil || current.Song == nil {
			operationLogger.Debug("No song to play. Playback stopped.")
			uc.mu.Lock()
			uc.playing = false
			uc.mu.Unlock()
			return
		}
		duration := current.Song.Duration - position

		if duration <= 0 {

			uc.mu.Lock()
			if current.Next != nil {
				if err := playlist.SetCurrent(current.Next); err != nil {
					operationLogger.Error("Failed to set current song", slog.String("error", err.Error()))
					uc.playing = false
					uc.mu.Unlock()
					return
				}
				uc.position = 0
				uc.mu.Unlock()
				continue
			} else {
				operationLogger.Debug("No next song. Playback completed.")
				uc.playing = false
				uc.mu.Unlock()
				return
			}
		}

		operationLogger.Debug(
			"Playing song",
			slog.String("title", current.Song.Title),
			slog.Duration("remaining_duration", duration),
		)

		ticker := time.NewTicker(1 * time.Second)
		startTime := time.Now()

	playLoop:
		for {
			elapsed := time.Since(startTime)
			if elapsed >= duration {
				ticker.Stop()
				break playLoop
			}
			select {
			case <-ticker.C:
				uc.mu.Lock()
				if !uc.playing {
					uc.mu.Unlock()
					ticker.Stop()
					return
				}
				uc.position = elapsed
				uc.mu.Unlock()

			case <-stopChan:
				ticker.Stop()
				operationLogger.Debug(
					"Playback stopped for song",
					slog.String("title", current.Song.Title),
					slog.String("op", op),
				)
				uc.mu.Lock()
				uc.playing = false
				uc.mu.Unlock()
				return
			}
		}

		ticker.Stop()

		uc.mu.Lock()
		if current.Next != nil {
			if err := playlist.SetCurrent(current.Next); err != nil {
				operationLogger.Error("Failed to set current song", slog.String("error", err.Error()))
				uc.playing = false
				uc.mu.Unlock()
				return
			}
			uc.position = 0
			uc.mu.Unlock()
		} else {
			operationLogger.Debug("No next song. Playback completed.")
			uc.mu.Unlock()
			uc.playing = false
			return
		}
	}
}
