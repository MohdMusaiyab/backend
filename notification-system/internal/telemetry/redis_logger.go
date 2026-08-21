package telemetry

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisPubSubWriter implements io.Writer. 
// It allows us to stream standard Go logs directly into a Redis Pub/Sub channel.
type RedisPubSubWriter struct {
	client  *redis.Client
	channel string
}

func NewRedisPubSubWriter(client *redis.Client, channel string) *RedisPubSubWriter {
	return &RedisPubSubWriter{
		client:  client,
		channel: channel,
	}
}

func (r *RedisPubSubWriter) Write(p []byte) (n int, err error) {
	// Publish the log payload to Redis so the API Gateway can pick it up.
	// We ignore errors here so a Redis blip doesn't crash the worker's execution.
	r.client.Publish(context.Background(), r.channel, string(p))
	return len(p), nil
}
