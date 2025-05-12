CREATE TABLE posts_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    userId INTEGER NOT NULL,
    avatar TEXT,
    content TEXT NOT NULL,
    image TEXT,
    privacy INTEGER NOT NULL DEFAULT 0,
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO posts_new (id, title, userId, avatar, content, image, privacy, createdAt)
SELECT id, title, userId, avatar, content, image, private, createdAt FROM posts;

DROP TABLE posts;

ALTER TABLE posts_new RENAME TO posts;