package archie

import (
	"archie/event"
	"context"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"time"
)

func (a *Archiver) removeObject(ctx context.Context, mLog zerolog.Logger, eventObjKey string, msg *nats.Msg, record event.Record) (error, string, AckType) {
	metadata, _ := msg.Metadata()

	if a.SkipLifecycleExpired && (record.Source.Host == "Internal: [ILM-EXPIRY]" || record.Source.UserAgent == "Internal: [ILM-EXPIRY]") {
		mLog.Info().
			Uint64("numDelivered", metadata.NumDelivered).
			Str("queueDuration", time.Now().Sub(metadata.Timestamp).String()).
			Msg("Lifecycle expiration event skipped")

		a.observeMessagesDeleteNumDeliveredMetric(float64(metadata.NumDelivered))
		a.observeMessagesDeleteQueueDurationMetric(time.Now().Sub(metadata.Timestamp).Seconds())

		return nil, "ILM_EXPIRY", SkipAck
	}

	for _, excludedPathRegexp := range a.ExcludePaths.RemoveObject {
		if excludedPathRegexp.MatchString(eventObjKey) {
			mLog.Info().
				Uint64("numDelivered", metadata.NumDelivered).
				Str("queueDuration", time.Now().Sub(metadata.Timestamp).String()).
				Str("pattern", excludedPathRegexp.String()).
				Msg("Excluded path match, remove event skipped")

			a.observeMessagesDeleteNumDeliveredMetric(float64(metadata.NumDelivered))
			a.observeMessagesDeleteQueueDurationMetric(time.Now().Sub(metadata.Timestamp).Seconds())

			return nil, "EXCLUDED_PATH", SkipAck
		}
	}

	start := time.Now()

	err := a.DestClient.RemoveObject(ctx, a.DestBucket, eventObjKey)
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			// minio error
			return err, "Failed to RemoveObject from destination bucket", NakThenTerm
		} else if err.Error() == "storage: object doesn't exist" {
			// gcs error
			return err, "Failed to RemoveObject from destination bucket", NakThenTerm
		} else {
			return err, "Failed to RemoveObject from destination bucket", Nak
		}
	}

	// measure delete time
	deleteElapsed := time.Now().Sub(start)

	// find how much time was spent in the queue
	totalTime := time.Now().Sub(metadata.Timestamp)
	queueDuration := totalTime - deleteElapsed

	mLog.Info().
		Str("deleteDuration", deleteElapsed.String()).
		Uint64("numDelivered", metadata.NumDelivered).
		Str("queueDuration", queueDuration.String()).
		Msg("Delete complete")

	// successful delete metrics
	a.observeMessagesDeleteDurationMetric(deleteElapsed.Seconds())
	a.observeMessagesDeleteNumDeliveredMetric(float64(metadata.NumDelivered))
	a.observeMessagesDeleteQueueDurationMetric(queueDuration.Seconds())

	return nil, "", Ack
}
