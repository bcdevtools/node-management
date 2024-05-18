package types

import (
	"net"
	"time"
)

type AddrBook struct {
	Key   string          `json:"key"`
	Addrs []*KnownAddress `json:"addrs"`
}

type KnownAddress struct {
	Addr        *NetAddress `json:"addr"`
	Src         *NetAddress `json:"src"`
	Buckets     []int       `json:"buckets"`
	Attempts    int32       `json:"attempts"`
	BucketType  byte        `json:"bucket_type"`
	LastAttempt time.Time   `json:"last_attempt"`
	LastSuccess time.Time   `json:"last_success"`
	LastBanTime time.Time   `json:"last_ban_time"`
}

type NetAddress struct {
	ID   NetAddressID `json:"id"`
	IP   net.IP       `json:"ip"`
	Port uint16       `json:"port"`
}

type NetAddressID string
