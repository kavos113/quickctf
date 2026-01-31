package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"connectrpc.com/connect"
)

type ConnectLoggingInterceptor struct {
	service string
}

func NewConnectLoggingInterceptor(service string) *ConnectLoggingInterceptor {
	return &ConnectLoggingInterceptor{service: service}
}

func (l *ConnectLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()

		clientIP := req.Header().Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = req.Header().Get("X-Real-IP")
		}
		userAgent := req.Header().Get("User-Agent")

		resp, err := next(ctx, req)

		duration := time.Since(start)
		statusCode := "OK"
		if err != nil {
			if connectErr, ok := err.(*connect.Error); ok {
				statusCode = connectErr.Code().String()
			} else {
				statusCode = "unknown"
			}
		}

		attrs := []any{
			slog.String("service", l.service),
			slog.String("method", req.Spec().Procedure),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.Time("start_time", start),
			slog.Duration("duration", duration),
			slog.String("status_code", statusCode),
		}
		logger.Info("Connect Unary Call", attrs...)

		return resp, err
	}
}

func (l *ConnectLoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (l *ConnectLoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()

		clientIP := conn.RequestHeader().Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = conn.RequestHeader().Get("X-Real-IP")
		}
		userAgent := conn.RequestHeader().Get("User-Agent")

		err := next(ctx, conn)

		duration := time.Since(start)
		statusCode := "OK"
		if err != nil {
			if connectErr, ok := err.(*connect.Error); ok {
				statusCode = connectErr.Code().String()
			} else {
				statusCode = "unknown"
			}
		}

		attrs := []any{
			slog.String("service", l.service),
			slog.String("method", conn.Spec().Procedure),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.Time("start_time", start),
			slog.Duration("duration", duration),
			slog.String("status_code", statusCode),
		}
		logger.Info("Connect Stream Call", attrs...)

		return err
	}
}
