package rdbms

import (
	"cloud-go-testtask/internal/entity"
	"cloud-go-testtask/internal/repository"
	"database/sql"
	"errors"
	"time"
)

type PlaylistRepositoryRDBMS struct {
	db                *sql.DB
	defaultPlaylistID int // Может есть способ лучше?...
}

func NewPlaylistRepositoryRDBMS(db *sql.DB) repository.PlaylistRepository {
	return &PlaylistRepositoryRDBMS{db: db}
}

/*
 Methods for Playlist CRUD implementation
*/

func (r *PlaylistRepositoryRDBMS) CreatePlaylist(name, description string) (int, error) {
	var createdPlaylistId int

	err := r.db.QueryRow(
		"INSERT INTO  playlists (name, description) VALUES ($1, $2) RETURNING id",
		name, description).Scan(&createdPlaylistId)

	if err != nil {
		return 0, err
	}

	return createdPlaylistId, nil
}

func (r *PlaylistRepositoryRDBMS) FindPlaylistIDByName(name string) (int, error) {
	var id int
	err := r.db.QueryRow("SELECT id FROM playlists WHERE name = $1", name).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, repository.ErrPlaylistNotFound
		}
		return 0, err
	}
	return id, nil
}

func (r *PlaylistRepositoryRDBMS) SetDefaultPlaylistID(id int) {
	r.defaultPlaylistID = id
}

func (r *PlaylistRepositoryRDBMS) GetPlaylistByID(id int) (*entity.Playlist, error) {

	var currentSongID sql.NullInt64 //int

	err := r.db.QueryRow("SELECT current_song_id FROM playlists WHERE id=$1", id).Scan(&currentSongID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return nil, repository.ErrPlaylistNotFound
		}

		return nil, err
	}

	playlist := entity.Playlist{}

	rows, err := r.db.Query(`
		SELECT s.id, s.title, s.artist, s.duration
		FROM playlist_songs ps
		JOIN songs s ON s.id = ps.song_id
		WHERE ps.playlist_id = $1
		ORDER BY ps.song_order`, id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currentNode *entity.PlaylistNode
	foundSongs := false

	for rows.Next() {
		foundSongs = true
		var s entity.Song
		var durationSec int
		if err := rows.Scan(&s.ID, &s.Title, &s.Artist, &durationSec); err != nil {
			return nil, err
		}
		s.Duration = time.Duration(durationSec) * time.Second
		node := playlist.AddToEnd(&s)

		if currentSongID.Valid && int(currentSongID.Int64) == s.ID {
			currentNode = node
		}
	}

	if !foundSongs {
		return nil, repository.ErrNoSongsInPlaylist
	}

	if currentSongID.Valid && currentNode == nil {
		return nil, repository.ErrCurrentSongNotFound
	}

	if currentNode != nil {
		if err := playlist.SetCurrent(currentNode); err != nil {
			return nil, err
		}
	}

	return &playlist, nil
}

func (r *PlaylistRepositoryRDBMS) UpdatePlaylistCurrentSong(playlistID, songID int) error {
	_, err := r.db.Exec("UPDATE playlists SET current_song_id = $1 WHERE id = $2", songID, playlistID)

	if err != nil {
		return err
	}

	return nil
}

func (r *PlaylistRepositoryRDBMS) DeletePlaylistByID(id int) error {
	_, err := r.db.Exec("DELETE FROM playlists WHERE id = $1", id)

	if err != nil {
		return err
	}

	return nil
}

/*
 Methods for Song CRUD implementation
*/

func (r *PlaylistRepositoryRDBMS) AddSong(song *entity.Song) error {

	duration := int(song.Duration.Seconds())

	_, err := r.db.Exec("INSERT INTO songs (title, artist, duration) VALUES ($1, $2, $3)",
		song.Title, song.Artist, duration)

	if err != nil { // TODO: Add additional err handling
		return repository.ErrAddSong
	}

	return nil
}

