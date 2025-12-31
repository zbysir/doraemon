package httpsrv

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/pires/go-proxyproto"
)

type Service struct {
	options options
	addr    string
}

type options struct {
	enableProxyProtocol bool
	handler             http.Handler
	tlsConfig           *tls.Config
}

type Option func(*options)

func WithEnableProxyProtocol(enable bool) Option {
	return func(o *options) {
		o.enableProxyProtocol = enable
	}
}

func WithHandler(handler http.Handler) Option {
	return func(o *options) {
		o.handler = handler
	}
}

func WithTls(tlsConfig *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = tlsConfig
	}
}

func NewService(addr string, ot ...Option) (*Service, error) {
	var opts options
	for _, o := range ot {
		o(&opts)
	}

	return &Service{addr: addr, options: opts}, nil
}

func (s *Service) Start(ctx context.Context) error {
	httpServer := &http.Server{
		Addr: s.addr,
		// baseContext can be used to allow request to get the ctx (global ctx has some data like i18n, etc.)
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	if s.options.handler != nil {
		httpServer.Handler = s.options.handler
	}
	if s.options.tlsConfig != nil {
		httpServer.TLSConfig = s.options.tlsConfig
	}

	var err error
	go func() {
		select {
		case <-ctx.Done():
			err = httpServer.Shutdown(ctx)
			if err != nil {
				return
			}
		}
	}()

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	if s.options.enableProxyProtocol {
		ln = &proxyproto.Listener{
			Listener:          ln,
			ReadHeaderTimeout: 10 * time.Second,
		}
	}

	if s.options.tlsConfig != nil {
		err = httpServer.ServeTLS(ln, "", "")
	} else {
		err = httpServer.Serve(ln)
	}

	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}

	return nil
}
