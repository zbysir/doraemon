package log

import (
	"bytes"
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
)

func Debugf(template string, args ...interface{}) { globalSugaredLoggerSkip1.Debugf(template, args...) }

func Infof(template string, args ...interface{})  { globalSugaredLoggerSkip1.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { globalSugaredLoggerSkip1.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { globalSugaredLoggerSkip1.Errorf(template, args...) }
func Fatalf(template string, args ...interface{}) { globalSugaredLoggerSkip1.Fatalf(template, args...) }
func Panicf(template string, args ...interface{}) { globalSugaredLoggerSkip1.Panicf(template, args...) }

type fieldCtxKey struct{}

func WithFieldContext(ctx context.Context, fields ...any) context.Context {
	if f, ok := ctx.Value(fieldCtxKey{}).([]any); ok {
		fields = append(f, fields...)
	}
	ctx = context.WithValue(ctx, fieldCtxKey{}, fields)
	return ctx
}

func WithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return globalSugaredLogger
	}
	if f, ok := ctx.Value(fieldCtxKey{}).([]any); ok {
		return globalSugaredLogger.With(f...)
	}
	return globalSugaredLogger
}

var globalSugaredLogger *zap.SugaredLogger
var globalSugaredLoggerSkip1 *zap.SugaredLogger // for Infof/Errorf in outermost

func init() {
	SetGlobalLogger(New(Options{
		To:            nil,
		DisableTime:   false,
		DisableCaller: false,
		CallerSkip:    0,
		Name:          "",
	}))
}

func SetGlobalLogger(o *zap.Logger) {
	globalSugaredLoggerSkip1 = o.WithOptions(zap.AddCallerSkip(1)).Sugar() // addCallerSkip for Infof/Errorf in outermost
	globalSugaredLogger = o.Sugar()
}

type BuffSink struct {
	buf bytes.Buffer
}

func (b *BuffSink) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

func (b *BuffSink) Sync() error {
	return nil
}

func (b *BuffSink) Close() error {
	return nil
}

type Options struct {
	LogLeave      zapcore.Level
	To            io.Writer
	DisableTime   bool
	DisableLevel  bool
	DisableCaller bool
	TimeLayout    string
	CallerSkip    int
	Name          string
}

func New(o Options) *zap.Logger {
	zapconfig := zap.NewDevelopmentConfig()
	if o.LogLeave != 0 {
		zapconfig.Level.SetLevel(o.LogLeave)
	}

	zapconfig.OutputPaths = []string{"stdout"} // default is stderr

	if o.DisableTime {
		zapconfig.EncoderConfig.EncodeTime = nil
	} else {
		if o.TimeLayout == "" {
			o.TimeLayout = "2006/01/02 15:04:05"
		}
		zapconfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(o.TimeLayout)
	}
	if o.DisableLevel {
		zapconfig.EncoderConfig.EncodeLevel = nil
	} else {
		zapconfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var ops []zap.Option

	ops = append(ops)
	if o.CallerSkip != 0 {
		ops = append(ops, zap.AddCallerSkip(o.CallerSkip))
	}

	var sink zapcore.WriteSyncer
	if o.To == nil {
		var err error
		var closeOut func()
		sink, closeOut, err = zap.Open(zapconfig.OutputPaths...)
		if err != nil {
			panic(err)

		}
		errSink, _, err := zap.Open(zapconfig.ErrorOutputPaths...)
		if err != nil {
			closeOut()
			panic(err)
		}

		ops = append(ops, zap.ErrorOutput(errSink))
	} else {
		sink = zapcore.AddSync(o.To)
	}
	if !o.DisableCaller {
		ops = append(ops, zap.AddCaller())
	}
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zapconfig.EncoderConfig), sink, zapconfig.Level), ops...)

	if o.Name != "" {
		logger = logger.Named(o.Name)
	}
	return logger
}
