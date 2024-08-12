package permissions

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

func (s *Store) UserHasPermission(userID, channelID int) bool {
	rows, err := s.db.Query(`SELECT Channels.* 
                            FROM Channels
                            JOIN ChannelsToUsers 
                            ON Channels.ID = ChannelsToUsers.channel_id
                            WHERE ChannelsToUsers.user_id = $1`, userID)
	if err != nil {
		log.Println("Error getting channels for user")
		return false
	}
	defer rows.Close()

	channels := make([]*types.Channel, 0)
	for rows.Next() {
		channel := &types.Channel{}
		err := rows.Scan(&channel.ID, &channel.Name, &channel.Avatar)
		if err != nil {
			log.Println("Error scanning channel")
			return false
		}
		channels = append(channels, channel)
	}
	for _, channel := range channels {
		if channel.ID == channelID {
			return true
		}
	}
	return false
}
