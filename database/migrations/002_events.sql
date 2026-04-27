CREATE TABLE IF NOT EXISTS events (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	place_id BIGINT NOT NULL,
	scheduled_at DATETIME NOT NULL,
	topic VARCHAR(512) NOT NULL DEFAULT '',
	notify_at DATETIME NULL,
	status ENUM('planned', 'started', 'completed') NOT NULL DEFAULT 'planned',
	started_at DATETIME NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	finished_at DATETIME NULL,
	INDEX idx_events_scheduled (scheduled_at),
	INDEX idx_events_status (status),
	INDEX idx_events_place (place_id),
	INDEX idx_events_notify (notify_at),
	INDEX idx_events_finished (finished_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
