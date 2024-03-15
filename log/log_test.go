package log

import (
	"context"
	"testing"
)

func TestColor(t *testing.T) {
	//SetDev(true)
	buf := BuffSink{}
	l := New(Options{
		To:            &buf,
		DisableCaller: false,
		CallerSkip:    0,
		Name:          "",
	}).Sugar()

	l.Infof("%v", "info")
	l.Debugf("%v", "debug")
	l.Warnf("%v", "warn")
	l.Errorf("%v", "error")

	t.Logf("buf: %s", buf.buf.String())
}

func TestFormat(t *testing.T) {
	l := New(Options{
		DisableCaller: false,
		CallerSkip:    0,
		Name:          "",
	}).Sugar()

	l.Infof("123")
}

func TestTestContext(t *testing.T) {
	ctx := WithFieldContext(context.Background(), "traceId", "bbb")
	WithContext(ctx).Infof("%v", "info")
}
