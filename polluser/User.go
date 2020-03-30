package polluser

import "time"

//An user object with info in it
type User struct {
	UID        string `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string
	Pass       []byte
	Salt       []byte
	CreateTime time.Time
}