func (r *PlaylistRepositoryRDBMS) GetSongByID(id int) (*entity.Song, error) {
	var song entity.Song
	var duration int

	err := r.db.QueryRow("SELECT id, title, artist, duration FROM songs WHERE id = $1", id).Scan(&song.ID, &song.Title, &duration)

	if err != nil { // TODO: Add additional err handling
		return nil, err
	}

	song.Duration = time.Duration(duration) * time.Second

	return &song, nil
}

func (r *PlaylistRepositoryRDBMS) UpdateSong(song *entity.Song) error {
	duration := int(song.Duration.Seconds())

	_, err := r.db.Exec(
		"UPDATE songs SET title = $1, artist = $2, duration = $3 WHERE id = $4",
		song.Title, song.Artist, duration, song.ID)

	if err != nil { // TODO: Add additional err handling
		return err
	}

	return nil
}

func (r *PlaylistRepositoryRDBMS) DeleteSong(id int) error {
	_, err := r.db.Exec("DELETE FROM songs WHERE id = $1", id)

	if err != nil { // TODO: Add additional err handling
		return err
	}
	return nil
}

/*
 Methods for Song-Playlist relatins
*/

func (r *PlaylistRepositoryRDBMS) AddSongToPlaylist(playlistID, songID int) error {
	var maxNumberInPlaylist sql.NullInt64

	err := r.db.QueryRow("SELECT MAX(song_order) FROM playlist_songs WHERE playlist_id = $1",
		playlistID).Scan(&maxNumberInPlaylist)

	if err != nil { // TODO: Add additional err handling
		return err
	}

	newNumber := 1
	if maxNumberInPlaylist.Valid {
		newNumber = int(maxNumberInPlaylist.Int64) + 1
	}

	_, err = r.db.Exec("INSERT INTO playlist_songs (playlist_id, song_id, song_order) VALUES ($1, $2, $3)",
		playlistID, songID, newNumber)

	if err != nil { // TODO: Add additional err handling
		return err
	}

	return nil
}

func (r *PlaylistRepositoryRDBMS) RemoveSongFromPlaylist(playlistID, songID int) error {
	_, err := r.db.Exec(
		"DELETE FROM  playlist_songs WHERE playlist_id = $1 AND song_id = $2",
		playlistID, songID)

	if err != nil { // TODO: Add additional err handling
		return err
	}

	return nil
}

/*
 PlaylistRepository interface methods impl
*/

func (r *PlaylistRepositoryRDBMS) GetPlaylist() (*entity.Playlist, error) {
	if r.defaultPlaylistID == 0 {
		return nil, repository.ErrDefaultPlaylistNotSet
	}

	return r.GetPlaylistByID(r.defaultPlaylistID)
}

func (r *PlaylistRepositoryRDBMS) SetCurrent(node *entity.PlaylistNode) error {
	if r.defaultPlaylistID == 0 {
		return repository.ErrDefaultPlaylistNotSet
	}
	if node == nil || node.Song == nil {
		return repository.ErrNilNode
	}

	return r.UpdatePlaylistCurrentSong(r.defaultPlaylistID, node.Song.ID)

}

func (r *PlaylistRepositoryRDBMS) GetCurrent() (*entity.PlaylistNode, error) {
	if r.defaultPlaylistID == 0 {
		return nil, repository.ErrDefaultPlaylistNotSet
	}

	var currentSongID sql.NullInt64

	err := r.db.QueryRow("SELECT current_song_id FROM playlists WHERE id = $1",
		r.defaultPlaylistID).Scan(&currentSongID)

	if err != nil { //TODO: add additional error handling
		return nil, err
	}

	if !currentSongID.Valid {
		return nil, nil
	}

	song, err := r.GetSongByID(int(currentSongID.Int64))
	if err != nil {
		return nil, err
	}

	node := &entity.PlaylistNode{
		Song: song,
	}
	return node, nil
}
