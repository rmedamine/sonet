CREATE TABLE IF NOT EXISTS group_post_reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    userId INTEGER NOT NULL,
    groupPostId INTEGER NOT NULL,
    reaction_type TEXT CHECK (reaction_type IN ('LIKE')),
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (groupPostId) REFERENCES group_posts(id) ON DELETE CASCADE,
    UNIQUE(userId, groupPostId)
);