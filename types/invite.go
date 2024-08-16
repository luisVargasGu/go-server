package types

import "time"

type Invite struct {
	ID         int       `json:"id"`
	ChanelID   int       `json:"channel_id"`
	InviterID  int       `json:"inviter_id"`
	InviteeID  int       `json:"invitee_id"`
	InviteCode string    `json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`
	Expiration time.Time `json:"expiration"`
}

type InviteStore interface {
	SaveInvite(inv *Invite) error
	FindInvite(inviteCode string) (*Invite, error)
	AcceptInvite(invtID, userID int) error
}

type InviteRespose struct {
	Link string `json:"link"`
}

type InviteAcceptedResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Inv     Invite `json:"invite"`
}
