package archie

import (
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"math"
	"time"
)

// server confirmed ack (double ack)
func sendAckSignal(msg *nats.Msg, mLog *zerolog.Logger) error {
	attempt := 1
	maxAttempt := 4

	for attempt <= maxAttempt {
		// TODO: make this timeout configurable
		err := msg.AckSync(nats.AckWait(3 * time.Second))
		if err != nil {
			if attempt == maxAttempt {
				mLog.Error().Err(err).Int("attempt", attempt).Msg("Reached maximum attempt")
				return err
			} else {
				mLog.Error().Err(err).
					Int("attempt", attempt).
					Int("maxAttempt", maxAttempt).
					Msg("Failed to complete JetStream Ack signal")
				time.Sleep(1 * time.Second)
				attempt++
				continue
			}
		}
		return nil
	}
	return errors.New("unknown error")
}

// terminate message redelivery
func sendTermSignal(msg *nats.Msg, mLog *zerolog.Logger) error {
	mLog.Info().Msgf("Sending JetStream Term signal to stop redelivery")

	err := msg.Term()
	if err != nil {
		mLog.Error().Err(err).Msg("Failed to complete JetStream Term signal")
		return err
	}
	return nil
}

// msgs will continue to redeliver via this exponential backoff,
// if the Nak fails just let jetstream redeliver after its timeout
func sendNakSignal(msg *nats.Msg, mLog *zerolog.Logger) {
	metadata, _ := msg.Metadata()
	var numDeliveredLimit uint64 = 15 // 54m36.6s

	numDeliveredPower := numDeliveredLimit
	if metadata.NumDelivered < numDeliveredLimit {
		numDeliveredPower = metadata.NumDelivered
	}

	delay := time.Duration(int64(math.Pow(2, float64(numDeliveredPower)))) * 100 * time.Millisecond

	mLog.Info().Msgf("Sending JetStream NAck signal and requesting redelivery in %s", delay.String())

	err := msg.NakWithDelay(delay)
	if err != nil {
		mLog.Error().Err(err).Msg("Failed to complete JetStream NAck signal")
	}
}
