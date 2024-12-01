package main

import (
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/propagation"
)

// NatsHeaderCarrier adapts the nats.Header to satisfy the TextMapCarrier interface.
type NatsHeaderCarrier nats.Header

var _ propagation.TextMapCarrier = NatsHeaderCarrier{}

// Get returns the value associated with the passed key.
func (hc NatsHeaderCarrier) Get(key string) string {
	return nats.Header(hc).Get(key)
}

// Set stores the key-value pair.
func (hc NatsHeaderCarrier) Set(key, value string) {
	nats.Header(hc).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (hc NatsHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	i := 0
	for k := range hc {
		keys[i] = k
		i++
	}
	return keys
}
