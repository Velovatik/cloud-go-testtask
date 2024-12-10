package usecase

import (
	"cloud-go-testtask/internal/entity"
	"sync"
)

type MockPlaylistRepo struct {
	mu       sync.Mutex
	playlist *entity.Playlist
}

func NewMockPlaylistRepo() *MockPlaylistRepo {
	return &MockPlaylistRepo{
		playlist: &entity.Playlist{},
	}
}

func (m *MockPlaylistRepo) AddSong(song *entity.Song) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playlist.AddToEnd(song)
	return nil
}

func (m *MockPlaylistRepo) GetPlaylist() (*entity.Playlist, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playlist, nil
}

func (m *MockPlaylistRepo) SetCurrent(node *entity.PlaylistNode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playlist.SetCurrent(node)
}

func (m *MockPlaylistRepo) GetCurrent() (*entity.PlaylistNode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playlist.GetCurrent(), nil
}
