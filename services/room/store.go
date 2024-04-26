package room

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

func (s *Store) GetRoomsInChannel(channelID int) ([]*types.Room, error) {
	rows, err := s.db.Query(`SELECT Rooms.* 
                            FROM Rooms
                            JOIN RoomsToChannels 
                            ON Rooms.ID= RoomsToChannels.room_id
                            WHERE RoomsToChannels.channel_id = $1`, channelID)
	if err != nil {
		log.Println("Error getting rooms in channel")
		return nil, err
	}
	defer rows.Close()

	rooms := make([]*types.Room, 0)
	for rows.Next() {
		room := &types.Room{}
		err := rows.Scan(&room.ID, &room.Name, &room.ChannelID)
		if err != nil {
			log.Println("Error scanning room")
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (s *Store) CreateRoom(room types.CreateRoomPayload) error {
	_, err := s.db.Exec(`INSERT INTO Rooms (Name, ChannelID) VALUES ($1, $2)`, room.Name, room.ChannelID)
	if err != nil {
		log.Println("Error creating room")
		return err
	}

	return nil
}

func (s *Store) DeleteRoom(roomID int) error {
	_, err := s.db.Exec(`DELETE FROM Rooms WHERE ID = $1`, roomID)
	if err != nil {
		log.Println("Error deleting room")
		return err
	}

	return nil
}
