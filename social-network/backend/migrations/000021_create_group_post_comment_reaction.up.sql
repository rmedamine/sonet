CREATE TABLE IF NOT EXISTS group_comment_reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    userId INTEGER NOT NULL,
    commentId INTEGER NOT NULL,
    reaction_type TEXT CHECK (reaction_type IN ('LIKE')),
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (commentId) REFERENCES group_comments(id) ON DELETE CASCADE,
    UNIQUE(userId, commentId)
);