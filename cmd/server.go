package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nradojcic/sentinel/internal/store"
	pb "github.com/nradojcic/sentinel/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var serverPort string

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

// startServer initializes and runs the gRPC server
func startServer() {
	port := viper.GetString("server.port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		slog.Error("Failed to listen", "port", port, "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	myStore := store.NewMonitorStore()
	pb.RegisterMetricsServiceServer(grpcServer, &metricsServer{store: myStore})

	slog.Info("Sentinel server starting", "port", port)

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Failed to serve gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // block until a signal is received
	slog.Info("Shutting down Sentinel server gracefully...")
	grpcServer.GracefulStop()
	slog.Info("Sentinel server stopped.")
}
