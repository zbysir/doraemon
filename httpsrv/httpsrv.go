package httpsrv

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/pires/go-proxyproto"
	"net"
	"net/http"
	"time"
)

// Service 支持 Start(ctx)
type Service struct {
	s       *http.Server
	options options
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
	httpServer := &http.Server{
		Addr: addr,
	}

	var opts options
	for _, o := range ot {
		o(&opts)
	}
	if opts.handler != nil {
		httpServer.Handler = opts.handler
	}
	if opts.tlsConfig != nil {
		httpServer.TLSConfig = opts.tlsConfig
	}

	return &Service{s: httpServer, options: opts}, nil
}

func (s *Service) Start(ctx context.Context) error {
	var err error
	go func() {
		select {
		case <-ctx.Done():
			err = s.s.Shutdown(ctx)
			if err != nil {
				return
			}
		}
	}()

	ln, err := net.Listen("tcp", s.s.Addr)
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
		err = s.s.ServeTLS(ln, "", "")
	} else {
		err = s.s.Serve(ln)
	}

	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}

	return nil
}
