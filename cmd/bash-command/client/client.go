package client

import "sync"

var (
	hostManagerSets = &sync.Map{}
)
