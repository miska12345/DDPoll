package polluser

import "time"

//An user object with info in it
type User struct {
	UID        string `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string
	pass       []byte
	salt       []byte
	pollGroup  map[string]uint32
	CreateTime time.Time
}

func (u *User) Salt() []byte {
	return u.salt
}

func (u *User) Pass() []byte {
	return u.pass
}
