package client

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

func JetStream(url, subject, stream, durableConsumer, streamMaxAge, rootCA, username, password string, streamReplicas, maxAckPending int, streamMaxMBytes int64, msgTimeout string, jetreamStreamRePublishEnabled bool) (*nats.Subscription, *nats.Conn) {
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

	// TODO: output some information about the JetStream server

	// build the stream
	maxAge, err := time.ParseDuration(streamMaxAge)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse jetstream-max-age duration argument")
	}

	maxBytes := int64(-1)
	if streamMaxMBytes != -1 {
		maxBytes = 1024 * 1024 * streamMaxMBytes
	}

	streamConfig := &nats.StreamConfig{
		Name:      stream,
		Subjects:  []string{subject},
		MaxAge:    maxAge,
		MaxBytes:  maxBytes,       // TODO: test this
		Replicas:  streamReplicas, // TODO: test this
		Retention: nats.LimitsPolicy,
	}

	if jetreamStreamRePublishEnabled {
		// set the main stream to delete msgs on ACK
		// then use the archive stream as a backup
		streamConfig.Retention = nats.InterestPolicy

		// setup the msg forwarding ot the archive stream
		streamConfig.RePublish = &nats.RePublish{Source: subject, Destination: fmt.Sprintf("%s-archive", subject)}

		// build an archive stream that will not be consumed
		archiveStreamConfig := &nats.StreamConfig{
			Name:      fmt.Sprintf("%s-archive", stream),
			Subjects:  []string{fmt.Sprintf("%s-archive", subject)},
			MaxAge:    maxAge,
			MaxBytes:  maxBytes,       // TODO: test this
			Replicas:  streamReplicas, // TODO: test this
			Retention: nats.LimitsPolicy,
		}

		archiveStreamInfo := createOrUpdateStream(jetStream, fmt.Sprintf("%s-archive", stream), archiveStreamConfig)

		log.Info().Msgf("JetStream archive stream %s configured with %d replicas and limited by %s max age, and %d max bytes",
			archiveStreamInfo.Config.Name, archiveStreamInfo.Config.Replicas, archiveStreamInfo.Config.MaxAge, archiveStreamInfo.Config.MaxBytes)
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
		//MaxWaiting:    jetStreamMaxWaiting,    // must match fetch() batch parameter
		//MaxRequestExpires: 60 * time.Second,   // limit max fetch() expires    (pull)
		//MaxRequestBatch:   10,                 // limit max fetch() batch size (pull)
		AckPolicy:     nats.AckExplicitPolicy, // always ack (default)
		AckWait:       ackWait,                // wait before retry
		DeliverPolicy: nats.DeliverNewPolicy,  // deliver since consumer creation
		Durable:       durableConsumer,        // consumer name
		FilterSubject: subject,                // nats subject for stream
		MaxAckPending: maxAckPending,          // stop offering msgs once we're waiting on too many acks
		MaxDeliver:    -1,                     // try to redeliver forever
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
