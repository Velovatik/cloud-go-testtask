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

func (r *PlaylistRepository) AddSong(song *entity.Song) {
	r.playlistMutex.Lock()
	defer r.playlistMutex.Unlock()

	newNode := &entity.PlaylistNode{
		Song: song,
	}

	if r.playlist.Tail == nil {
		r.playlist.Head = newNode
		r.playlist.Tail = newNode
		r.playlist.Current = newNode
	} else {
		newNode.Prev = r.playlist.Tail
		r.playlist.Tail.Next = newNode
		r.playlist.Tail = newNode
	}

}

func (r *PlaylistRepository) GetPlaylist() *entity.Playlist {
	return r.playlist
}
