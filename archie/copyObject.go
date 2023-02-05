package archie

import (
	"archie/client"
	"archie/event"
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"time"
)

func (a *Archiver) copyObject(ctx context.Context, mLog zerolog.Logger, eventObjKey string, msg *nats.Msg, record event.Record) (error, string, AckType) {
	metadata, _ := msg.Metadata()

	for _, excludedPathRegexp := range a.ExcludePaths.CopyObject {
		if excludedPathRegexp.MatchString(eventObjKey) {
			mLog.Info().
				Uint64("numDelivered", metadata.NumDelivered).
				Str("queueDuration", time.Now().Sub(metadata.Timestamp).String()).
				Str("pattern", excludedPathRegexp.String()).
				Msg("Excluded path match, copy event skipped")

			a.observeMessagesTransferNumDeliveredMetric(float64(metadata.NumDelivered))
			a.observeMessagesTransferQueueDurationMetric(time.Now().Sub(metadata.Timestamp).Seconds())

			return nil, "EXCLUDED_PATH", SkipAck
		}
	}

	// get src object
	start := time.Now()
	srcObject, err := a.SrcClient.GetObject(ctx, a.SrcBucket, eventObjKey)
	if err != nil {
		return err, "Failed to GetObject from the source bucket", Nak
	}

	// get source size, the event's object size wasn't good enough
	srcStat, err := srcObject.Stat(ctx)
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			// minio error
			return err, "Failed to Stat the source object", NakThenTerm
		} else if err.Error() == "storage: object doesn't exist" {
			// gcs error
			return err, "Failed to Stat the source object", NakThenTerm
		} else {
			return err, "Failed to Stat the source object", Nak
		}
	}

	if a.WaitForMatchingETag {
		if srcStat.ETag != record.S3.Object.ETag {
			return fmt.Errorf(
				"mismatch of ETags from the event (%s) and source (%s)", record.S3.Object.ETag, srcStat.ETag,
			), "ETag mismatch", Nak
		}
	}

	mLog.Info().
		Int64("size", srcStat.Size).
		Str("hSize", size(srcStat.Size)).
		Msg("Transfer started")

	// put dest object
	destPartSizeBytes := 1024 * 1024 * a.DestPartSize
	putOpts := client.PutOptions{
		ContentType: srcStat.ContentType,
		NumThreads:  a.DestThreads,
		PartSize:    destPartSizeBytes,
	}

	if record.S3.Object.ETag != "" {
		putOpts.ETag = record.S3.Object.ETag
	}

	start = time.Now()
	_, err = a.DestClient.PutObject(ctx, a.DestBucket, eventObjKey, srcObject.GetReader(), srcStat.Size, putOpts)
	if err != nil {
		return err, "Failed to PutObject to the destination bucket", Nak
	}

	// measure transfer time
	putElapsed := time.Now().Sub(start)

	// find how much time was spent in the queue
	totalTime := time.Now().Sub(metadata.Timestamp)
	queueDuration := totalTime - putElapsed

	mLog.Info().
		Int64("size", srcStat.Size).
		Str("hSize", size(srcStat.Size)).
		Str("transferDuration", putElapsed.String()).
		Str("rate", rate(srcStat.Size, putElapsed.Seconds())).
		Uint64("numDelivered", metadata.NumDelivered).
		Str("queueDuration", queueDuration.String()).
		Msg("Transfer complete")

	// successful transfer metrics
	a.observeMessagesTransferDurationMetric(putElapsed.Seconds())
	a.observeMessagesTransferRateMetric(float64(srcStat.Size) / putElapsed.Seconds())
	a.observeMessagesTransferSizeMetric(float64(srcStat.Size))
	a.observeMessagesTransferNumDeliveredMetric(float64(metadata.NumDelivered))
	a.observeMessagesTransferQueueDurationMetric(queueDuration.Seconds())

	return nil, "", Ack
}
