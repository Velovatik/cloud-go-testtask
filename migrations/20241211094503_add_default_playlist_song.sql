-- +goose Up
-- +goose StatementBegin
INSERT INTO songs (id, title, artist, duration)
VALUES
    (1, 'Default Song 1', 'Default Artist 1', 300),
    (2, 'Default Song 2', 'Default Artist 2', 250)
ON CONFLICT (id) DO NOTHING;

INSERT INTO playlist_songs (playlist_id, song_id, song_order)
VALUES
    (1, 1, 1),
    (1, 2, 2)
ON CONFLICT (playlist_id, song_order) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM playlist_songs WHERE playlist_id = 1 AND song_id IN (1, 2);
DELETE FROM songs WHERE id IN (1, 2);
-- +goose StatementEnd
