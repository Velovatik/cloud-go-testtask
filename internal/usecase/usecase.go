package usecase

import "errors"

var (
	ErrAddSongToDB          = errors.New("failed to add song to DB")
	ErrAddSongToCache       = errors.New("failed to add song to cache")
	ErrAlreadyPaused        = errors.New("already paused")
	ErrNotPlaying           = errors.New("not playing")
	ErrGetPlaylistFromCache = errors.New("failed to get playlist from cache")
	ErrNoNextSong           = errors.New("no next song")
	ErrNoPrevSong           = errors.New("no previous song")
	ErrSetCurrentInCache    = errors.New("failed to set current in cache")
	ErrSetCurrentInDB       = errors.New("failed to set current in DB")
	ErrGetCurrentNode       = errors.New("failed to get current node")
	ErrNoCurrentSong        = errors.New("no current song")
)
