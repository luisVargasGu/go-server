package image

import (
	"database/sql"
	"user/server/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetImage(id int) ([]byte, error) {
	var data []byte
	err := s.db.QueryRow("SELECT data FROM images WHERE id = $1", id).Scan(&data)
	return data, err
}

func (s *Store) UpdateUserAvatar(data []byte, user *types.User) error {
	_, err := s.db.Exec("UPDATE Users SET Avatar = $1 WHERE id = $2", data, user.ID)
	return err
}
