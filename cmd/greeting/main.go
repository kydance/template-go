package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kydance/ziwi/log"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/viper"

	"template-go/internal/greeting"
)

const (
	defaultConfigDir  = "etc"
	defaultConfigName = "ziwi"

	envPrefix = "ZIWI"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user home directory: %v", err))
	}

	viper.AddConfigPath(filepath.Join(home, defaultConfigDir)) // $HOME/defaultConfigDir
	viper.AddConfigPath(filepath.Join(".", defaultConfigDir))  // ./defaultConfigDir
	viper.AddConfigPath(filepath.Join("/Users/kyden/git-space/mcp-svr", defaultConfigDir))

	viper.SetConfigType("yaml")
	viper.SetConfigName(defaultConfigName)

	// Read matched environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file.
	// If a config file is specified, use it. Otherwise, search in defaultConfigDir.
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Failed to read viper config file: %v", err))
	}

	log.NewLogger(log.NewOptions().
		WithPrefix(viper.GetString("log.prefix")).
		WithDirectory("log.directory").
		WithLevel(viper.GetString("log.level")).
		WithTimeLayout(viper.GetString("log.time-layout")).
		WithFormat(viper.GetString("log.format")).
		WithDisableCaller(viper.GetBool("log.disable-caller")).
		WithDisableStacktrace(viper.GetBool("log.disable-stacktrace")).
		WithDisableSplitError(viper.GetBool("log.disable-split-error")).
		WithMaxSize(viper.GetInt("log.max-size")).
		WithMaxBackups(viper.GetInt("log.max-backups")).
		WithCompress(viper.GetBool("log.compress")))
}

func main() {
	// // Profiling
	// go func() {
	// 	svr := &http.Server{
	// 		Addr:              "127.0.0.1:9999",
	// 		ReadHeaderTimeout: 4 * time.Second,
	// 		ReadTimeout:       10 * time.Second,
	// 		WriteTimeout:      10 * time.Second,
	// 		IdleTimeout:       30 * time.Second,
	// 		Handler:           nil,
	// 	}
	// 	log.Fatalln(svr.ListenAndServe())
	// }()

	// // Prometheus Metrics
	// go func() {
	// 	http.Handle("/metrics", promhttp.Handler())

	// 	svr := &http.Server{
	// 		Addr:              "0.0.0.0:2112",
	// 		ReadHeaderTimeout: 4 * time.Second,
	// 		ReadTimeout:       10 * time.Second,
	// 		WriteTimeout:      10 * time.Second,
	// 		IdleTimeout:       30 * time.Second,
	// 	}

	// 	log.Fatalln(svr.ListenAndServe())
	// }()

	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or sse)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or sse)")
	flag.Parse()

	// Start the server
	mcpServer := greeting.NewMCPServer()
	switch transport {
	case "stdio":
		log.Info("Using stdio transport")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	case "streamable-http":
		log.Info("Using Streamable HTTP transport")
		steamableServer := server.NewStreamableHTTPServer(mcpServer,
			server.WithEndpointPath("/streamable"),
			server.WithHeartbeatInterval(5*time.Second),
			server.WithStateLess(false),
		)
		addr := fmt.Sprintf("%s:%d", "127.0.0.1", 5568)
		if err := steamableServer.Start(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	case "sse":
		log.Info("Using SSE transport")
		sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL("/sse"))
		log.Info("SSE server listening on :5568")
		addr := fmt.Sprintf("%s:%d", "127.0.0.1", 5568)
		if err := sseServer.Start(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	default:
		log.Errorf("invalid transport type: %s (must be 'stdio', 'sse' or 'streamable-http')", transport)
	}
}
