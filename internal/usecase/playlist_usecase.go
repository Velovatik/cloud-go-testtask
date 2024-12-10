package usecase

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository"
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
	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.rdbmsRepo.GetPlaylist()
	if err != nil {
		return err
	}

	current := playlist.GetHead()
	for current != nil {
		if err := uc.cacheRepo.AddSong(current.Song); err != nil {
			return err
		}

		current = current.Next
	}

	currentNode := playlist.GetCurrent()
	if currentNode != nil {
		if err := uc.cacheRepo.SetCurrent(currentNode); err != nil {
			return err // TODO: add additional err handling
		}
	}

	return nil
}

func (uc *PlaylistUseCase) AddSong(title, artist string, duration time.Duration) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	song := &entity.Song{
		Title:    title,
		Artist:   artist,
		Duration: duration,
	}

	if err := uc.rdbmsRepo.AddSong(song); err != nil {
		return ErrAddSongToDB //TODO: Add error description within default Err
	}

	if err := uc.cacheRepo.AddSong(song); err != nil {
		return ErrAddSongToCache
	}

	return nil
}

func (uc *PlaylistUseCase) Play() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.playing && !uc.paused {
		return nil // TODO: Add already playing err/log
	}

	uc.paused = false
	if !uc.playing {
		uc.playing = true
		uc.stopChan = make(chan struct{}, 1)
		go uc.playCurrentSong() // Playback emulation
	} else {
		go uc.playCurrentSong()
	}

	return nil // TODO: add run goroutine playing
}

func (uc *PlaylistUseCase) Pause() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if !uc.playing {
		return ErrNotPlaying
	}
	if uc.paused {
		return ErrAlreadyPaused
	}

	uc.paused = true
	uc.playing = false

	select {
	case uc.stopChan <- struct{}{}:
	default:
	}

	return nil
}

func (uc *PlaylistUseCase) Next() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.cacheRepo.GetPlaylist()
	if err != nil {
		return ErrGetPlaylistFromCache
	}

	current := playlist.GetCurrent()
	if current == nil || current.Next == nil {
		return ErrNoNextSong
	}

	uc.position = 0
	uc.playing = false
	uc.paused = false

	select {
	case uc.stopChan <- struct{}{}:
	default:
	}

	if err := uc.cacheRepo.SetCurrent(current.Next); err != nil {
		return ErrSetCurrentInCache
	}

	if err := uc.rdbmsRepo.SetCurrent(current.Next); err != nil {
		return ErrSetCurrentInDB
	}

	uc.playing = true
	uc.stopChan = make(chan struct{}, 1)
	go uc.playCurrentSong()
	return nil
}

func (uc *PlaylistUseCase) Prev() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.cacheRepo.GetPlaylist()
	if err != nil {
		return ErrGetPlaylistFromCache
	}

	current := playlist.GetCurrent()
	if current == nil || current.Prev == nil {
		return ErrNoPrevSong
	}

	uc.position = 0
	uc.playing = false
	uc.paused = false

	select {
	case uc.stopChan <- struct{}{}:
	default:
	}

	if err := uc.cacheRepo.SetCurrent(current.Prev); err != nil {
		return ErrSetCurrentInCache
	}
	if err := uc.rdbmsRepo.SetCurrent(current.Prev); err != nil {
		return ErrSetCurrentInDB
	}

	uc.playing = true
	uc.stopChan = make(chan struct{}, 1)
	go uc.playCurrentSong()
	return nil
}

func (uc *PlaylistUseCase) GetCurrentSong() (*entity.Song, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	node, err := uc.cacheRepo.GetCurrent()
	if err != nil {
		return nil, ErrGetCurrentNode
	}
	if node == nil || node.Song == nil {
		return nil, ErrNoCurrentSong
	}
	return node.Song, nil
}

func (uc *PlaylistUseCase) GetPlaylist() (*entity.Playlist, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	playlist, err := uc.cacheRepo.GetPlaylist()
	if err != nil {
		return nil, ErrGetPlaylistFromCache
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
			operationLogger.Info("No song to play. Playback stopped.")
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
				operationLogger.Info("No next song. Playback completed.")
				uc.playing = false
				uc.mu.Unlock()
				return
			}
		}

		operationLogger.Info(
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
				operationLogger.Info(
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
			operationLogger.Info("No next song. Playback completed.")
			uc.mu.Unlock()
			uc.playing = false
			return
		}
	}

}
