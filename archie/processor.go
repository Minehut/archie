package archie

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

func (a *Archiver) MessageProcessor(baseCtx context.Context, msgCtx context.Context, sub *nats.Subscription, batchSize int) {
	msgTimeout, _ := time.ParseDuration(a.MsgTimeout)
	// ensure the message/transfer is canceled/terminated just before
	// the stream hits its timeout and the server re-queues the message
	msgTimeout = msgTimeout - (5 * time.Second)

	defer func() {
		log.Trace().Msg("Deferred message processor context canceled")
		a.WaitGroup.Done()
	}()
	a.WaitGroup.Add(1)

	for {
		// wait until both clients are online to fetch new messages from jetstream
		if a.SrcClient.IsOffline() || a.DestClient.IsOffline() {
			// only log on state change
			if a.IsOffline == false {
				if a.SrcClient.IsOffline() {
					log.Info().Msgf("Waiting while %s is offline", a.SrcClient.EndpointURL())
				} else {
					log.Info().Msgf("Waiting while %s is offline", a.DestClient.EndpointURL())
				}
			}
			a.IsOffline = true
			time.Sleep(time.Second * 10)
			continue
		}

		// source and dest must be online
		a.IsOffline = false

		// fetch will stop (error forever) if the context is canceled
		msgs, err := sub.Fetch(batchSize, nats.Context(baseCtx))
		if err != nil {
			if err == context.DeadlineExceeded {
				// no messages to fetch - try again later
				continue
			} else if err == context.Canceled {
				// base context canceled - graceful shutdown signal
				log.Info().Msg("Stopping event pull processing")
				a.FetchDone <- "graceful"
				return
			} else {
				log.Error().Err(err).Msg("Failed to fetch a new batch of JetStream messages")
				continue
			}
		}

		for _, msg := range msgs {
			// check for each message in the batch if we are processing more than one
			if batchSize > 1 && checkContextDone(baseCtx) {
				log.Info().Msg("Stopping event pull processing")
				a.FetchDone <- "graceful"
				return
			}

			// wrap this so we can use defer()
			func() {
				// create message sub-context with a timeout
				// that will cancel the transfer immediately
				perMsgCtx, perMsgCancel := context.WithTimeout(msgCtx, msgTimeout)
				defer func() {
					log.Trace().Msg("Deferred individual message context canceled")
					perMsgCancel()
				}()

				// main message func
				a.message(perMsgCtx, msg)
			}()
		}
	}
}
