CREATE TABLE IF NOT EXISTS follows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    followerId INTEGER NOT NULL,
    followingId INTEGER NOT NULL,
    followerName TEXT NOT NULL,
    followingName TEXT NOT NULL,
    accepted TEXT CHECK (accepted IN ('pending', 'accepted', 'declined')) DEFAULT 'pending',
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (followerId) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (followingId) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (followerId, followingId)
);
