-- +goose Up
-- +goose StatementBegin
INSERT INTO playlists (id, name, description)
VALUES (1, 'Default Playlist', 'This is the default playlist')
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM playlists WHERE id = 1;
-- +goose StatementEnd
