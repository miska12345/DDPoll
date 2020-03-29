package polluser

import "time"

//An user object with info in it
type User struct {
	UID  string
	Name string
	pass []byte
	salt []byte

	CreateTime time.Time
}
