-- +goose Up
CREATE TABLE songs (
                       id SERIAL PRIMARY KEY,
                       title VARCHAR(255) NOT NULL,
                       artist VARCHAR(255) NOT NULL,
                       duration INT NOT NULL
);

CREATE TABLE playlists (
                           id SERIAL PRIMARY KEY,
                           name VARCHAR(255) NOT NULL,
                           description TEXT,
                           current_song_id INT,
                           created_at TIMESTAMP DEFAULT now(),
                           FOREIGN KEY (current_song_id) REFERENCES songs(id)
);

CREATE TABLE playlist_songs (
                                id SERIAL PRIMARY KEY,
                                playlist_id INT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
                                song_id INT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
                                song_order INT NOT NULL,
                                UNIQUE (playlist_id, song_order),
                                UNIQUE (playlist_id, song_id)
);

-- +goose Down
DROP TABLE playlist_songs;
DROP TABLE playlists;
DROP TABLE songs;
