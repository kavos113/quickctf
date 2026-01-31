package middleware

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct{}

func NewLoggingInterceptor() *LoggingInterceptor {
	return &LoggingInterceptor{}
}

func (l *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		clientIP := getClientIP(ctx)
		userAgent := getUserAgent(ctx)

		log.Printf("[gRPC] --> %s | client=%s | user-agent=%s",
			info.FullMethod,
			clientIP,
			userAgent,
		)

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

		log.Printf("[gRPC] <-- %s | status=%s | duration=%v | client=%s",
			info.FullMethod,
			statusCode.String(),
			duration,
			clientIP,
		)

		return resp, err
	}
}

func (l *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
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

		log.Printf("[gRPC Stream] --> %s | client=%s | user-agent=%s",
			info.FullMethod,
			clientIP,
			userAgent,
		)

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

		log.Printf("[gRPC Stream] <-- %s | status=%s | duration=%v | client=%s",
			info.FullMethod,
			statusCode.String(),
			duration,
			clientIP,
		)

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
