package archie

import (
	evt "archie/event"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"net/url"
	"strings"
)

func (a *Archiver) message(ctx context.Context, msg *nats.Msg) {
	aLog := log.With().Logger()

	metadata, err := msg.Metadata()
	if err != nil {
		log.Error().Msg("Failed to retrieve metadata from the event message")
		sendNakSignal(msg, &aLog)
		return
	}

	msgMetadata, err := json.Marshal(metadata)
	if err != nil {
		log.Error().Msg("Failed to marshal metadata to json")
		sendNakSignal(msg, &aLog)
		return
	}

	event := evt.Minio{}
	err = json.Unmarshal(msg.Data, &event)
	if err != nil {
		errMsg := "Failed to unmarshal raw event payload"
		if isJSON(msg.Data) {
			log.Error().RawJSON("metadata", msgMetadata).RawJSON("payload", msg.Data).Err(err).Msg(errMsg)
		} else {
			log.Error().RawJSON("metadata", msgMetadata).Str("payload", string(msg.Data)).Err(err).Msg(errMsg)
		}
		sendNakSignal(msg, &aLog)
		return
	}

	log.Debug().RawJSON("metadata", msgMetadata).RawJSON("payload", msg.Data).Msg("Message received - Raw")

	// parse top-level
	eventBucket, eventKey := parseEventPath(event.Key)
	eventType := strings.Join(strings.Split(event.EventName, ":")[0:2], ":")

	// per-message logger
	mLog := log.With().Str("key", eventKey).Str("event", event.EventName).Uint64("seq", metadata.Sequence.Stream).Logger()

	// validation
	err = a.validateEventName(event)
	if err != nil {
		mLog.Error().Err(err).Msg("Failed to validate the event name")
		err = sendTermSignal(msg, &mLog)
		if err != nil {
			// logging already happened
			return
		}
		a.cleanupAndCountMessagesProcessedMetric("terminated", "Event name not in list of valid events", "INT_TERM_INVALID_EVENT_NAME", event.EventName, eventType)
		return
	}

	err = a.validateEventBucket(eventBucket)
	if err != nil {
		mLog.Error().Err(err).Msg("Failed to validate the event bucket")
		err = sendTermSignal(msg, &mLog)
		if err != nil {
			// logging already happened
			return
		}
		a.cleanupAndCountMessagesProcessedMetric("terminated", "Event bucket and config bucket do not match", "INT_TERM_INVALID_EVENT_BUCKET", event.EventName, eventType)
		return
	}

	for _, eventRecord := range event.Records {
		// object key in the event record needs url decode
		eventObjKey, err := url.QueryUnescape(eventRecord.S3.Object.Key)
		if err != nil {
			mLog.Error().Err(err).Msg("Failed to unescape source object key from event")
			sendNakSignal(msg, &mLog)
			continue
		}

		mLog.Info().
			Str("eventBucket", eventRecord.S3.Bucket.Name).
			Str("srcBucket", a.SrcBucket).
			Str("destBucket", a.DestBucket).
			Str("etag", eventRecord.S3.Object.ETag).
			Int64("bytes", eventRecord.S3.Object.Size).
			Uint64("numDelivered", metadata.NumDelivered).
			Str("sourceHost", eventRecord.Source.Host).
			Msg("Message received")

		var ack AckType
		var s3ErrMsg, s3ErrCode, execContext string

		// message type router
		switch eventType {
		case "s3:ObjectCreated":
			err, execContext, ack = a.copyObject(ctx, mLog, eventObjKey, msg)
			if err != nil {
				s3ErrMsg, s3ErrCode = logS3Error(err, execContext, &mLog)
			}
		case "s3:ObjectRemoved":
			err, execContext, ack = a.removeObject(ctx, mLog, eventObjKey, msg, eventRecord)
			if err != nil {
				s3ErrMsg, s3ErrCode = logS3Error(err, execContext, &mLog)
			}
		default:
			mLog.Error().Msgf("Unable to process the %s event type", event.EventName, eventType)
			ack = Nak
		}

		// ack router with metrics
		switch ack {
		case Ack:
			err = sendAckSignal(msg, &mLog)
			if err != nil {
				// logging already happened
				continue
			}
			a.cleanupAndCountMessagesProcessedMetric("success", "", "", event.EventName, eventType)
		case SkipAck:
			err = sendAckSignal(msg, &mLog)
			if err != nil {
				// logging already happened
				continue
			}
			a.cleanupAndCountMessagesProcessedMetric("skipped", "", execContext, event.EventName, eventType)
		case Nak:
			sendNakSignal(msg, &mLog)
			a.cleanupAndCountMessagesProcessedMetric("failed", s3ErrMsg, s3ErrCode, event.EventName, eventType)
		case FiveNakThenTerm:
			maxDelivered := uint64(5)
			if metadata.NumDelivered > maxDelivered {
				mLog.Error().Uint64("numDelivered", metadata.NumDelivered).Msg("Reached max delivered")
				termErr := sendTermSignal(msg, &mLog)
				if termErr != nil {
					// logging already happened
					continue
				}
				a.cleanupAndCountMessagesProcessedMetric("terminated", fmt.Sprintf("Term %s", err.Error()), "INT_TERM5", event.EventName, eventType)
			} else {
				sendNakSignal(msg, &mLog)
				a.cleanupAndCountMessagesProcessedMetric("failed", s3ErrMsg, s3ErrCode, event.EventName, eventType)
			}
		case Term:
			termErr := sendTermSignal(msg, &mLog)
			if termErr != nil {
				// logging already happened
				continue
			}
			a.cleanupAndCountMessagesProcessedMetric("terminated", termErr.Error(), "INT_TERM", event.EventName, eventType)
		case None:
			continue
		default:
			mLog.Error().Msgf("Unable to process the %s ack type", ack)
			sendNakSignal(msg, &mLog)
			a.cleanupAndCountMessagesProcessedMetric("failed", fmt.Sprintf("Unable to process %s ack type", ack), s3ErrCode, event.EventName, eventType)
			continue
		}
	}
}
