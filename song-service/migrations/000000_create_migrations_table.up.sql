-- song-service/migrations/000000_create_migrations_table.up.sql
CREATE TABLE migrations (
                            id SERIAL PRIMARY KEY,
                            version VARCHAR(255) NOT NULL,
                            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);