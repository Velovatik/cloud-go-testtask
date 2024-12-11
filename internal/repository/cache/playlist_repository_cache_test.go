package cache_test

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository"
	"cloud-go-testtask/internal/repository/cache"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddSong(t *testing.T) {
	repo := cache.NewPlaylistRepositoryCache()

	testSong := &entity.Song{
		ID:       1,
		Title:    "Song for test",
		Duration: 10 * time.Second,
	}

	err := repo.AddSong(testSong)
	assert.NoError(t, err)

	playlist, err := repo.GetPlaylist()
	assert.NoError(t, err)
	assert.NotNil(t, playlist)

	assert.Equal(t, testSong, playlist.GetHead().Song)
	assert.Equal(t, testSong, playlist.GetTail().Song)
}

func TestAddNilSong(t *testing.T) {
	repo := cache.NewPlaylistRepositoryCache()

	err := repo.AddSong(nil)

	assert.Error(t, err)
	assert.Equal(t, repository.ErrNullSong, err)
}

func TestPlaylistNotInitialized(t *testing.T) {
	repo := &cache.PlaylistRepositoryCache{}

	_, err := repo.GetPlaylist()
	assert.Error(t, err)
	assert.Equal(t, repository.ErrPlaylistNotInitialized, err)
}
