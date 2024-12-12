package repository

import (
	"cloud-go-testtask/internal/entity"
	"errors"
)

var (
	ErrNullSong               = errors.New("received null instead of Song struct")
	ErrPlaylistNotInitialized = errors.New("playlist is not initialized")
	ErrPlaylistNotFound       = errors.New("playlist not found")
	ErrNoSongsInPlaylist      = errors.New("no songs found in playlist")
	ErrCurrentSongNotFound    = errors.New("current song not found in playlist")
	ErrDefaultPlaylistNotSet  = errors.New("default playlist not set")
	ErrNilNode                = errors.New("node or node.Song is nil")
	ErrAddSong                = errors.New("failed to add song")
	ErrPlaylistCreationFailed = errors.New("playlist creation failed")
)

type PlaylistRepository interface {
	GetPlaylist() (*entity.Playlist, error)
	AddSong(song *entity.Song) error
	SetCurrent(node *entity.PlaylistNode) error
	GetCurrent() (*entity.PlaylistNode, error)
}
