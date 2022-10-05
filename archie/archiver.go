package archie

import (
	"github.com/minio/minio-go/v7"
	"sync"
)

type Archiver struct {
	DestBucket           string
	DestClient           *minio.Client
	DestName             string
	DestPartSize         uint64
	DestThreads          uint
	FetchDone            chan string
	HealthCheckEnabled   bool
	IsOffline            bool
	MsgTimeout           string
	SkipLifecycleExpired bool
	SrcBucket            string
	SrcClient            *minio.Client
	SrcName              string
	WaitGroup            *sync.WaitGroup
}

type AckType int

const (
	Ack AckType = iota
	Nak
	SkipAck
	Term
	FiveNakThenTerm
	None
)

func (s AckType) String() string {
	switch s {
	case Ack:
		return "ack"
	case Nak:
		return "nak"
	case Term:
		return "term"
	case FiveNakThenTerm:
		return "5nak_then_term"
	case None:
		return "none"
	}
	return "unknown"
}