package types

import "time"

type Invite struct {
	ID         int       `json:"id"`
	ChanelID   int       `json:"channel_id"`
	InviterID  int       `json:"inviter_id"`
	InviteCode string    `json:"invite_code"`
	CreatedAt  string    `json:"created_at"`
	Expiration time.Time `json:"expiration"`
}

type InviteStore interface {
	SaveInvite(inv Invite) error
	AcceptInvite(inviteCode string, userID int) error
}

type InviteRespose struct {
	Link string `json:"link"`
}
