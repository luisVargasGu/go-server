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

func (s *Store) SaveInvite(invite *types.Invite) error {
	_, err := s.db.Exec(`INSERT INTO Invites (ChannelID, InviterID, InviteCode, Expiration)
	VALUES ($1, $2, $3, $4)`, invite.ChanelID, invite.InviterID, invite.InviteCode, invite.Expiration)
	if err != nil {
		log.Println("Error linking channel to user")
		return err
	}

	return nil
}

func (s *Store) FindInvite(inviteLink string) (*types.Invite, error) {
	rows, err := s.db.Query(`SELECT ID, ChannelID, InviterID, InviteeID, InviteCode, CreatedAt, Expiration
	FROM Invites WHERE InviteCode = $1`, inviteLink)
	if err != nil {
		log.Println("Error getting Invite")
		return nil, err
	}
	defer rows.Close()

	invite := types.Invite{}
	for rows.Next() {
		err = rows.Scan(
			&invite.ID, &invite.ChanelID, &invite.InviterID,
			&invite.InviteeID, &invite.InviteCode,
			&invite.CreatedAt, &invite.Expiration)
		if err != nil {
			log.Println("Error scanning invite")
			return nil, err
		}
	}

	return &invite, nil
}

func (s *Store) AcceptInvite(inviteID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE Invites SET InviteeID = $1 WHERE ID = $2`, userID, inviteID)
	if err != nil {
		tx.Rollback()
		return err
	}

	var channelID int
	err = tx.QueryRow(`SELECT ChannelID FROM Invites WHERE ID = $1`, inviteID).Scan(&channelID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`INSERT INTO ChannelsToUsers (user_id, channel_id) VALUES ($1, $2)
                      ON CONFLICT DO NOTHING`, userID, channelID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

