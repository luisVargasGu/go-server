package types

type PermissionsStore interface {
	UserHasPermission(userID ,channelID int) bool
}
