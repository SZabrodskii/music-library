CREATE TABLE songs (
                       id SERIAL PRIMARY KEY,
                       created_at TIMESTAMP NOT NULL,
                       updated_at TIMESTAMP NOT NULL,
                       deleted_at TIMESTAMP,
                       group_name VARCHAR(255) NOT NULL,
                       song_name VARCHAR(255) NOT NULL,
                       release_date VARCHAR(255),
                       link VARCHAR(255)
);

CREATE TABLE verses (
                        id SERIAL PRIMARY KEY,
                        created_at TIMESTAMP NOT NULL,
                        updated_at TIMESTAMP NOT NULL,
                        deleted_at TIMESTAMP,
                        song_id INT NOT NULL,
                        text TEXT NOT NULL,
                        FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
);
