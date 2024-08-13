package user

import (
	"database/sql"
	"errors"
	"log"
	"user/server/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(user types.User) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = $1", user.Username).Scan(&count)
	if err != nil {
		log.Println("Error querying database: ", err)
		return -1, err
	}

	if count > 0 {
		return -1, errors.New("user already exists")
	}

	var userID int
	err = s.db.QueryRow(`
				INSERT INTO Users 
				(Username, Password, Avatar) 
				VALUES ($1, $2, $3) RETURNING ID`,
		user.Username, user.Password, user.Avatar).Scan(&userID)
	if err != nil {
		log.Println("Error creating user")
		return -1, err
	}
	return userID, nil
}

// TODO: don't fetch password
func (s *Store) GetUserByEmail(username string) (*types.User, error) {
	rows, err := s.db.Query("SELECT ID, Username, Password, CreatedAt, Avatar FROM Users WHERE Username = $1", username)
	defer rows.Close()
	if err != nil {
		log.Println("Error querying database: ", err)
		return nil, err
	}

	user := new(types.User)
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.Avatar)
		if err != nil {
			log.Println("Error scanning rows: ", err)
			return nil, err
		}

		if err == nil {
			return user, nil
		}
	}
	log.Println("Invalid credentials")
	return nil, errors.New("invalid credentials")
}

func (s *Store) GetUserByID(userID int) (*types.User, error) {
	rows, err := s.db.Query("SELECT * FROM Users WHERE ID = $1", userID)
	defer rows.Close()
	if err != nil {
		log.Println("Error querying database: ", err)
		return nil, err
	}

	user := new(types.User)
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.Avatar)
		if err != nil {
			log.Println("Error scanning rows: ", err)
			return nil, err
		}

		if err == nil {
			return user, nil
		}
	}
	log.Println("Invalid credentials")
	return nil, errors.New("invalid credentials")
}
