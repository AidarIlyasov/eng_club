package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/valyala/fasthttp"

	"eng_club/models"
	"eng_club/services"
)

type eventBody struct {
	PlaceID     int64      `json:"place_id"`
	ScheduledAt string     `json:"scheduled_at"`
	Topic       string     `json:"topic"`
	Type        string     `json:"type,omitempty"`
	FinishedAt  string     `json:"finished_at,omitempty"`
	NotifyAt    *string    `json:"notify_at,omitempty"`
	Activities  []Activity `json:"activities,omitempty"`
}

type Activity struct {
	Name      string `json:"name"`
	Duration  int    `json:"duration"`
	GroupSize int    `json:"groupSize"`
}

type startEventParticipant struct {
	Name     string `json:"name"`
	Telegram string `json:"telegram"`
	MemberID int    `json:"member_id"`
}

type startEventBody struct {
	Participants []startEventParticipant `json:"participants"`
	Durations    map[string]int          `json:"durations,omitempty"`
}

type moveToTableBody struct {
	MemberID int `json:"member_id"`
	TableID  int `json:"table_id"`
}

// ListEvents handles GET /api/events.
func (d *Deps) ListEvents(ctx *fasthttp.RequestCtx) {
	events, err := d.DB.ListEvents("", "", "", "", 0, 0)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if events == nil {
		events = []models.Event{}
	}
	WriteJSON(ctx, fasthttp.StatusOK, events)
}

// GetEvent handles GET /api/events/{id}.
func (d *Deps) GetEvent(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	e, err := d.DB.GetEvent(id)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if e == nil {
		WriteErr(ctx, fasthttp.StatusNotFound, "event not found")
		return
	}
	WriteJSON(ctx, fasthttp.StatusOK, e)
}

// CreateEvent handles POST /api/events.
func (d *Deps) CreateEvent(ctx *fasthttp.RequestCtx) {
	var body eventBody
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.PlaceID <= 0 {
		WriteErr(ctx, fasthttp.StatusBadRequest, "place_id is required")
		return
	}
	place, err := d.DB.GetPlace(body.PlaceID)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if place == nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "place not found")
		return
	}
	t, err := time.Parse(time.RFC3339, body.ScheduledAt)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "scheduled_at must be RFC3339 (e.g. 2026-04-10T19:30:00+03:00)")
		return
	}

	// Handle notify_at
	var notifyAt *time.Time
	if body.NotifyAt != nil {
		if *body.NotifyAt != "" {
			notifyTime, err := time.Parse(time.RFC3339, *body.NotifyAt)
			if err != nil {
				WriteErr(ctx, fasthttp.StatusBadRequest, "notify_at must be RFC3339")
				return
			}
			notifyAt = &notifyTime
		}
		// If NotifyAt is present but empty string, notifyAt stays nil (notifications disabled)
	} else {
		// Default to 8 hours before scheduled_at if not specified
		defaultNotify := t.Add(-8 * time.Hour)
		notifyAt = &defaultNotify
	}

	// Handle finished_at for single activity events
	var finishedAt *time.Time
	if body.Type == "single_activity" {
		if body.FinishedAt == "" {
			WriteErr(ctx, fasthttp.StatusBadRequest, "finished_at is required for single_activity events")
			return
		}
		finishedTime, err := time.Parse(time.RFC3339, body.FinishedAt)
		if err != nil {
			WriteErr(ctx, fasthttp.StatusBadRequest, "finished_at must be RFC3339")
			return
		}
		finishedAt = &finishedTime
	}

	e, err := d.DB.CreateEvent(body.PlaceID, t, body.Topic, body.Type, finishedAt, notifyAt)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	// Save activities
	if len(body.Activities) > 0 {
		for _, activity := range body.Activities {
			err := d.DB.CreateEventActivity(e.ID, activity.Name, activity.Duration, activity.GroupSize)
			if err != nil {
				// Rollback by deleting the event
				d.DB.DeleteEvent(e.ID)
				WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to create activities: "+err.Error())
				return
			}
		}
	}

	WriteJSON(ctx, fasthttp.StatusCreated, e)
}

