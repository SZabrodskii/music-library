CREATE TABLE songs (
                       id SERIAL PRIMARY KEY,
                       group_name VARCHAR(255) NOT NULL,
                       song_name VARCHAR(255) NOT NULL,
                       release_date VARCHAR(255),
                       link VARCHAR(255)
);

CREATE TABLE verses (
                        id SERIAL PRIMARY KEY,
                        song_id INT NOT NULL,
                        text TEXT NOT NULL,
                        FOREIGN KEY (song_id) REFERENCES songs(id)
);

CREATE INDEX idx_songs_group_name ON songs (group_name);
CREATE INDEX idx_songs_song_name ON songs (song_name);
CREATE INDEX idx_verses_song_id ON verses (song_id);