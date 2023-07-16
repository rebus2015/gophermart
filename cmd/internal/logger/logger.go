package logger

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

const (
	LevelCritical = "crit"
	LevelError    = "error"
	LevelInfo     = "info"
	LevelDebug    = "debug"
)

const (
	ModeDevelopment = "dev"
	ModeProduction  = "prod"
)

type message struct {
	ctx    context.Context
	fields map[string]interface{}
}

func New(ctx context.Context) *message {
	return &message{ctx: ctx, fields: make(map[string]interface{})}
}

func (m *message) Field(key string, value interface{}) *message {
	m.fields[key] = value
	return m
}

func ContextFromRequest(r *http.Request) context.Context {
	ctx := r.Context()
	if requestID, ok := hlog.IDFromRequest(r); ok {
		getZerologLogger(ctx).With().Bytes("request_id", requestID.Bytes())
	}
	return ctx
}