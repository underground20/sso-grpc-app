package id

import (
	u "github.com/google/uuid"
)

func Generate() u.UUID {
	id, err := u.NewV7()
	if err != nil {
		panic(err)
	}

	return id
}

func CreateFromString(uuid string) u.UUID {
	id, err := u.Parse(uuid)
	if err != nil {
		panic(err)
	}

	return id
}
