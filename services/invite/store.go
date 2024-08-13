package invite

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

func (s *Store) SaveInvite(invite types.Invite) error {
	_, err := s.db.Exec(`INSERT INTO Invites (ChannelID, InviterID, InviteCode, Expiration)
	VALUES ($1, $2, $3, $4, $5)`, invite.ChanelID, invite.InviterID, invite.InviteCode, invite.Expiration)
	if err != nil {
		log.Println("Error linking channel to user")
		return err
	}

	return nil
}
