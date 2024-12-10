package service_test

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository/cache"
	"cloud-go-testtask/internal/service"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPlaylistService_Concurrency(t *testing.T) {
	repo := cache.NewPlaylistRepository()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	PlaylistService := service.NewPlaylistService(repo, logger)

	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			song := &entity.Song{ID: id, Title: fmt.Sprintf("Song %d", id), Duration: 5 * time.Second}
			assert.NoError(t, PlaylistService.AddSong(song))
		}(i)
	}

	wg.Wait()

	playlist, err := repo.GetPlaylist()
	assert.NoError(t, err)

	count := 0
	node := playlist.GetHead()
	for node != nil {
		count++
		node = node.Next
	}
	assert.Equal(t, 10, count)
}

func TestAddSong(t *testing.T) {
	repo := cache.NewPlaylistRepository()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := service.NewPlaylistService(repo, logger)

	song := &entity.Song{ID: 1, Title: "Test Song", Duration: 5 * time.Second}
	err := service.AddSong(song)
	assert.NoError(t, err)

	playlist, err := repo.GetPlaylist()
	assert.NoError(t, err)
	assert.NotNil(t, playlist.GetHead())
	assert.Equal(t, song, playlist.GetHead().Song)
}
