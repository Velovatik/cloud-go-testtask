package cache_test

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository/cache"
	"sync"
	"testing"
	"time"
)

func TestAddSongConcurrency(t *testing.T) {
	repo := cache.NewPlaylistRepository()
	numGoroutines := 20
	songsPerGoroutine := 1000
	totalSongs := numGoroutines * songsPerGoroutine

	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	errChan := make(chan error, totalSongs)

	for i := 0; i < numGoroutines; i++ {
		go func(startID int) {
			defer wg.Done()
			for j := 0; j < songsPerGoroutine; j++ {
				song := &entity.Song{
					ID:       startID + j,
					Title:    "Concurrent Song",
					Duration: 3 * time.Minute,
				}
				err := repo.AddSong(song)
				if err != nil {
					errChan <- err
				}
			}
		}(i * songsPerGoroutine)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	playlist, err := repo.GetPlaylist()
	if err != nil {
		t.Fatalf("failed to get playlist: %v", err)
	}

	node := playlist.GetHead()
	count := 0
	for node != nil {
		count++
		node = node.Next
	}

	if count != totalSongs {
		t.Errorf("unexpected song count: got %d, want %d", count, totalSongs)
	}
}
