package repository

import "errors"

var (
	ErrNullSong               = errors.New("received null instead of Song struct")
	ErrPlaylistNotInitialized = errors.New("playlist is not initialized")
)
