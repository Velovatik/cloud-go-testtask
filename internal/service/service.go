package service

import "cloud-go-testtask/internal/entity"

type PlayListServiceInterface interface { // TODO: move interface to place of usage when implemented
	Play() error
	Pause() error
	AddSong(song *entity.Song) error
	Next() error
	Prev() error
}
