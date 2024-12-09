package repository_test

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddSong(t *testing.T) {
	repo := repository.NewPlaylistRepository()

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
	repo := repository.NewPlaylistRepository()

	err := repo.AddSong(nil)

	assert.Error(t, err)
	assert.Equal(t, repository.ErrNullSong, err)
}

func TestPlaylistNotInitialized(t *testing.T) {
	repo := &repository.PlaylistRepository{}

	_, err := repo.GetPlaylist()
	assert.Error(t, err)
	assert.Equal(t, repository.ErrPlaylistNotInitialized, err)
}
