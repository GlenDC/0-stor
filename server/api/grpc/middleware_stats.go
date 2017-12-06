package grpc

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/zero-os/0-stor/server/api"
	"github.com/zero-os/0-stor/server/stats"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// unaryStatsInterceptor creates an interceptor for a unary server method,
// which collects global read/write statistics.
// The method name defines whether it counts as read or write.
func unaryStatsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		go statsLogger(ctx, info.FullMethod)

		return handler(ctx, req)
	}
}

// streamStatsInterceptor creates an interceptor for a streaming server method,
// which collects global read/write statistics.
// The method name defines whether it counts as read or write.
func streamStatsInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		go statsLogger(stream.Context(), info.FullMethod)

		return handler(srv, stream)
	}
}

func statsLogger(ctx context.Context, grpcMethod string) {
	label, err := extractStringFromContext(ctx, api.GRPCMetaLabelKey)
	if err != nil {
		log.Errorf("Stat was not logged due to error: %v", err)
	}

	statsFunc, err := getStatsFunc(grpcMethod)
	if err != nil {
		log.Errorf("Stat was not logged due to error: %v", err)
	}
	statsFunc(label)
}

func getStatsFunc(grpcMethod string) (labelStatsFunc, error) {
	switch {
	case strings.HasPrefix(grpcMethod, objectPrefix):
		m := grpcMethod[objectPrefixLength:]
		f, ok := _StatsObjectMethodsMap[m]
		if !ok {
			return nil, errors.New("namespace object does not contain method " + m)
		}
		return f, nil

	case strings.HasPrefix(grpcMethod, namespacePrefix):
		m := grpcMethod[namespacePrefixLength:]
		f, ok := _StatsNamespaceMethodsMap[m]
		if !ok {
			return nil, errors.New("namespace namespace does not contain method " + m)
		}
		return f, nil

	default:
		return nil, fmt.Errorf("namespace `%s` not recognized by authentication middleware", grpcMethod)
	}
}

type labelStatsFunc func(label string)

var (
	_StatsObjectMethodsMap = map[string]labelStatsFunc{
		"Get":                 stats.IncrRead,
		"List":                stats.IncrRead,
		"Exists":              stats.IncrRead,
		"Check":               stats.IncrRead,
		"Create":              stats.IncrWrite,
		"SetReferenceList":    stats.IncrWrite,
		"AppendReferenceList": stats.IncrWrite,
		"RemoveReferenceList": stats.IncrWrite,
		"Delete":              stats.IncrWrite,
	}
	_StatsNamespaceMethodsMap = map[string]labelStatsFunc{
		"Get": stats.IncrRead,
	}
)