// UpdateEvent handles PUT /api/events/{id}.
func (d *Deps) UpdateEvent(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	// Get current event
	currentEvent, err := d.DB.GetEvent(id)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if currentEvent == nil {
		WriteErr(ctx, fasthttp.StatusNotFound, "event not found")
		return
	}

	var body eventBody
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}

	// Use current place_id if not provided
	placeID := body.PlaceID
	if placeID <= 0 {
		placeID = currentEvent.PlaceID
	}

	place, err := d.DB.GetPlace(placeID)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if place == nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "place not found")
		return
	}

	// Use current scheduled_at if not provided
	var t time.Time
	if body.ScheduledAt != "" {
		t, err = time.Parse(time.RFC3339, body.ScheduledAt)
		if err != nil {
			WriteErr(ctx, fasthttp.StatusBadRequest, "scheduled_at must be RFC3339")
			return
		}
	} else {
		t = currentEvent.ScheduledAt
	}

	// Use current topic if not provided
	topic := body.Topic
	if topic == "" {
		topic = currentEvent.Topic
	}

	// Use current type if not provided
	eventType := body.Type
	if eventType == "" {
		eventType = currentEvent.Type
	}

	// Handle notify_at
	var notifyAt *time.Time
	if body.NotifyAt != nil {
		// NotifyAt field is explicitly present in JSON
		if *body.NotifyAt == "" || *body.NotifyAt == "null" {
			// Empty string or null means disable notifications
			notifyAt = nil
		} else {
			// Valid time string
			notifyTime, err := time.Parse(time.RFC3339, *body.NotifyAt)
			if err != nil {
				WriteErr(ctx, fasthttp.StatusBadRequest, "notify_at must be RFC3339")
				return
			}
			notifyAt = &notifyTime
		}
	} else {
		// NotifyAt field not present in JSON, keep current value or set default
		if currentEvent.NotifyAt != nil {
			notifyAt = currentEvent.NotifyAt
		} else {
			// Default to 8 hours before scheduled_at
			defaultNotify := t.Add(-8 * time.Hour)
			notifyAt = &defaultNotify
		}
	}

	// Handle finished_at for single activity events
	var finishedAt *time.Time
	if eventType == "single_activity" {
		if body.FinishedAt != "" {
			finishedTime, err := time.Parse(time.RFC3339, body.FinishedAt)
			if err != nil {
				WriteErr(ctx, fasthttp.StatusBadRequest, "finished_at must be RFC3339")
				return
			}
			finishedAt = &finishedTime
		} else if currentEvent.FinishedAt != nil {
			finishedAt = currentEvent.FinishedAt
		}
	}

	e, err := d.DB.UpdateEvent(id, placeID, t, topic, notifyAt, eventType, finishedAt)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if e == nil {
		WriteErr(ctx, fasthttp.StatusNotFound, "event not found or not in planned status")
		return
	}
	WriteJSON(ctx, fasthttp.StatusOK, e)
}

// DeleteEvent handles DELETE /api/events/{id}.
func (d *Deps) DeleteEvent(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	err = d.DB.DeleteEvent(id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteErr(ctx, fasthttp.StatusNotFound, "event not found or not in planned status")
			return
		}
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// StartEvent handles POST /api/events/{id}/start.
func (d *Deps) StartEvent(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	ev, err := d.DB.GetEvent(id)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if ev == nil {
		WriteErr(ctx, fasthttp.StatusNotFound, "event not found")
		return
	}
	if ev.Status != "planned" {
		WriteErr(ctx, fasthttp.StatusConflict, "event already started or completed")
		return
	}

	var body startEventBody
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}
	// if len(body.Participants) == 0 {
	// 	WriteErr(ctx, fasthttp.StatusBadRequest, "at least one participant is required")
	// 	return
	// }

	// Load durations from activity_durations table
	duration, err := d.DB.GetActivityDurations(id)
	if err != nil {

	}

	// Mark event as started
	ok, err := d.DB.MarkEventStarted(id)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		WriteErr(ctx, fasthttp.StatusConflict, "could not mark event started (race?)")
		return
	}

	manager, err := services.NewEventManager(id, d.DB, d.TelegramBot)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	go func() {
		err = manager.StartEvent()
		if err != nil {
			log.Fatal("Failed to start event:", err)
		}

		participats, err := d.DB.GetParticipants(id)

		// Send session start notification
		d.TelegramBot.NotifySessionStart(id, participats, duration)

		if err != nil {
			WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
	}()

	WriteJSON(ctx, fasthttp.StatusOK, map[string]any{
		"event_id": id,
		"status":   "started",
	})
}

