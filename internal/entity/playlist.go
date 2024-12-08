package entity

import "sync"

type PlaylistNode struct {
	Song *Song
	Prev *PlaylistNode
	Next *PlaylistNode
}

type Playlist struct {
	mu       sync.Mutex
	Head     *PlaylistNode
	Current  *PlaylistNode
	Tail     *PlaylistNode
	stopChan chan struct{}
}

//type PlayList interface {
//	Play() error
//	Pause() error
//	AddSong(song Song) error
//	Next() error
//	Prev() error
//}
