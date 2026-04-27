CREATE TABLE IF NOT EXISTS members (
	id INT PRIMARY KEY AUTO_INCREMENT,
	name VARCHAR(255) NOT NULL,
	telegram_username VARCHAR(100),
	telegram_chat_id VARCHAR(100),
	phone_number VARCHAR(20),
	joined_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	is_active BOOLEAN DEFAULT TRUE,
	INDEX idx_telegram_chat_id (telegram_chat_id),
	UNIQUE INDEX idx_members_telegram_username (telegram_username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
