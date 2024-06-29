package types

type ImageStore interface {
	UpdateUserAvatar(avatar []byte, user *User) error
	GetImage(imageID int) ([]byte, error)
}
