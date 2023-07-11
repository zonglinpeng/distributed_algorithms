package raft

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	rand "math/rand"
	"time"
)

func (r *Raft) initRand() *rand.Rand {
	bs := make([]byte, 8)
	_, err := crypto_rand.Read(bs)
	if err != nil {
		r.logger.Errorf("init rand with cryptographically secure random number generator failed, try use unixnano time source")
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(bs))))
}

func (r *Raft) randomTimeout(min int, max int) time.Duration {
	randomNumInRange := r.rand.Intn(max-min) + min
	return time.Millisecond * time.Duration(randomNumInRange)
}

func (r *Raft) randomElectionTimeout() time.Duration {
	return r.randomTimeout(ElectionTimeoutMin, ElectionTimeoutMax)
}

func minInt(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
