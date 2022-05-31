package gnet

import (
	"sync"
	"time"
)

//一共52bit
//ServerID:		8
//TimeSeconds:	30
//Sequence:		14
var (
	serverID    int64
	timeSeconds int64
	sequence    int64
	lock        sync.Mutex
)

const timeSecondsShift = 14
const serverIDShift = timeSecondsShift + 30
const maxSeq = ^(-1 << timeSecondsShift)
const maxServer = int64(^(-1 << 8))
const maxTime = int64(^(-1 << 30))

func getRelativeSeconds() int64 {
	return time.Now().Unix()
}

func createSessionID() int64 {
	lock.Lock()
	defer lock.Unlock()

	s := getRelativeSeconds()
	if timeSeconds != s {
		timeSeconds = s
		sequence = (serverID & maxServer << serverIDShift) | (timeSeconds & maxTime << timeSecondsShift)
	} else {
		sequence++
		// 达到seq最大等待下一秒
		if sequence&maxSeq == 0 {
			timeSeconds = waitNextSecond(s)
			sequence = (serverID & maxServer << serverIDShift) | (timeSeconds & maxTime << timeSecondsShift)
		}
	}
	return sequence
}

func waitNextSecond(last int64) int64 {
	current := getRelativeSeconds()
	for current <= last {
		current = getRelativeSeconds()
	}
	return current
}

func SetServerID(sid int) {
	serverID = int64(sid)
}
