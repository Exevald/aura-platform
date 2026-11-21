// Package proxy owns HTTP/1, HTTP/2 and streaming transport details.
package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/distributed-programming-2026/gateway/internal/gateway/config"
)

type Proxy struct {
	http1   *http.Transport
	h2c     *http.Transport
	h2tls   *http.Transport
	timeout time.Duration
	idle    time.Duration
}

func New(c config.Gateway) *Proxy {
	newTransport := func(protocols *http.Protocols) *http.Transport {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.DisableCompression = true
		transport.ResponseHeaderTimeout = 0
		transport.Protocols = protocols
		return transport
	}
	http1Protocols := new(http.Protocols)
	http1Protocols.SetHTTP1(true)
	h2cProtocols := new(http.Protocols)
	h2cProtocols.SetUnencryptedHTTP2(true)
	h2tlsProtocols := new(http.Protocols)
	h2tlsProtocols.SetHTTP2(true)
	return &Proxy{
		http1:   newTransport(http1Protocols),
		h2c:     newTransport(h2cProtocols),
		h2tls:   newTransport(h2tlsProtocols),
		timeout: config.Value(c.ProxyTimeouts.Timeout, time.Minute),
		idle:    config.Value(c.ProxyTimeouts.IdleTimeout, 10*time.Minute),
	}
}

func (p *Proxy) Close() {
	p.http1.CloseIdleConnections()
	p.h2c.CloseIdleConnections()
	p.h2tls.CloseIdleConnections()
}

func (p *Proxy) Forward(w http.ResponseWriter, req *http.Request, route *config.Route, modify func(*http.Response) error, onError func(error)) {
	duration := config.Value(route.Timeout, p.timeout)
	ctx := req.Context()
	cancel := func() {}
	if duration > 0 {
		ctx, cancel = context.WithTimeout(ctx, duration)
	}
	defer cancel()
	incoming := req.WithContext(ctx)
	transport := http.RoundTripper(p.http1)
	if route.UseHTTP2 {
		if route.URL.Scheme == "https" {
			transport = p.h2tls
		} else {
			transport = p.h2c
		}
	}
	reverse := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.URL.Scheme = route.URL.Scheme
			r.Out.URL.Host = route.URL.Host
			r.Out.URL.RawQuery = r.In.URL.RawQuery
			r.Out.RequestURI = ""
			r.Out.Trailer = r.In.Trailer
			if !route.DisableHostRewrite {
				r.Out.Host = route.URL.Host
			}
			r.SetXForwarded()
		},
		Transport:      &idleTransport{base: responseTransport{base: transport}, idle: config.Value(route.IdleTimeout, p.idle)},
		ModifyResponse: func(response *http.Response) error { response.Request = incoming; return modify(response) },
		ErrorHandler:   func(_ http.ResponseWriter, _ *http.Request, err error) { onError(err) },
	}
	reverse.ServeHTTP(w, incoming)
}
