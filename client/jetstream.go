package client

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

func JetStream(
	url, subject, stream, durableConsumer, streamMaxAgeDur, rootCA, username, password string,
	streamReplicas, maxAckPending int,
	streamMaxSize int64,
	msgTimeout, streamRetention, streamRepublishSubject string,
	provisioningDisabled bool,
) (*nats.Subscription, *nats.Conn) {
	var connectOptions []nats.Option
	if rootCA != "" {
		connectOptions = append(connectOptions, nats.RootCAs(rootCA))
	}
	if username != "" && password != "" {
		connectOptions = append(connectOptions, nats.UserInfo(username, password))
	}

	natsClient, err := nats.Connect(url, connectOptions...)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup JetStream client")
	}

	log.Info().Msgf("Connected to nats at %s", natsClient.ConnectedUrl())

	jetStream, err := natsClient.JetStream()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize JetStream context")
	}

	accountInfo, err := jetStream.AccountInfo()
	log.Info().Uint64("memory", accountInfo.Tier.Memory).
		Uint64("storage", accountInfo.Tier.Store).
		Int("streams", accountInfo.Tier.Streams).
		Int("consumers", accountInfo.Tier.Consumers).
		Msg("JetStream account info")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get JetStream account info")
	}

	if !provisioningDisabled {
		// build the stream
		streamMaxBytes := int64(-1)
		if streamMaxSize != -1 {
			streamMaxBytes = 1_000_000 * streamMaxSize // Megabytes
		}

		streamConfig := &nats.StreamConfig{
			Name:      stream,
			Subjects:  []string{subject},
			MaxBytes:  streamMaxBytes,
			Replicas:  streamReplicas,
			Retention: nats.LimitsPolicy,
		}

		if streamMaxAgeDur != "" {
			streamMaxAge, err := time.ParseDuration(streamMaxAgeDur)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to parse jetstream-max-age duration argument")
			}
			streamConfig.MaxAge = streamMaxAge
		}

		if streamRetention == "interest" {
			streamConfig.Retention = nats.InterestPolicy
		} else if streamRetention == "work_queue" {
			streamConfig.Retention = nats.WorkQueuePolicy
		}

		if streamRepublishSubject != "" {
			streamConfig.RePublish = &nats.RePublish{Source: subject, Destination: streamRepublishSubject}
		}

		streamInfo := createOrUpdateStream(jetStream, stream, streamConfig)

		log.Info().Msgf("JetStream stream %s configured with %d replicas and limited by %s max age, and %d max bytes",
			streamInfo.Config.Name, streamInfo.Config.Replicas, streamInfo.Config.MaxAge, streamInfo.Config.MaxBytes)

		// build the stream consumer
		ackWait, err := time.ParseDuration(msgTimeout)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse jetstream-msg-timeout duration argument")
		}

		desiredConsumerConfig := &nats.ConsumerConfig{
			AckPolicy:       nats.AckExplicitPolicy, // always ack (default)
			AckWait:         ackWait,                // wait before retry
			DeliverPolicy:   nats.DeliverNewPolicy,  // deliver since consumer creation
			Durable:         durableConsumer,        // consumer name
			FilterSubject:   subject,                // nats subject for stream
			MaxAckPending:   maxAckPending,          // stop offering msgs once we're waiting on too many acks
			MaxDeliver:      -1,                     // try to redeliver forever
			SampleFrequency: "100",                  // deliver all messages as a percentage, required by tf module
		}

		consumerInfo, err := jetStream.ConsumerInfo(stream, durableConsumer)
		if err != nil {
			if err.Error() == "nats: consumer not found" {
				consumerInfo, err = jetStream.AddConsumer(stream, desiredConsumerConfig)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to add JetStream consumer")
				}
			} else {
				log.Fatal().Err(err).Msg("Failed to get JetStream consumer info")
			}
		} else {
			activeConsumerConfig := consumerInfo.Config

			if desiredConsumerConfig != &activeConsumerConfig {
				consumerInfo, err = jetStream.UpdateConsumer(stream, desiredConsumerConfig)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to update JetStream consumer")
				}
			}
		}

		log.Info().Msgf("JetStream Consumer %s configured with %s message timeout and %d max ack pending",
			consumerInfo.Config.Durable, consumerInfo.Config.AckWait, consumerInfo.Config.MaxAckPending)
	}

	// pull mode consumer
	sub, err := jetStream.PullSubscribe(subject, durableConsumer)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to JetStream")
	}

	log.Info().Msgf("Subscribed to JetStream consumer %s on subject %s", durableConsumer, subject)

	return sub, natsClient
}

func createOrUpdateStream(jetStream nats.JetStreamContext, stream string, streamConfig *nats.StreamConfig) *nats.StreamInfo {
	streamInfo, err := jetStream.StreamInfo(stream)
	if err != nil {
		if err.Error() == "nats: stream not found" {
			streamInfo, err = jetStream.AddStream(streamConfig)
			if err != nil {
				log.Fatal().Err(err).Msgf("Failed to add the JetStream stream %s", stream)
			}
		} else {
			log.Fatal().Err(err).Msgf("Failed to get JetStream stream %s info", stream)
		}
	} else {
		streamInfo, err = jetStream.UpdateStream(streamConfig)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to update the JetStream stream %s", stream)
		}
	}
	return streamInfo
}
