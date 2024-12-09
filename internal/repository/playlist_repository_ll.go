package repository

import (
	"cloud-go-testtask/internal/entity"
	"sync"
)

/*
PlaylistRepository is an imitation of repository realization with linked list and synchronization
AddSong is a method of PlaylistRepository to add a node to linked list
*/
type PlaylistRepository struct {
	playlist      *entity.Playlist
	playlistMutex sync.Mutex
}

func NewPlaylistRepository() *PlaylistRepository {
	return &PlaylistRepository{
		playlist: &entity.Playlist{},
	}
}

func (r *PlaylistRepository) AddSong(song *entity.Song) error {
	r.playlistMutex.Lock()
	defer r.playlistMutex.Unlock()

	if song == nil {
		return ErrNullSong
	}

	newNode := &entity.PlaylistNode{
		Song: song,
	}

	tail := r.playlist.GetTail()
	if tail == nil {
		r.playlist.SetHead(newNode)
		r.playlist.SetTail(newNode)
		if err := r.playlist.SetCurrent(newNode); err != nil {
			return err
		}
	} else {
		newNode.Prev = tail
		tail.Next = newNode
		r.playlist.SetTail(newNode)
	}

	return nil

}

func (r *PlaylistRepository) GetPlaylist() (*entity.Playlist, error) {
	r.playlistMutex.Lock()
	defer r.playlistMutex.Unlock()

	if r.playlist == nil {
		return nil, ErrPlaylistNotInitialized
	}

	return r.playlist, nil
}
