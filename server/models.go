package server

import (
	"sync"
	"time"
)

const uArrayNumElements = 2

const uArrayUserName = 0

const uParamsTopic = 0

const uParamsContext = 1

const uParamsCategory = 2

const uParamsPublic = 3

const uParamsUsername = 0

const uParamsPassword = 1

type networkClient struct {
	userid         string
	username       string
	startTime      time.Time
	lastActiveTime time.Time
	sync.Mutex
}
