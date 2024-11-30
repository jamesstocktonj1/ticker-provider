package main

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	wrpcnats "wrpc.io/go/nats"
)

var _ propagation.TextMapCarrier = &wrpcProp{}

type wrpcProp struct {
	nats.Header
}

func (p *wrpcProp) Keys() []string {
	keys := []string{}
	for k := range p.Header {
		keys = append(keys, k)
	}
	return keys
}

func injectTraceHeader(_ctx context.Context) context.Context {
	carrier := &wrpcProp{nats.Header{}}
	otel.GetTextMapPropagator().Inject(_ctx, carrier)
	return wrpcnats.ContextWithHeader(_ctx, carrier.Header)
}
