package entity

import (
	"sync"
)

type PlaylistNode struct {
	Song *Song
	Prev *PlaylistNode
	Next *PlaylistNode
}

type Playlist struct {
	mu       sync.Mutex
	head     *PlaylistNode
	current  *PlaylistNode
	tail     *PlaylistNode
	stopChan chan struct{}
}

func (p *Playlist) GetCurrent() *PlaylistNode {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

func (p *Playlist) SetCurrent(node *PlaylistNode) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if node == nil {
		return ErrNullEntity
	}

	p.current = node

	return nil
}

func (p *Playlist) GetTail() *PlaylistNode {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.tail
}

func (p *Playlist) SetTail(node *PlaylistNode) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tail = node
}

func (p *Playlist) GetHead() *PlaylistNode {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.head
}

func (p *Playlist) SetHead(node *PlaylistNode) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.head = node
	if p.current == nil {
		p.current = node
	}
}
