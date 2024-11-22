-- song-service/migrations/000002_create_indexes.down.sql
DROP INDEX idx_songs_group_name;
DROP INDEX idx_songs_song_name;
DROP INDEX idx_verses_song_id;