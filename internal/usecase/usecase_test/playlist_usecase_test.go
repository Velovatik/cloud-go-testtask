package usecase

import (
	"log/slog"
	"testing"
	"time"

	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/usecase"
)

func TestPlayCurrentSongUnderLoad(t *testing.T) {
	rdbmsRepo := NewMockPlaylistRepo()
	cacheRepo := NewMockPlaylistRepo()

	logger := slog.Default()

	uc := usecase.NewPlaylistUseCase(rdbmsRepo, cacheRepo, logger)

	songs := []*entity.Song{
		{ID: 1, Title: "Song1", Artist: "Artist1", Duration: 5 * time.Second},
		{ID: 2, Title: "Song2", Artist: "Artist2", Duration: 5 * time.Second},
		{ID: 3, Title: "Song3", Artist: "Artist3", Duration: 5 * time.Second},
	}

	for _, s := range songs {
		if err := rdbmsRepo.AddSong(s); err != nil {
			t.Fatalf("failed to add song to rdbms: %v", err)
		}
	}

	for _, s := range songs {
		if err := cacheRepo.AddSong(s); err != nil {
			t.Fatalf("Failed to add song to cache: %v", err)
		}
	}

	pl, _ := cacheRepo.GetPlaylist()
	if err := cacheRepo.SetCurrent(pl.GetHead()); err != nil {
		t.Fatalf("failed to set current in cache: %v", err)
	}
	if err := rdbmsRepo.SetCurrent(pl.GetHead()); err != nil {
		t.Fatalf("failed to set current in rdbms: %v", err)
	}

	// Запустим воспроизведение
	if err := uc.Play(); err != nil {
		t.Fatalf("failed to start play: %v", err)
	}

	const goroutines = 10
	const iterations = 20

	done := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				switch j % 3 {
				case 0:
					uc.Pause()
				case 1:
					uc.Next()
				case 2:
					uc.Prev()
				}

				time.Sleep(100 * time.Millisecond)
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	time.Sleep(2 * time.Second)

	if _, err := uc.GetCurrentSong(); err != nil {
		t.Logf("Could not get current song at the end: %v", err)
	}

	t.Log("Load test completed successfully")
}
