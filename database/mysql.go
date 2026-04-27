package database

import (
	"database/sql"
	"embed"
	"eng_club/models"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed migrations/*.sql
var migrationSQL embed.FS

// MySQLDatabaseManager handles all MySQL database operations
type MySQLDatabaseManager struct {
	db *sql.DB
}

// MySQLConfig holds MySQL connection configuration
type MySQLConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Charset  string
}

// NewMySQLDatabaseManager creates and initializes a new MySQL database connection.
// It does not run schema migrations; run those once with RunMigrations or `go run ./cmd/migrate`.
func NewMySQLDatabaseManager(config MySQLConfig) (*MySQLDatabaseManager, error) {
	db, err := OpenMySQL(config)
	if err != nil {
		return nil, err
	}
	return &MySQLDatabaseManager{db: db}, nil
}

func OpenMySQL(config MySQLConfig) (*sql.DB, error) {
	dsn := mysqlDSN(config)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

func mysqlDSN(config MySQLConfig) string {
	dialHost := strings.TrimSpace(config.Host)
	if dialHost == "" {
		dialHost = "localhost"
	}
	// dialHost is where this process connects (e.g. localhost). MySQL errors like
	// Access denied for 'user'@'%' refer to the mysql.user account row that matched
	// ('%' = any client host), not to this DSN host field.
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Europe%%2FMoscow",
		config.Username,
		config.Password,
		dialHost,
		config.Port,
		config.Database,
		config.Charset,
	)
}

const migrationsDir = "migrations"

// RunMigrations applies all embedded SQL files in database/migrations in sorted order.
// Use this from cmd/migrate or before deploying; it is not called by NewMySQLDatabaseManager.
func RunMigrations(config MySQLConfig) error {
	db, err := sql.Open("mysql", mysqlDSN(config))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return migrateEmbedded(db)
}

func migrateEmbedded(db *sql.DB) error {
	entries, err := migrationSQL.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".sql") {
			names = append(names, n)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		path := migrationsDir + "/" + name
		data, err := migrationSQL.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		stmt := strings.TrimSpace(string(data))
		if stmt == "" {
			continue
		}
		// Foreign-key ALTERs may already exist after a partial or repeat run (errno 121, etc.).
		if strings.Contains(name, "_fk_") {
			_, _ = db.Exec(stmt)
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

func (dm *MySQLDatabaseManager) GetDB() *sql.DB {
	return dm.db
}

// AddMember adds a new member to the database
func (dm *MySQLDatabaseManager) AddMember(id int, name, telegramUsername, telegramChatID, phoneNumber string) error {
	query := `
		INSERT INTO members (id, name, telegram_username, telegram_chat_id, phone_number)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			name = VALUES(name),
			telegram_username = VALUES(telegram_username),
			telegram_chat_id = VALUES(telegram_chat_id),
			phone_number = VALUES(phone_number)
	`
	_, err := dm.db.Exec(query, id, name, telegramUsername, telegramChatID, phoneNumber)
	return err
}

// GetPairingCount gets the number of times two members have been paired
func (dm *MySQLDatabaseManager) GetPairingCount(member1ID, member2ID int) (int, error) {
	if member1ID > member2ID {
		member1ID, member2ID = member2ID, member1ID
	}

	query := `
		SELECT COUNT(*) FROM pairing_history
		WHERE member1_id = ? AND member2_id = ?
	`
	var count int
	err := dm.db.QueryRow(query, member1ID, member2ID).Scan(&count)
	return count, err
}

// SaveTableAssignment saves where a member sits for an activity
func (dm *MySQLDatabaseManager) SaveTableAssignment(sessionID, activityID int64, memberID, tableID int) error {
	query := `
		INSERT INTO table_assignments (session_id, activity_id, member_id, table_id)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE table_id = VALUES(table_id)
	`
	_, err := dm.db.Exec(query, sessionID, activityID, memberID, tableID)
	return err
}

// GetActivityInfo retrieves activity type ID and name for a given activity ID
func (dm *MySQLDatabaseManager) GetActivityInfo(activityID int64) (activityTypeID int, activityName string, err error) {
	query := `
		SELECT sa.activity_type_id, at.name
		FROM session_activities sa
		JOIN activity_types at ON sa.activity_type_id = at.id
		WHERE sa.id = ?
	`
	err = dm.db.QueryRow(query, activityID).Scan(&activityTypeID, &activityName)
	return
}

// GetAllPairingHistory gets complete pairing history
func (dm *MySQLDatabaseManager) GetAllPairingHistory() (map[string]int, error) {
	query := `
		SELECT member1_id, member2_id, COUNT(*) as count
		FROM pairing_history
		GROUP BY member1_id, member2_id
	`
	rows, err := dm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make(map[string]int)
	for rows.Next() {
		var m1, m2, count int
		if err := rows.Scan(&m1, &m2, &count); err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%d-%d", m1, m2)
		history[key] = count
	}
	return history, nil
}

// Close closes the database connection
func (dm *MySQLDatabaseManager) Close() error {
	return dm.db.Close()
}

// CreateEventActivity creates a new activity for an event
func (dm *MySQLDatabaseManager) CreateEventActivity(eventID int64, name string, durationMinutes, groupSize int) error {
	_, err := dm.db.Exec(
		`INSERT INTO activities (event_id, name, duration_minutes, group_size) VALUES (?, ?, ?, ?)`,
		eventID, name, durationMinutes, groupSize,
	)
	return err
}

// GetEventActivities returns activities for an event
func (dm *MySQLDatabaseManager) GetEventActivities(eventID int64) ([]map[string]interface{}, error) {
	rows, err := dm.db.Query(`
		SELECT 
			id, 
			name, 
			duration_minutes,
			group_size,
			created_at
		FROM activities
		WHERE event_id = ?
		ORDER BY id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var durationMinutes int
		var groupSize int
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &durationMinutes, &groupSize, &createdAt); err != nil {
			return nil, err
		}

		activities = append(activities, map[string]interface{}{
			"id":               id,
			"name":             name,
			"duration_minutes": durationMinutes,
			"group_size":       groupSize,
			"created_at":       createdAt,
		})
	}

	return activities, rows.Err()
}

type Member struct {
}

// UpsertMember inserts or updates a member by telegram_username and returns the member id.
func (dm *MySQLDatabaseManager) UpsertMember(name, telegramUsername, telegramChatID string, phoneNumber *string) (int, error) {
	tu := normalizeTelegram(telegramUsername)
	if tu == "" {
		return 0, fmt.Errorf("telegram username is required")
	}
	query := `
		INSERT INTO members (name, telegram_username, telegram_chat_id, phone_number)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			name = VALUES(name),
			telegram_chat_id = VALUES(telegram_chat_id),
			phone_number = VALUES(phone_number)
	`
	res, err := dm.db.Exec(query, name, tu, nullIfEmpty(telegramChatID), phoneNumber)
	if err != nil {
		return 0, err
	}
	lid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if lid == 0 {
		return 0, fmt.Errorf("upsert member: could not resolve id for %q", telegramUsername)
	}
	return int(lid), nil
}

func normalizeTelegram(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "@") {
		s = "@" + s
	}
	return s
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func (dm *MySQLDatabaseManager) ListMembers() ([]models.Member, error) {
	rows, err := dm.db.Query(`
		SELECT 
			id, 
			name, 
			telegram_username
		FROM members`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member

	for rows.Next() {
		var member models.Member
		err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.Telegram,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

// ListPlaces returns all places ordered by id.
func (dm *MySQLDatabaseManager) ListPlaces() ([]models.Place, error) {
	rows, err := dm.db.Query(`
		SELECT id, name, metro_area, map_url, image_url, created_at, updated_at
		FROM places ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Place
	for rows.Next() {
		var p models.Place
		if err := rows.Scan(&p.ID, &p.Name, &p.MetroArea, &p.MapURL, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetPlace returns a place by id.
func (dm *MySQLDatabaseManager) GetPlace(id int64) (*models.Place, error) {
	var p models.Place
	err := dm.db.QueryRow(`
		SELECT id, name, metro_area, map_url, image_url, created_at, updated_at
		FROM places WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.MetroArea, &p.MapURL, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreatePlace inserts a place and returns it with id set.
func (dm *MySQLDatabaseManager) CreatePlace(name, metroArea, mapURL, imageURL string) (*models.Place, error) {
	res, err := dm.db.Exec(`
		INSERT INTO places (name, metro_area, map_url, image_url)
		VALUES (?, ?, ?, ?)`, name, metroArea, mapURL, imageURL)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return dm.GetPlace(id)
}

// UpdatePlace updates a place by id.
func (dm *MySQLDatabaseManager) UpdatePlace(id int64, name, metroArea, mapURL, imageURL string) (*models.Place, error) {
	res, err := dm.db.Exec(`
		UPDATE places SET name = ?, metro_area = ?, map_url = ?, image_url = ?
		WHERE id = ?`, name, metroArea, mapURL, imageURL, id)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	return dm.GetPlace(id)
}

// DeletePlace removes a place by id.
func (dm *MySQLDatabaseManager) DeletePlace(id int64) error {
	res, err := dm.db.Exec(`DELETE FROM places WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListEvents returns events with place information and optional filters.
// dateFrom: filter by scheduled_at >= dateFrom (empty = no filter)
// dateTo: filter by scheduled_at <= dateTo (empty = no filter)
// placeID: filter by place_id (empty = no filter)
// status: filter by status (empty = no filter)
// limit: max number of results (0 = no limit)
// offset: skip first N results (0 = no offset)
func (dm *MySQLDatabaseManager) ListEvents(dateFrom, dateTo, placeID, status string, limit, offset int) ([]models.Event, error) {
	query := `
		SELECT 
			e.id, e.place_id, e.scheduled_at, e.topic, e.notify_at, 
			e.status, e.type, e.started_at, e.created_at, e.finished_at,
			p.name, p.metro_area, p.map_url, p.image_url
		FROM events e
		JOIN places p ON e.place_id = p.id
		WHERE 1=1`

	args := []interface{}{}

	if dateFrom != "" {
		query += " AND e.scheduled_at >= ?"
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		query += " AND e.scheduled_at <= ?"
		// Add end of day
		args = append(args, dateTo+" 23:59:59")
	}

	if placeID != "" {
		query += " AND e.place_id = ?"
		args = append(args, placeID)
	}

	if status != "" {
		query += " AND e.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY e.scheduled_at ASC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := dm.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEventsWithPlace(rows)
}

func scanEventsWithPlace(rows *sql.Rows) ([]models.Event, error) {
	var out []models.Event
	for rows.Next() {
		var e models.Event
		var notifyAt sql.NullTime
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		var placeName, placeMetroArea string
		var placeMapURL, placeImageURL sql.NullString

		if err := rows.Scan(
			&e.ID, &e.PlaceID, &e.ScheduledAt, &e.Topic, &notifyAt,
			&e.Status, &e.Type, &startedAt, &e.CreatedAt, &finishedAt,
			&placeName, &placeMetroArea, &placeMapURL, &placeImageURL); err != nil {
			return nil, err
		}

		if notifyAt.Valid {
			e.NotifyAt = &notifyAt.Time
		}
		if startedAt.Valid {
			e.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			e.FinishedAt = &finishedAt.Time
		}

		// Build nested place object
		e.Place = &models.Place{
			ID:        e.PlaceID,
			Name:      placeName,
			MetroArea: placeMetroArea,
		}
		if placeMapURL.Valid {
			e.Place.MapURL = placeMapURL.String
		}
		if placeImageURL.Valid {
			e.Place.ImageURL = placeImageURL.String
		}

		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEvent returns an event by id.
func (dm *MySQLDatabaseManager) GetEvent(id int64) (*models.Event, error) {
	var e models.Event
	var notifyAt sql.NullTime
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	err := dm.db.QueryRow(`
		SELECT id, place_id, scheduled_at, topic, notify_at, status, type, started_at, created_at, finished_at
		FROM events WHERE id = ?`, id).Scan(
		&e.ID, &e.PlaceID, &e.ScheduledAt, &e.Topic, &notifyAt, &e.Status, &e.Type, &startedAt, &e.CreatedAt, &finishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if notifyAt.Valid {
		e.NotifyAt = &notifyAt.Time
	}
	if startedAt.Valid {
		e.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		e.FinishedAt = &finishedAt.Time
	}
	return &e, nil
}

// CreateEvent inserts an event.
func (dm *MySQLDatabaseManager) CreateEvent(placeID int64, scheduledAt time.Time, topic string, eventType string, finishedAt *time.Time, notifyAt *time.Time) (*models.Event, error) {
	// Default to many_activities if type is not specified
	if eventType == "" {
		eventType = "many_activities"
	}

	res, err := dm.db.Exec(`
		INSERT INTO events (place_id, scheduled_at, topic, notify_at, status, type, finished_at)
		VALUES (?, ?, ?, ?, 'planned', ?, ?)`, placeID, scheduledAt, topic, notifyAt, eventType, finishedAt)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return dm.GetEvent(id)
}

// UpdateEvent updates an event (not status — use MarkEventStarted for that).
func (dm *MySQLDatabaseManager) UpdateEvent(id int64, placeID int64, scheduledAt time.Time, topic string, notifyAt *time.Time, eventType string, finishedAt *time.Time) (*models.Event, error) {
	res, err := dm.db.Exec(`
		UPDATE events SET place_id = ?, scheduled_at = ?, topic = ?, notify_at = ?, type = ?, finished_at = ?
		WHERE id = ? AND status = 'planned'`, placeID, scheduledAt, topic, notifyAt, eventType, finishedAt, id)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	return dm.GetEvent(id)
}

// DeleteEvent removes a planned event.
func (dm *MySQLDatabaseManager) DeleteEvent(id int64) error {
	res, err := dm.db.Exec(`DELETE FROM events WHERE id = ? AND status = 'planned'`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MarkEventStarted sets status for an event. Returns false if no row was updated (e.g. already started).
func (dm *MySQLDatabaseManager) MarkEventStarted(eventID int64) (bool, error) {
	res, err := dm.db.Exec(`
		UPDATE events SET status = 'started', started_at = NOW() WHERE id = ? AND status = 'planned'`,
		eventID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// MarkEventCompleted sets status to completed for an event.
func (dm *MySQLDatabaseManager) MarkEventCompleted(eventID int64) error {
	_, err := dm.db.Exec(`
		UPDATE events SET status = 'completed' WHERE id = ? AND status = 'started'`,
		eventID)
	return err
}

// AddEventParticipant adds a participant to an event
func (dm *MySQLDatabaseManager) AddEventParticipant(eventID int64, memberID int, tableID *int64, attended bool) error {
	var nullTableID sql.NullInt64
	if tableID != nil {
		nullTableID = sql.NullInt64{Int64: *tableID, Valid: true}
	}

	_, err := dm.db.Exec(`
		INSERT INTO event_attendance (event_id, member_id, table_id, attended)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			event_id = VALUES(event_id),
			table_id = VALUES(table_id),
			attended = VALUES(attended)`,
		eventID, memberID, nullTableID, attended)
	return err
}

func (dm *MySQLDatabaseManager) ListWithoutPairs() ([]models.Member, error) {
	rows, err := dm.db.Query(`
		SELECT 
			wp.member_id, 
			m.name, 
			m.telegram_username
		FROM without_pairs wp join members m
		on wp.member_id = m.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member

	for rows.Next() {
		var member models.Member
		err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.Telegram,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

func (dm *MySQLDatabaseManager) AddWithoutPairs(memberID int) error {
	_, err := dm.db.Exec(`
		INSERT IGNORE INTO without_pairs (member_id)
		VALUES (?)`, memberID)
	return err
}

// MarkAttendance marks a participant as attended or not attended
func (dm *MySQLDatabaseManager) MarkAttendance(eventID int64, memberID int, attended bool) error {
	_, err := dm.db.Exec(`
		UPDATE event_attendance 
		SET attended = ?
		WHERE event_id = ? AND member_id = ?`,
		attended, eventID, memberID)
	return err
}

func (dm *MySQLDatabaseManager) GetMemberIDByChatID(chatID int64) (int, error) {
	var memberID int
	err := dm.db.QueryRow(`
		SELECT m.id
		FROM members m 
		where m.telegram_chat_id = ?`,
		chatID).Scan(&memberID)

	if err != nil {
		return 0, err
	}
	return memberID, nil
}

// RemoveEventParticipant removes a participant from an event
func (dm *MySQLDatabaseManager) RemoveEventParticipant(eventID int64, memberID int) error {
	_, err := dm.db.Exec(`
		DELETE FROM event_attendance 
		WHERE event_id = ? AND member_id = ?`,
		eventID, memberID)
	return err
}

// IsUserRegisteredForEvent checks if a user is registered for an event
func (dm *MySQLDatabaseManager) IsUserRegisteredForEvent(eventID int64, chatID int64) (bool, error) {
	var count int
	err := dm.db.QueryRow(`
		SELECT COUNT(*) 
		FROM event_attendance ea
		join members m on ea.member_id = m.id
		WHERE ea.event_id = ? AND m.telegram_chat_id = ?`,
		eventID, chatID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetActivityDurations returns the total duration of all activities for an event
func (dm *MySQLDatabaseManager) GetActivityDurations(eventID int64) (int, error) {
	var totalDuration int
	err := dm.db.QueryRow(`
		SELECT COALESCE(SUM(duration_minutes), 0)
		FROM activities
		WHERE event_id = ?`, eventID).Scan(&totalDuration)
	if err != nil {
		return 0, err
	}
	return totalDuration, nil
}

func (dm *MySQLDatabaseManager) GetParticipants(eventID int64) ([]models.Participant, error) {
	rows, err := dm.db.Query(`
		SELECT 
			m.id, 
			m.name, 
			m.telegram_username, 
			m.telegram_chat_id, 
			m.phone_number,
			ea.attended,
			ea.table_id
		FROM event_attendance ea 
		JOIN members m ON ea.member_id = m.id
		WHERE ea.event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.Participant

	for rows.Next() {
		var participant models.Participant
		var telegram, telegramChatID, phone sql.NullString
		var tableID sql.NullInt64

		err := rows.Scan(
			&participant.ID,
			&participant.Name,
			&telegram,
			&telegramChatID,
			&phone,
			&participant.Attended,
			&tableID,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullString to *string
		if telegram.Valid {
			participant.Telegram = &telegram.String
		}
		if telegramChatID.Valid {
			participant.TelegramChatID = &telegramChatID.String
		}
		if phone.Valid {
			participant.Phone = &phone.String
		}
		if tableID.Valid {
			participant.TableID = &tableID.Int64
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

// GetEventsReadyForNotification finds events that need notifications sent
func (dm *MySQLDatabaseManager) GetEventsReadyForNotification() ([]models.Event, error) {
	now := time.Now()

	query := `
		SELECT 
			e.id, e.place_id, e.scheduled_at, e.topic, e.notify_at, 
			e.status, e.type, e.started_at, e.created_at, e.finished_at,
			p.name as place_name, p.metro_area, p.map_url, p.image_url
		FROM events e
		LEFT JOIN places p ON e.place_id = p.id
		WHERE e.notify_at IS NOT NULL 
		  AND e.notify_at <= ?
		  AND e.status = 'planned'
		ORDER BY e.notify_at ASC`

	rows, err := dm.db.Query(query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		var notifyAt sql.NullTime
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		var placeName, placeMetroArea string
		var placeMapURL, placeImageURL sql.NullString

		if err := rows.Scan(
			&e.ID, &e.PlaceID, &e.ScheduledAt, &e.Topic, &notifyAt,
			&e.Status, &e.Type, &startedAt, &e.CreatedAt, &finishedAt,
			&placeName, &placeMetroArea, &placeMapURL, &placeImageURL); err != nil {
			return nil, err
		}

		if notifyAt.Valid {
			e.NotifyAt = &notifyAt.Time
		}
		if startedAt.Valid {
			e.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			e.FinishedAt = &finishedAt.Time
		}

		// Build nested place object
		if e.PlaceID > 0 {
			e.Place = &models.Place{
				ID:        e.PlaceID,
				Name:      placeName,
				MetroArea: placeMetroArea,
			}
			if placeMapURL.Valid {
				e.Place.MapURL = placeMapURL.String
			}
			if placeImageURL.Valid {
				e.Place.ImageURL = placeImageURL.String
			}
		}

		events = append(events, e)
	}

	return events, nil
}

// MarkEventAsNotified marks an event as having notifications sent by clearing notify_at
func (dm *MySQLDatabaseManager) MarkEventAsNotified(eventID int64) error {
	_, err := dm.db.Exec(`
		UPDATE events 
		SET notify_at = NULL 
		WHERE id = ?`, eventID)
	return err
}

func (dm *MySQLDatabaseManager) GetActivities(eventID int64) ([]models.Activity, error) {
	rows, err := dm.db.Query("select id, name, duration_minutes, group_size from activities where event_id = ?", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.Activity

	for rows.Next() {
		var activity models.Activity
		err := rows.Scan(&activity.ID, &activity.Name, &activity.Duration, &activity.GroupSize)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (dm *MySQLDatabaseManager) GetPairingHistory() (map[int][]int, error) {
	query := `SELECT member_id, pair_id FROM pairing_history order by pairing_date ASC`
	rows, err := dm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]int)
	for rows.Next() {
		var (
			member_id int
			pair_id   int
		)
		if err := rows.Scan(&member_id, &pair_id); err != nil {
			return nil, err
		}
		result[member_id] = append(result[member_id], pair_id)
	}
	return result, nil
}

func (dm *MySQLDatabaseManager) ClearWithoutPairsTable() error {
	// Clear the without_pairs table
	_, err := dm.db.Exec("DELETE FROM without_pairs")
	if err != nil {
		return fmt.Errorf("Error clearing without_pairs table: %v\n", err)
	}

	return nil
}

// RemoveFromWithoutPairs removes a specific member from the without_pairs table
func (dm *MySQLDatabaseManager) RemoveFromWithoutPairs(memberID int) error {
	_, err := dm.db.Exec("DELETE FROM without_pairs WHERE member_id = ?", memberID)
	return err
}
