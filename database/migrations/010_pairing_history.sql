CREATE TABLE IF NOT EXISTS pairing_history (
	member_id INT NOT NULL,
	pair_id INT NOT NULL,
	pairing_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (member_id, pair_id),
	INDEX idx_pairing_date (pairing_date),
	FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
	FOREIGN KEY (pair_id) REFERENCES members(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
