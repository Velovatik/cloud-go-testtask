package entity

type PlaylistNode struct {
	Song *Song
	Prev *PlaylistNode
	Next *PlaylistNode
}

type Playlist struct {
	head    *PlaylistNode
	current *PlaylistNode
	tail    *PlaylistNode
}

func (p *Playlist) GetCurrent() *PlaylistNode {
	return p.current
}

func (p *Playlist) SetCurrent(node *PlaylistNode) error {
	if node == nil {
		return ErrNullEntity
	}
	p.current = node
	return nil
}

func (p *Playlist) GetTail() *PlaylistNode {
	return p.tail
}

func (p *Playlist) SetTail(node *PlaylistNode) {
	p.tail = node
}

func (p *Playlist) GetHead() *PlaylistNode {
	return p.head
}

func (p *Playlist) SetHead(node *PlaylistNode) {
	p.head = node
	if p.current == nil {
		p.current = node
	}
}

func (p *Playlist) AddToEnd(song *Song) *PlaylistNode {
	node := &PlaylistNode{Song: song}
	if p.tail == nil {
		p.head = node
		p.tail = node
		p.current = node
	} else {
		node.Prev = p.tail
		p.tail.Next = node
		p.tail = node
	}
	return node
}

func (p *Playlist) RemoveSong(songID int) error {
	node := p.head
	for node != nil {
		if node.Song.ID == songID {
			if node.Prev != nil {
				node.Prev.Next = node.Next
			} else {
				p.head = node.Next
			}
			if node.Next != nil {
				node.Next.Prev = node.Prev
			} else {
				p.tail = node.Prev
			}
			if p.current == node {
				p.current = node.Next
			}
			return nil
		}
		node = node.Next
	}
	return ErrSongNotFound
}
