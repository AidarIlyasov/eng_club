package services

import (
	"fmt"
	"log"

	"eng_club/database"
	"eng_club/telegram"
)

// NotificationService handles event notification scheduling
type NotificationService struct {
	db       *database.MySQLDatabaseManager
	notifier *telegram.Notifier
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *database.MySQLDatabaseManager, notifier *telegram.Notifier) *NotificationService {
	return &NotificationService{
		db:       db,
		notifier: notifier,
	}
}

// CheckAndSendEventNotifications checks for events that need notifications and sends them
func (ns *NotificationService) CheckAndSendEventNotifications() error {
	// Get events ready for notification
	events, err := ns.db.GetEventsReadyForNotification()
	if err != nil {
		return fmt.Errorf("failed to get events ready for notification: %v", err)
	}

	fmt.Printf("events: %d\n", len(events))

	for _, event := range events {
		// Get participants for this event
		participants, err := ns.db.GetParticipants(event.ID)
		if err != nil {
			log.Printf("Failed to get participants for event %d: %v", event.ID, err)
			continue
		}
		fmt.Printf("participants: %d\n", len(participants))

		// Get total duration for the event
		totalDuration := 60 // Default duration in minutes
		if duration, err := ns.db.GetActivityDurations(event.ID); err == nil {
			totalDuration = duration
		}

		fmt.Printf("duration: %d\n", totalDuration)

		// Send notifications
		if err := ns.notifier.NotifyUpcomingEvent(event, participants, totalDuration); err != nil {
			log.Printf("Failed to send notification for event %d: %v", event.ID, err)
			continue
		}

		// Mark as notified
		if err := ns.db.MarkEventAsNotified(event.ID); err != nil {
			log.Printf("Failed to mark event %d as notified: %v", event.ID, err)
		}

		log.Printf("Successfully sent notification for event %d: %s", event.ID, event.Topic)
	}

	return nil
}
