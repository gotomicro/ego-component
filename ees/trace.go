package ees

import (
	"net/http"

	"github.com/gotomicro/ego/core/etrace"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Transport for tracing Elastic operations.
type Transport struct {
	rt http.RoundTripper
}

// TransportOption signature for specifying options, e.g. WithRoundTripper.
type TransportOption func(t *Transport)

// WithRoundTripper specifies the http.RoundTripper to call
// next after this transport. If it is nil (default), the
// transport will use http.DefaultTransport.
func WithRoundTripper(rt http.RoundTripper) TransportOption {
	return func(t *Transport) {
		t.rt = rt
	}
}

// NewTransport specifies a transport that will trace Elastic
// and report back via OpenTracing.
func NewTransport(opts ...TransportOption) *Transport {
	t := &Transport{}
	for _, o := range opts {
		o(t)
	}
	return t
}

// RoundTrip captures the request and starts an OpenTracing span
// for Elastic PerformRequest operation.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	tracer := etrace.NewTracer(trace.SpanKindClient)
	ctx, span := tracer.Start(req.Context(), "PerformRequest", nil)
	req = req.WithContext(ctx)
	defer span.End()

	span.SetAttributes(
		etrace.String("peer.service", "elasticsearch"),
		etrace.String("db.system", "elasticsearch"),
		etrace.String("http.method", req.Method),
		etrace.String("http.url", req.URL.String()),
		etrace.String("net.peer.name", req.URL.Hostname()),
		etrace.String("net.peer.port", req.URL.Port()),
	)

	var (
		resp *http.Response
		err  error
	)
	if t.rt != nil {
		resp, err = t.rt.RoundTrip(req)
	} else {
		resp, err = http.DefaultTransport.RoundTrip(req)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	if resp != nil {
		span.SetAttributes(
			etrace.String("http.status_code", cast.ToString(resp.StatusCode)),
		)
	}

	return resp, err
}
