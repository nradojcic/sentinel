package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nradojcic/sentinel/internal/dashboard"
	"github.com/nradojcic/sentinel/internal/store"
	pb "github.com/nradojcic/sentinel/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var serverPort string
var httpPort string

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the Sentinel monitoring server",
	Long:  `The server collects and stores metrics streamed from agents.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVar(&serverPort, "port", "50051", "The server port to listen on")
	viper.BindPFlag("server.port", serverCmd.PersistentFlags().Lookup("port"))

	serverCmd.PersistentFlags().StringVar(&httpPort, "http-port", "8080", "The HTTP server port for the dashboard")
	viper.BindPFlag("server.http-port", serverCmd.PersistentFlags().Lookup("http-port"))
}

// metricsServer implements the MetricsServiceServer interface
type metricsServer struct {
	pb.UnimplementedMetricsServiceServer
	store *store.MonitorStore
}

// ReportMetrics is the gRPC method where agents stream their metric payloads
func (s *metricsServer) ReportMetrics(stream pb.MetricsService_ReportMetricsServer) error {
	agentID := "unknown" // Default agent ID until first payload
	reportsReceived := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			slog.Info(fmt.Sprintf("Agent %s stream closed. Total reports: %d", agentID, reportsReceived))
			return stream.SendAndClose(&pb.SummaryResponse{
				ReportsReceived: int32(reportsReceived),
				Message:         fmt.Sprintf("Stream from agent %s closed. Total reports: %d", agentID, reportsReceived),
			})
		}
		if err != nil {
			slog.Error("Error receiving metrics from agent", "error", err, "agent_id", agentID)
			return err
		}

		agentID = req.AgentId

		s.store.Update(
			agentID,
			req.CpuUsage,
			req.MemUsage,
		)
		slog.Info("Received metrics",
			"agent_id", req.AgentId,
			"cpu", fmt.Sprintf("%.2f%%", req.CpuUsage),
			"ram", fmt.Sprintf("%.2f%%", req.MemUsage),
			"timestamp", req.Timestamp.AsTime(),
		)
		reportsReceived++
	}
}

// startServer initializes and runs the gRPC server and HTTP dashboard
func startServer() {
	// --- gRPC Server Setup ---
	grpcPort := viper.GetString("server.port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		slog.Error("Failed to listen for gRPC", "port", grpcPort, "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	myStore := store.NewMonitorStore()
	pb.RegisterMetricsServiceServer(grpcServer, &metricsServer{store: myStore})

	slog.Info("Sentinel gRPC server starting", "port", grpcPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Failed to serve gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	// --- HTTP Dashboard Setup ---
	httpPort := viper.GetString("server.http-port")
	httpServer, err := dashboard.NewServer(httpPort, "web/index.html", myStore)
	if err != nil {
		slog.Error("Failed to create HTTP server", "error", err)
		os.Exit(1)
	}

	// Start the HTTP server in a goroutine
	go func() {
		slog.Info("Sentinel HTTP dashboard starting", "port", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to serve HTTP dashboard", "error", err)
			os.Exit(1)
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // block until a signal is received

	slog.Info("Shutting down Sentinel servers...")

	// Create a context for HTTP server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server first (gracefully)
	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	// Then stop gRPC server (immediately)
	grpcServer.Stop()

	slog.Info("Sentinel servers stopped.")
}
