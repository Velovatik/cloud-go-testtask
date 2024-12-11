package cache

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository"
	"sync"
)

/*
PlaylistRepositoryCache is an imitation of repository realization with linked list and synchronization
AddSong is a method of PlaylistRepository to add a node to linked list
*/
type PlaylistRepositoryCache struct {
	mu       sync.RWMutex
	playlist *entity.Playlist
}

func NewPlaylistRepositoryCache() *PlaylistRepositoryCache {
	return &PlaylistRepositoryCache{
		playlist: &entity.Playlist{},
	}
}

func (r *PlaylistRepositoryCache) AddSong(song *entity.Song) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if song == nil {
		return repository.ErrNullSong
	}

	r.playlist.AddToEnd(song)

	return nil

}

func (r *PlaylistRepositoryCache) GetPlaylist() (*entity.Playlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.playlist == nil {
		return nil, repository.ErrPlaylistNotInitialized
	}

	return r.playlist, nil
}

func (r *PlaylistRepositoryCache) SetCurrent(node *entity.PlaylistNode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.playlist.SetCurrent(node)
}

func (r *PlaylistRepositoryCache) GetCurrent() (*entity.PlaylistNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.playlist.GetCurrent(), nil
}
