package archie

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"time"
)

func (a *Archiver) copyObject(ctx context.Context, mLog zerolog.Logger, eventObjKey string, msg *nats.Msg) (error, string, AckType) {
	metadata, _ := msg.Metadata()

	// get src object
	start := time.Now()
	srcObject, err := a.SrcClient.GetObject(ctx, a.SrcBucket, eventObjKey, minio.GetObjectOptions{})
	if err != nil {
		return err, "Failed to GetObject from the source bucket", Nak
	}

	// get source size, the event's source size was inaccurate sometimes
	srcStat, err := srcObject.Stat()
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			return err, "Failed to Stat the source object", FiveNakThenTerm
		} else {
			return err, "Failed to Stat the source object", Nak
		}
	}

	mLog.Info().
		Int64("size", srcStat.Size).
		Str("hSize", size(srcStat.Size)).
		Msg("Transfer started")

	// put dest object
	destPartSize := 1024 * 1024 * a.DestPartSize
	putOpts := minio.PutObjectOptions{
		ContentType: srcStat.ContentType,
		NumThreads:  a.DestThreads,
		PartSize:    destPartSize,
	}

	start = time.Now()
	_, err = a.DestClient.PutObject(ctx, a.DestBucket, eventObjKey, srcObject, srcStat.Size, putOpts)
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