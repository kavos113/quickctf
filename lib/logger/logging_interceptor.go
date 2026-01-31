package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct{
	service string
}

func NewLoggingInterceptor(service string) *LoggingInterceptor {
	return &LoggingInterceptor{service: service}
}

func (l *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		clientIP := getClientIP(ctx)
		userAgent := getUserAgent(ctx)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		attrs := []any{
			slog.String("service", l.service),
			slog.String("method", info.FullMethod),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.Time("start_time", start),
			slog.Duration("duration", duration),
			slog.String("status_code", statusCode.String()),
			slog.Any("response", resp),
		}
		logger.Info("gRPC Unary Call", attrs...)

		return resp, err
	}
}

func (l *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		ctx := ss.Context()
		clientIP := getClientIP(ctx)
		userAgent := getUserAgent(ctx)

		err := handler(srv, ss)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		attrs := []any{
			slog.String("service", l.service),
			slog.String("method", info.FullMethod),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.Time("start_time", start),
			slog.Duration("duration", duration),
			slog.String("status_code", statusCode.String()),
		}
		logger.Info("gRPC Stream Call", attrs...)

		return err
	}
}

func getClientIP(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if xff := md.Get("x-forwarded-for"); len(xff) > 0 {
			return xff[0]
		}
		if xri := md.Get("x-real-ip"); len(xri) > 0 {
			return xri[0]
		}
	}

	return "unknown"
}

func getUserAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "unknown"
	}

	if ua := md.Get("user-agent"); len(ua) > 0 {
		return ua[0]
	}

	return "unknown"
}
