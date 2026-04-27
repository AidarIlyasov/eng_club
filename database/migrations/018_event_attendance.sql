CREATE TABLE IF NOT EXISTS event_attendance (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	event_id BIGINT NOT NULL,
	member_id INT NOT NULL,
	table_id INT NULL,
	attended BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	UNIQUE KEY unique_event_member (event_id, member_id),
	INDEX idx_event_attendance_event (event_id),
	INDEX idx_event_attendance_member (member_id),
	INDEX idx_event_attendance_table (table_id),
	FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
	FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
