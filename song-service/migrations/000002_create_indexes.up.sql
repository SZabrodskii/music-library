-- song-service/migrations/000002_create_indexes.up.sql
CREATE INDEX idx_songs_group_name ON songs (group_name);
CREATE INDEX idx_songs_song_name ON songs (song_name);
CREATE INDEX idx_verses_song_id ON verses (song_id);