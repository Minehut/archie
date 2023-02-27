package archie

import (
	"archie/client"
	"go.arsenm.dev/pcre"
	"sync"
)

type Archiver struct {
	BackoffDurationMultiplier uint64
	BackoffNumCeiling         uint64
	DestBucket                string
	DestClient                client.Client
	DestName                  string
	DestPartSize              uint64
	DestThreads               uint
	FetchDone                 chan string
	HealthCheckDisabled       bool
	IsOffline                 bool
	MaxRetries                uint64
	MsgTimeout                string
	SkipEventBucketValidation bool
	SkipLifecycleExpired      bool
	SrcBucket                 string
	SrcClient                 client.Client
	SrcName                   string
	WaitForMatchingETag       bool
	WaitGroup                 *sync.WaitGroup
	ExcludePaths              struct {
		CopyObject   []*pcre.Regexp
		RemoveObject []*pcre.Regexp
	}
}

type AckType int

const (
	Ack AckType = iota
	Nak
	SkipAck
	Term
	NakThenTerm
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
	case NakThenTerm:
		return "nak_then_term"
	case None:
		return "none"
	}
	return "unknown"
}
