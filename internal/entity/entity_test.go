package entity_test

import (
	"cloud-go-testtask/internal/entity"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetCurrent(t *testing.T) {
	playlist := &entity.Playlist{}

	node := &entity.PlaylistNode{
		Song: &entity.Song{
			Title:    "Song for test",
			Duration: 3 * time.Second,
		},
	}

	err := playlist.SetCurrent(node)
	assert.NoError(t, err)
	assert.Equal(t, node, playlist.GetCurrent())
}

func TestSetCurrentNil(t *testing.T) {
	playlist := &entity.Playlist{}

	err := playlist.SetCurrent(nil)
	assert.Error(t, err)
	assert.Equal(t, entity.ErrNullEntity, err)
}
