package usecase

import (
	"cloud-go-testtask/internal/entity"
	"fmt"
	"log/slog"
	"time"
)

func (uc *PlaylistUseCase) AddSong(title, artist string, duration time.Duration) error {
	const op = "usecase.SongUseCase.AddSong"
	operationLogger := uc.logger.With(slog.String("op", op))

	operationLogger.Debug("Adding new song",
		slog.String("title", title),
		slog.String("artist", artist),
		slog.Duration("duration", duration),
	)

	uc.mu.Lock()
	defer uc.mu.Unlock()

	song := &entity.Song{
		Title:    title,
		Artist:   artist,
		Duration: duration,
	}

	if err := uc.rdbmsRepo.AddSong(song); err != nil {
		operationLogger.Error("Failed to add song to DB",
			slog.String("title", title),
			slog.String("artist", artist),
			slog.Duration("duration", duration),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", ErrAddSongToDB, err)
	}

	if err := uc.cacheRepo.AddSong(song); err != nil {
		operationLogger.Error("Failed to add song to Cache",
			slog.String("title", title),
			slog.String("artist", artist),
			slog.Duration("duration", duration),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", ErrAddSongToCache, err)
	}

	operationLogger.Debug("Song added successfully",
		slog.String("title", title),
		slog.String("artist", artist),
	)

	return nil
}
