package message

import (
	"database/sql"
	"log"
	"user/server/services/image"
	"user/server/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetMessagesInRoom(roomID int) ([]*types.Message, error) {
	rows, err := s.db.Query(`
				SELECT 
			Messages.ID, 
			Messages.RoomID, 
			Messages.SenderID, 
			Messages.Content, 
			Messages.Timestamp, 
			Messages.IsRead, 
			Users.Avatar, 
			Users.Username 
		FROM Messages 
		JOIN Users ON Users.ID = Messages.SenderID 
		WHERE RoomID = $1	
				`, roomID)
	if err != nil {
		log.Println("Error getting messages in room: ", err)
		return nil, err
	}
	defer rows.Close()

	messages := make([]*types.Message, 0)
	for rows.Next() {
		m := &types.Message{}
		avatarBytes := &[]byte{}
		err := rows.Scan(
			&m.ID,
			&m.RoomID,
			&m.SenderID,
			&m.Content,
			&m.Timestamp,
			&m.IsRead,
			&avatarBytes,
			&m.SenderName)
		m.SenderAvatar = image.EncodeB64Image(*avatarBytes)
		if err != nil {
			log.Println("Error scanning message row: ", err)
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (s *Store) CreateMessage(m types.Message) error {
	_, err := s.db.Exec("INSERT INTO Messages (RoomID, SenderID, Content, IsRead) VALUES ($1, $2, $3, $4)",
		m.RoomID, m.SenderID, m.Content, false)
	if err != nil {
		log.Println("Error creating message: ", err)
		return err
	}
	return nil
}