// GetEventParticipants handles GET /api/events/{id}/participants
func (d *Deps) GetEventParticipants(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	participants, err := d.DB.GetParticipants(id)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if participants == nil {
		participants = []models.Participant{}
	}

	WriteJSON(ctx, fasthttp.StatusOK, participants)
}

// MarkAttendance handles POST /api/events/{id}/attendance
func (d *Deps) MarkAttendance(ctx *fasthttp.RequestCtx) {
	eventID, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		MemberID int  `json:"member_id"`
		Attended bool `json:"attended"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}

	if body.MemberID <= 0 {
		WriteErr(ctx, fasthttp.StatusBadRequest, "member_id is required")
		return
	}

	if err := d.DB.MarkAttendance(eventID, body.MemberID, body.Attended); err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(ctx, fasthttp.StatusOK, map[string]any{
		"success": true,
	})
}

// AddParticipantToEvent handles POST /api/events/{id}/add-participant
func (d *Deps) AddParticipantToEvent(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		Name     string `json:"name"`
		Telegram string `json:"telegram"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}

	if body.Name == "" || body.Telegram == "" {
		WriteErr(ctx, fasthttp.StatusBadRequest, "name and telegram are required")
		return
	}

	// Upsert member
	memberID, err := d.DB.UpsertMember(body.Name, body.Telegram, "", nil)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	// Add to event_attendance
	if err := d.DB.AddEventParticipant(id, memberID, nil, false); err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(ctx, fasthttp.StatusOK, map[string]any{
		"success":   true,
		"member_id": memberID,
	})
}

// RemoveParticipantFromEvent handles DELETE /api/events/{id}/participants/{member_id}
func (d *Deps) RemoveParticipantFromEvent(ctx *fasthttp.RequestCtx) {
	eventID, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	memberID, err := PathID(ctx, "member_id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	if err := d.DB.RemoveEventParticipant(eventID, int(memberID)); err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// GetSessionActivities handles GET /api/events/{id}/activities
func (d *Deps) GetSessionActivities(ctx *fasthttp.RequestCtx) {
	eventID, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	event, err := d.DB.GetEvent(eventID)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if event == nil {
		WriteErr(ctx, fasthttp.StatusNotFound, "event not found")
		return
	}

	activities, err := d.DB.GetEventActivities(eventID)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	if activities == nil {
		activities = []map[string]interface{}{}
	}

	WriteJSON(ctx, fasthttp.StatusOK, activities)
}

// MoveToTable handles POST /api/events/{id}/move-to-table
func (d *Deps) MoveToTable(ctx *fasthttp.RequestCtx) {
	eventID, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}

	var body moveToTableBody
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}

	if body.MemberID <= 0 || body.TableID <= 0 {
		WriteErr(ctx, fasthttp.StatusBadRequest, "member_id and table_id are required and must be positive")
		return
	}

	// Remove from without_pairs table
	if err := d.DB.RemoveFromWithoutPairs(body.MemberID); err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to remove from without_pairs: "+err.Error())
		return
	}

	// Add or update in event_attendance table
	tableID := int64(body.TableID)
	if err := d.DB.AddEventParticipant(eventID, body.MemberID, &tableID, true); err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to add to event_attendance: "+err.Error())
		return
	}

	WriteJSON(ctx, fasthttp.StatusOK, map[string]any{
		"success": true,
		"message": "Member moved to table successfully",
	})
}
