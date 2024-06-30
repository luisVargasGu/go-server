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

func (s *Store) GetAllChannels() ([]*types.Channel, error) {
	rows, err := s.db.Query(`SELECT ID, Name FROM Channels`)

	if err != nil {
		log.Println("Error getting all channels")
		return nil, err
	}
	defer rows.Close()

	channels := make([]*types.Channel, 0)
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

	channels := make([]*types.Channel, 0)
	for rows.Next() {
		channel := &types.Channel{}
		err := rows.Scan(&channel.ID, &channel.Name, &channel.Avatar)
		if err != nil {
			log.Println("Error scanning channel")
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func (s *Store) CreateChannel(channel *types.Channel, user *types.User) error {
	err := s.db.QueryRow(`INSERT INTO Channels (Name, Avatar) VALUES ($1, $2)
			      RETURNING ID`,
		channel.Name, channel.Avatar).Scan(&channel.ID)
	if err != nil {
		log.Println("Error creating channel")
		return err
	}

	_, err = s.db.Exec(`INSERT INTO ChannelsToUsers (channel_id, user_id) VALUES ($1, $2)`,
		channel.ID, user.ID)
	if err != nil {
		log.Println("Error linking channel to user")
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
