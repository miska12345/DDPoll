package user

import "time"

type User struct {
	UID  string
	name string
	pass []byte
	salt []byte

	CreateTime time.Time
}
