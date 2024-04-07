package channel

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

func (s *Store) GetChannelsForUser(userID int) ([]*types.Channel, error) {
	rows, err := s.db.Query(`SELECT Channels.* 
                            FROM Channels
                            JOIN ChannelsToUsers 
                            ON Channels.ID = ChannelsToUsers.channel_id
                            WHERE ChannelsToUsers.user_id = $1`, userID)
	if err != nil {
		log.Println("Error getting channels for user")
		return nil, err
	}
	defer rows.Close()

	channels := make([]*types.Channel, 10)
	for rows.Next() {
		channel := &types.Channel{}
		err := rows.Scan(&channel.ID, &channel.Name)
		if err != nil {
			log.Println("Error scanning channel")
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func (s *Store) CreateChannel(channel types.CreateChannelPayload) error {
	_, err := s.db.Exec(`INSERT INTO Channels (Name) VALUES ($1)`, channel.Name)
	if err != nil {
		log.Println("Error creating channel")
		return err
	}

	return nil
}

func (s *Store) DeleteChannel(channelID int) error {
	_, err := s.db.Exec(`DELETE FROM Channels WHERE ID = $1`, channelID)
	if err != nil {
		log.Println("Error deleting channel")
		return err
	}

	return nil
}
