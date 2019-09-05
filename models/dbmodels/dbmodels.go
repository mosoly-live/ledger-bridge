package dbmodels

import (
	"time"
)

// User is DB user.
type User struct {
	ID            int       `db:"id"`
	InviteURLHash string    `db:"invite_url_hash"`
	Account       string    `db:"account"`
	UpdatedAt     time.Time `db:"updated_at"`
	Validated     bool      `db:"validated"`
	Mentorees     []Mentoree
	Mentors       []Mentor
}

// Mentor is DB mentor - needed part of User fields.
type Mentor struct {
	ID      int
	Account string
	Users   []string
}

// Mentoree is DB mentoree - needed part of User fields.
type Mentoree struct {
	ID      int
	Account string
}

// Project is DB project.
type Project struct {
	ID              int       `db:"id"`
	Name            string    `db:"name"`
	UpdatedAt       time.Time `db:"updated_at"`
	PassportAddress string    `db:"passport_address"`
}
