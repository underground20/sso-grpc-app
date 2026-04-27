package models

import "time"

type User struct {
	ID        string     `db:"id"`
	Email     string     `db:"email"`
	PassHash  []byte     `db:"pass_hash"`
	Username  *string    `db:"username"`
	Status    int        `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	LastLogin *time.Time `db:"last_login"`
}
