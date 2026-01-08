# Go Project Template

A Go web service template with rate limiting, Redis integration, linting, formatting, and git hooks pre-configured.

## Features

- **Gin Web Framework**: Fast HTTP router and middleware support
- **Custom Logging Middleware**: Integrated with kydenul/log for unified logging
- **Redis Integration**: Connection pooling with retry logic and health monitoring
- **Rate Limiting**: Multi-dimensional rate limiting (user ID, IP, endpoint)
- **Structured Logging**: File rotation, log levels, and configurable formats
- **Code Quality**: Pre-configured golangci-lint with 34+ linters
- **Git Hooks**: Automated formatting, linting, and commit message validation

## Quick Start

1. Clone this template
2. Update `go.mod` module name:

   ```bash
   # Replace with your module path
   go mod edit -module github.com/your-username/your-project
   ```

3. Install dependencies:

   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install mvdan.cc/gofumpt@latest
   go mod tidy
   ```

4. Configure in `config/config.yaml`:

    ```yaml
    KLOG:
      # Basic settings
      prefix: "ZIWI_"
      directory: "../logs"
      filename: "ziwi"
      level: "info"
      format: "console" # or "json"
      time_layout: "2006-01-02 15:04:05.000"

      disable_caller: false
      disable_stacktrace: false
      disable_split_error: false

      # File rotation
      max_size: 100 # MB
      max_backups: 5 
      compress: true

      # Console output control
      console_output: true

      # Sampling (reduces log volume in high-traffic scenarios)
      enable_sampling: true
      sample_initial: 100
      sample_thereafter: 1000

    Redis:
      host: 127.0.0.1
      port: 6379
      password: ""
      max_active: 100
      max_idle: 10
    ```

5. Build and run:

   ```bash
   make build
   ./bin/template-go
   ```

## Project Structure

```
.
├── main.go        # Application entry point
├── config/        # Configuration files (YAML)
├── internal/      # Private application code
│   ├── middleware/  # Custom middleware (Logger, Recovery)
│   └── models/    # Data models
├── pkg/           # Public library code
│   └── rate_limit/  # Rate limiting implementation
├── logs/          # Log output directory
├── docs/          # Documentation
└── script/        # Build/utility scripts
```

## Release

```bash
make release    # Generate changelog and bump version
```
