package message

import (
	"database/sql"
	"log"
	"user/server/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetMessagesInRoom(roomID int) ([]*types.Message, error) {
	rows, err := s.db.Query("SELECT * FROM Messages WHERE RoomID = $1", roomID)
	if err != nil {
		log.Println("Error getting messages in room: ", err)
		return nil, err
	}
	defer rows.Close()

	messages := make([]*types.Message, 0)
	for rows.Next() {
		m := &types.Message{}
		err := rows.Scan(&m.ID, &m.RoomID, &m.SenderID, &m.Content, &m.Timestamp)
		if err != nil {
			log.Println("Error scanning message row: ", err)
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (s *Store) CreateMessage(m types.CreateMessagePayload) error {
	_, err := s.db.Exec("INSERT INTO Messages (RoomID, SenderID, Content) VALUES ($1, $2, $3)", m.RoomID, m.SenderID, m.Content)
	if err != nil {
		log.Println("Error creating message: ", err)
		return err
	}
	return nil
}
