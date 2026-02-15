package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nradojcic/sentinel/internal/collector"
	pb "github.com/nradojcic/sentinel/proto"
)

var (
	serverAddr      string
	agentID         string
	reportIntervalS int
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Starts the Sentinel monitoring agent",
	Long:  `The agent collects system metrics (CPU, RAM) and streams them to the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		startAgent()
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.PersistentFlags().StringVar(&serverAddr, "server-addr", "localhost:50051", "The address of the Sentinel server")
	agentCmd.PersistentFlags().StringVar(&agentID, "agent-id", "default-agent", "A unique ID for this agent")
	agentCmd.PersistentFlags().IntVar(&reportIntervalS, "interval", 5, "Report interval in seconds")

	viper.BindPFlag("agent.server-addr", agentCmd.PersistentFlags().Lookup("server-addr"))
	viper.BindPFlag("agent.agent-id", agentCmd.PersistentFlags().Lookup("agent-id"))
	viper.BindPFlag("agent.interval", agentCmd.PersistentFlags().Lookup("interval"))
}

func startAgent() {
	addr := viper.GetString("agent.server-addr")
	id := viper.GetString("agent.agent-id")
	interval := time.Duration(viper.GetInt("agent.interval")) * time.Second

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to server", "address", addr, "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewMetricsServiceClient(conn)
	stream, err := client.ReportMetrics(context.Background())
	if err != nil {
		slog.Error("Failed to open metrics stream", "error", err)
		os.Exit(1)
	}

	slog.Info("Sentinel agent started", "agent_id", id, "server_addr", addr, "interval", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := collector.GetStats()
		if err != nil {
			slog.Error("Failed to collect system stats", "error", err)
			continue
		}

		payload := &pb.MetricPayload{
			AgentId:   id,
			Timestamp: timestamppb.Now(),
			CpuUsage:  stats.CPU,
			MemUsage:  stats.RAM,
		}

		if err := stream.Send(payload); err != nil {
			slog.Error("Failed to send metrics", "error", err)
			break
		}

		slog.Info("Sent metrics",
			"agent_id", id,
			"cpu", fmt.Sprintf("%.2f%%", stats.CPU),
			"ram", fmt.Sprintf("%.2f%%", stats.RAM),
		)
	}

	// Close the stream and receive summary from server
	summary, err := stream.CloseAndRecv()
	if err != nil {
		slog.Error("Error closing stream or receiving summary", "error", err)
	} else {
		slog.Info("Stream closed. Server summary",
			"reports_received", summary.ReportsReceived,
			"message", summary.Message,
		)
	}
}
