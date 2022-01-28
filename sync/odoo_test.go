package sync

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func newTestContext(t *testing.T) context.Context {
	zlogger := zaptest.NewLogger(t, zaptest.Level(zapcore.Level(-2)))
	return logr.NewContext(context.Background(), zapr.NewLogger(zlogger))
}
