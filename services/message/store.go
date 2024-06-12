package message

import (
	"database/sql"
	"encoding/json"
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
			m.ID, 
			m.RoomID, 
			m.SenderID, 
			m.Content, 
			m.Timestamp, 
			m.IsRead, 
			u.Avatar, 
			u.Username,
			COALESCE(json_agg(json_build_object(
			'avatar', encode(su.Avatar, 'base64'),
			'username', su.Username)) FILTER (WHERE su.ID IS NOT NULL), '[]') as SeenBy
		FROM Messages m
		JOIN Users u ON u.ID = m.SenderID
		LEFT JOIN SeenMessages sm ON sm.message_id = m.ID
		LEFT JOIN Users su ON su.ID = sm.user_id
		WHERE m.RoomID = $1
		GROUP BY m.ID, u.Avatar, u.Username
	`, roomID)
	if err != nil {
		log.Println("Error getting messages in room: ", err)
		return nil, err
	}
	defer rows.Close()

	messages := make([]*types.Message, 0)
	for rows.Next() {
		m := &types.Message{}
		var avatarBytes []byte
		var seenByJson string

		err := rows.Scan(
			&m.ID,
			&m.RoomID,
			&m.SenderID,
			&m.Content,
			&m.Timestamp,
			&m.IsRead,
			&avatarBytes,
			&m.SenderName,
			&seenByJson)
		if err != nil {
			log.Println("Error scanning message row: ", err)
			return nil, err
		}

		m.SenderAvatar = image.EncodeB64Image(avatarBytes)

		var seenBy []struct {
			Avatar   []byte `json:"avatar"`
			Username string `json:"username"`
		}
		err = json.Unmarshal([]byte(seenByJson), &seenBy)
		if err != nil {
			log.Println("Error unmarshalling seen by data: ", err)
			return nil, err
		}

		m.SeenBy = make([]types.SeenByUser, len(seenBy))
		for i, sb := range seenBy {
			m.SeenBy[i] = types.SeenByUser{
				Avatar:   image.EncodeB64Image(sb.Avatar),
				Username: sb.Username,
			}
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

// TODO: change to use websocket to do this kind of update
func (s *Store) MarkMessageAsSeen(userID int, messageID int) error {
	_, err := s.db.Exec(`
        INSERT INTO SeenMessages (message_id, user_id)
        VALUES ($1, $2)
        ON CONFLICT (message_id, user_id) DO NOTHING
    `, messageID, userID)
	if err != nil {
		log.Println("Error marking message as seen:", err)
		return err
	}

	return err
}
