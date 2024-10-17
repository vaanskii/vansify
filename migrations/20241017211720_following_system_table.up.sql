CREATE TABLE followers (
    follower_username VARCHAR(255) NOT NULL,
    following_username VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_username, following_username),
    FOREIGN KEY (follower_username) REFERENCES users(username) ON DELETE CASCADE,
    FOREIGN KEY (following_username) REFERENCES users(username) ON DELETE CASCADE
);