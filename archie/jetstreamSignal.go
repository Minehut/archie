package archie

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"math"
	"time"
)

// server confirmed ack (double ack)
func sendAckSignal(msg *nats.Msg, mLog *zerolog.Logger) error {
	// TODO: make this timeout configurable?
	err := msg.AckSync(nats.AckWait(2 * time.Second))
	if err != nil {
		mLog.Error().Err(err).Msg("Failed to complete JetStream Ack signal")
		sendNakSignal(msg, mLog)
		return err
	}
	return nil
}

// terminate message redelivery
func sendTermSignal(msg *nats.Msg, mLog *zerolog.Logger) error {
	mLog.Info().Msgf("Sending JetStream Term signal to stop redelivery")

	err := msg.Term()
	if err != nil {
		mLog.Error().Err(err).Msg("Failed to complete JetStream Term signal")
		sendNakSignal(msg, mLog)
		return err
	}
	return nil
}

// msgs will continue to redeliver via with this exponential backoff
// if the Nak fails just let jetstream redeliver after its timeout
func sendNakSignal(msg *nats.Msg, mLog *zerolog.Logger) {
	metadata, _ := msg.Metadata()
	var maxDelivered uint64 = 15 // 54m36.6s

	numDelivered := maxDelivered
	if metadata.NumDelivered < maxDelivered {
		numDelivered = metadata.NumDelivered
	}

	delay := time.Duration(int64(math.Pow(2, float64(numDelivered)))) * 100 * time.Millisecond

	mLog.Info().Msgf("Sending JetStream NAck signal and requesting redelivery in %s", delay.String())

	err := msg.NakWithDelay(delay)
	if err != nil {
		mLog.Error().Err(err).Msg("Failed to complete JetStream NAck signal")
	}
}
