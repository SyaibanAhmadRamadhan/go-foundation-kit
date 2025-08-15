CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users (name)
SELECT 'user-' || g
FROM generate_series(1, 10000) AS g;
-- sebuah query yang akan sering dipakai di load test
CREATE INDEX IF NOT EXISTS idx_users_name ON users (name);