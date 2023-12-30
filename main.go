package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

const defaultPort = 4317

type MetricsServiceDumpServer struct {
	w io.Writer
	metricspb.UnimplementedMetricsServiceServer
}

func (s *MetricsServiceDumpServer) Export(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
	if err := json.NewEncoder(s.w).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode the request: %v", err)
	}
	return &metricspb.ExportMetricsServiceResponse{}, nil
}

type TraceServiceDumpServer struct {
	w io.Writer
	tracepb.UnimplementedTraceServiceServer
}

func (s *TraceServiceDumpServer) Export(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
	if err := json.NewEncoder(s.w).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode the request: %v", err)
	}
	return &tracepb.ExportTraceServiceResponse{}, nil
}

type LogsServiceDumpServer struct {
	w io.Writer
	logspb.UnimplementedLogsServiceServer
}

func (s *LogsServiceDumpServer) Export(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
	if err := json.NewEncoder(s.w).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode the request: %v", err)
	}
	return &logspb.ExportLogsServiceResponse{}, nil
}

func main() {
	var port int
	flag.IntVar(&port, "port", defaultPort, "listen port")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen socket: %v", err)
	}
	server := grpc.NewServer()

	metricsDump := &MetricsServiceDumpServer{w: os.Stdout}
	traceDump := &TraceServiceDumpServer{w: os.Stdout}
	logsDump := &LogsServiceDumpServer{w: os.Stdout}

	metricspb.RegisterMetricsServiceServer(server, metricsDump)
	tracepb.RegisterTraceServiceServer(server, traceDump)
	logspb.RegisterLogsServiceServer(server, logsDump)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to start gRPC server: %v", err)
	}
}
