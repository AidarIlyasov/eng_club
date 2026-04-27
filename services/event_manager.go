package services

import (
	"fmt"
	"time"

	"eng_club/club"
	"eng_club/database"
	"eng_club/models"
	"eng_club/telegram"
)

type EventManager struct {
	ID          int64
	db          *database.MySQLDatabaseManager
	telegramBot *telegram.Notifier
}

func NewEventManager(id int64, db *database.MySQLDatabaseManager, telegramBot *telegram.Notifier) (*EventManager, error) {
	return &EventManager{
		ID:          id,
		db:          db,
		telegramBot: telegramBot,
	}, nil
}

func (em *EventManager) StartEvent() error {
	// Get event details to determine type
	event, err := em.db.GetEvent(em.ID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Get participants for notifications
	participants, err := em.db.GetParticipants(em.ID)
	if err != nil {
		return err
	}

	if event.Type == "single_activity" {
		return em.handleSingleActivity(event, participants)
	} else {
		return em.handleManyActivities(participants)
	}
}

func (em *EventManager) handleSingleActivity(event *models.Event, participants []models.Participant) error {
	if event.FinishedAt == nil {
		return fmt.Errorf("single activity event must have finished_at time")
	}

	fmt.Printf("Starting single activity event (ID: %d)\n", em.ID)
	fmt.Printf("Activity will end at: %s\n", event.FinishedAt.Format("2006-01-02 15:04:05"))

	// Calculate duration until finished_at
	now := time.Now()
	duration := event.FinishedAt.Sub(now)

	if duration <= 0 {
		fmt.Printf("Activity already finished, marking as completed immediately\n")
		return em.completeEvent(participants)
	}

	fmt.Printf("Waiting %v until activity ends...\n", duration)
	timer := time.NewTimer(duration)
	<-timer.C

	fmt.Printf("Single activity completed, marking event as finished\n")
	return em.completeEvent(participants)
}

func (em *EventManager) handleManyActivities(participants []models.Participant) error {
	activities, err := em.db.GetActivities(em.ID)
	if err != nil {
		return err
	}

	if len(activities) == 0 {
		fmt.Printf("No activities found for event %d, marking as completed\n", em.ID)
		return em.completeEvent(participants)
	}

	for i, activity := range activities {
		fmt.Printf("Starting activity: %s (duration: %d minutes)\n", activity.Name, activity.Duration)

		if err := em.runActivity(activity); err != nil {
			return err
		}

		// Wait for activity duration before starting next activity (except for last one)
		if i < len(activities)-1 {
			fmt.Printf("Waiting %d minutes before next activity...\n", activity.Duration)
			timer := time.NewTimer(time.Duration(activity.Duration) * time.Minute)
			<-timer.C
		}
	}

	fmt.Printf("All activities completed, marking event as finished\n")
	return em.completeEvent(participants)
}

func (em *EventManager) completeEvent(participants []models.Participant) error {
	// Mark event as completed in database
	if err := em.db.MarkEventCompleted(em.ID); err != nil {
		return fmt.Errorf("failed to mark event as completed: %w", err)
	}

	// Convert slice to map for NotifySessionEnd
	participantMap := make(map[int]models.Participant)
	for _, p := range participants {
		participantMap[p.ID] = p
	}

	// Send session end notifications
	if err := em.telegramBot.NotifySessionEnd(em.ID, participantMap); err != nil {
		// Log error but don't fail the completion
		fmt.Printf("Failed to send session end notifications: %v\n", err)
	}

	return nil
}

func (em *EventManager) runActivity(activity models.Activity) error {
	if err := em.db.ClearWithoutPairsTable(); err != nil {
		return err
	}

	pairHistory, err := em.db.GetPairingHistory()
	if err != nil {
		return err
	}

	participantsList, err := em.db.GetParticipants(em.ID)
	if err != nil {
		return err
	}

	// Convert slice to map for shuffler
	participants := make(map[int]models.Participant)
	for _, p := range participantsList {
		participants[p.ID] = p
	}

	shuffler := club.NewShuffler(em.db.GetDB(), em.telegramBot, em.ID, participants, pairHistory, activity.GroupSize)
	shuffler.Start()

	return nil
}

// Close closes database connections
func (ecm *EventManager) Close() error {
	return ecm.db.Close()
}
