package club

import (
	"database/sql"
	"eng_club/models"
	"eng_club/telegram"
	"fmt"
	"slices"
	"strings"
)

type Shuffler struct {
	db                    *sql.DB
	eventID               int64
	groupSize             int
	availableParticipants map[int]models.Participant
	pairingHistory        PairingHistory
	telegramBot           *telegram.Notifier
}

type PairingHistory map[int][]int

func NewShuffler(db *sql.DB, telegramBot *telegram.Notifier, eventID int64, availableParticipants map[int]models.Participant, pairingHistory PairingHistory, groupSize int) *Shuffler {
	return &Shuffler{
		db:                    db,
		eventID:               eventID,
		availableParticipants: availableParticipants,
		pairingHistory:        pairingHistory,
		groupSize:             groupSize,
		telegramBot:           telegramBot,
	}
}

// updatePairingHistory обновляет историю встреч для всех участников группы в БД
func (s *Shuffler) updatePairingHistory(group []models.Participant) {
	for _, participant := range group {
		history := s.pairingHistory[participant.ID]

		// Для каждого участника добавляем всех остальных членов группы в его историю
		for _, other := range group {
			if participant.ID == other.ID {
				continue
			}

			s.db.Exec("INSERT INTO pairing_history(member_id, pair_id) VALUES (?,?) ON DUPLICATE KEY UPDATE pairing_date = CURRENT_TIMESTAMP", participant.ID, other.ID)
			// Удаляем старую запись об этом партнере, если она есть (FIFO)
			history = s.removeFromSlice(history, other.ID)
			// Добавляем партнера в конец (самая свежая встреча)
			history = append(history, other.ID)
		}

		s.pairingHistory[participant.ID] = history
	}

}

func (s *Shuffler) removeFromSlice(slice []int, val int) []int {
	for i, v := range slice {
		if v == val {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func (s *Shuffler) bestCandidate(member models.Participant) []models.Participant {
	group := []models.Participant{member}
	memberHistory := s.pairingHistory[member.ID]

	// 1. Сначала ищем среди тех, с кем встреч вообще не было
	for _, candidate := range s.availableParticipants {
		if len(group) == s.groupSize {
			break
		}
		// Проверяем, что это не сам участник и его нет в истории встреч
		if candidate.ID != member.ID && !slices.Contains(memberHistory, candidate.ID) {
			group = append(group, candidate)
		}
	}

	// 2. Если среди новых людей не хватило, ищем по истории (самые старые встречи)
	if len(group) < s.groupSize {
		for _, candidateId := range memberHistory {
			if len(group) == s.groupSize {
				break
			}
			if participant, available := s.availableParticipants[candidateId]; available && candidateId != member.ID {
				group = append(group, participant)
			}
		}
	}

	// 3. Обновляем историю и удаляем из доступных ТОЛЬКО если группа полная
	if len(group) == s.groupSize {
		s.updatePairingHistory(group)
		s.updateAvailableParticipants(group)
		return group
	}

	return nil // Возвращаем nil, если не набрали группу
}

func (s *Shuffler) updateAvailableParticipants(group []models.Participant) {
	// remove all group members from availableParticipants list
	for _, participant := range group {
		delete(s.availableParticipants, participant.ID)
	}
}

func (s *Shuffler) Start() {
	tableID := 1

	for len(s.availableParticipants) >= s.groupSize {
		// Берем первого попавшегося ID из доступных
		var currentParticipant models.Participant
		for _, participant := range s.availableParticipants {
			currentParticipant = participant
			break
		}

		pairs := s.bestCandidate(currentParticipant)

		if len(pairs) == s.groupSize {
			msg := "Created group: "
			for _, participant := range pairs {
				s.telegramBot.NotifyTableAssignments(participant, tableID)
				// Update table_id in event_attendance
				_, err := s.db.Exec("UPDATE event_attendance SET table_id = ? WHERE event_id = ? AND member_id = ?",
					tableID, s.eventID, participant.ID)
				if err != nil {
					fmt.Printf("Error updating table_id for member %d: %v\n", participant.ID, err)
				}

				msg += participant.Name + ","
			}
			fmt.Println(msg)
		} else {
			// Если вдруг не смогли набрать группу (например, из-за фильтров внутри)
			fmt.Printf("Could not form a full group for member %s\n", currentParticipant.Name)
			// Чтобы не попасть в бесконечный цикл, если группа не собралась,
			// нужно либо удалить участника, либо изменить логику.
			delete(s.availableParticipants, currentParticipant.ID)
		}
		tableID++
	}

	// Сообщение о тех, кто остался без группы
	if len(s.availableParticipants) > 0 {
		rest := []string{}
		for _, participant := range s.availableParticipants {
			rest = append(rest, participant.Name)

			// Insert remaining member into without_pairs table
			_, err := s.db.Exec("INSERT INTO without_pairs (member_id) VALUES (?)", participant.ID)
			if err != nil {
				fmt.Printf("Error inserting member %d into without_pairs: %v\n", participant.ID, err)
			}
		}

		fmt.Printf("Not enough participants to form another group. Remaining: %s\n", strings.Join(rest, ","))
	}
}
