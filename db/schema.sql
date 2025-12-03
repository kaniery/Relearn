CREATE TABLE IF NOT EXISTS users (
    user_id          INT PRIMARY KEY AUTO_INCREMENT, 
    username         VARCHAR(255) NOT NULL, 
    email            VARCHAR(255) UNIQUE NOT NULL, 
    password_hash    VARCHAR(255) NOT NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login       DATETIME,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS qualifications (
    qualification_id INT PRIMARY KEY AUTO_INCREMENT,
    name             VARCHAR(255) NOT NULL,
    provider         TEXT,
    exam_date        DATETIME,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS topics (
    topic_id         INT PRIMARY KEY AUTO_INCREMENT,
    qualification_id INT NOT NULL,
    name             VARCHAR(255) NOT NULL,
    parent_topic_id  INT,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (qualification_id) REFERENCES qualifications(qualification_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_topic_id)  REFERENCES topics(topic_id)              ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS tags (
    tag_id INT PRIMARY KEY AUTO_INCREMENT,
    name   VARCHAR(255) NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS questions (
    question_id        INT PRIMARY KEY AUTO_INCREMENT,
    qualification_id   INT,
    topic_id           INT,
    author_user_id     INT,
    question_data      TEXT NOT NULL,
    created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (qualification_id) REFERENCES qualifications(qualification_id) ON DELETE SET NULL,
    FOREIGN KEY (topic_id)         REFERENCES topics(topic_id)              ON DELETE SET NULL,
    FOREIGN KEY (author_user_id)   REFERENCES users(user_id)                ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS question_tags (
    question_id INT NOT NULL,
    tag_id      INT NOT NULL,
    PRIMARY KEY (question_id, tag_id),
    FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id)      REFERENCES tags(tag_id)           ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS study_sets (
    set_id          INT PRIMARY KEY AUTO_INCREMENT,
    owner_user_id   INT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    is_public       TINYINT NOT NULL DEFAULT 0, -- INTEGERからTINYINTに変更 (0, 1チェック)
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CHECK (is_public IN (0,1))
);
CREATE TABLE IF NOT EXISTS study_set_questions (
    set_id      INT NOT NULL,
    question_id INT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (set_id, question_id),
    FOREIGN KEY (set_id)      REFERENCES study_sets(set_id)   ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS qualification_enrollments (
    user_id          INT NOT NULL,
    qualification_id INT NOT NULL,
    target_date      DATETIME,
    priority         INT DEFAULT 0,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, qualification_id),
    FOREIGN KEY (user_id)          REFERENCES users(user_id)          ON DELETE CASCADE,
    FOREIGN KEY (qualification_id) REFERENCES qualifications(qualification_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS responses (
    response_id INT PRIMARY KEY AUTO_INCREMENT,
    user_id     INT NOT NULL,
    question_id INT NOT NULL,
    is_correct  TINYINT NOT NULL CHECK (is_correct IN (0,1)), -- INTEGERからTINYINTに変更
    elapsed_ms  INT,
    answered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id)     REFERENCES users(user_id)     ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS reviews (
    user_id          INT NOT NULL,
    question_id      INT NOT NULL,
    last_review_at   DATETIME,
    next_review_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, question_id),
    FOREIGN KEY (user_id)     REFERENCES users(user_id)     ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS attachments (
    attachment_id INT PRIMARY KEY AUTO_INCREMENT,
    question_id   INT,
    kind          VARCHAR(50) NOT NULL, -- TEXTからVARCHARに変更推奨
    url           TEXT NOT NULL,
    meta_json     TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS question_favorites (
    user_id     INT NOT NULL,
    question_id INT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, question_id),
    FOREIGN KEY (user_id)     REFERENCES users(user_id)     ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS user_favorites (
    follower_user_id INT NOT NULL,
    favorite_user_id INT NOT NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_user_id, favorite_user_id),
    FOREIGN KEY (follower_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (favorite_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CHECK (follower_user_id <> favorite_user_id)
);
CREATE TABLE IF NOT EXISTS user_data_hashes (
    user_id     INT PRIMARY KEY,
    hash_value  TEXT NOT NULL,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);