package polluser

import "time"

//An user object with info in it
type User struct {
	UID  string
	name string
	pass []byte
	salt []byte

	CreateTime time.Time
}
