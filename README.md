
# URL Shortener

A simple and efficient URL shortening service built with Go and Redis.

## Features

- 🔗 Shorten long URLs to compact, shareable links
- 📊 Track visit metrics for shortened URLs
- ⚡ Fast Redis-based storage
- 🎨 Clean, modern UI with Tailwind CSS
- 🐳 Docker support for easy deployment
- 🔄 CI/CD pipelines with GitHub Actions

## Tech Stack

- **Backend**: Go 1.25
- **Storage**: Redis
- **Frontend**: HTML templates with Tailwind CSS
- **Deployment**: Docker & Docker Compose

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Redis
- Docker (optional)

### Local Development

1. Clone the repository
2. Copy the environment file:
   ```bash
   cp .env.example .env
   ```
3. Update `.env` with your configuration
4. Run the application:
   ```bash
   go run main.go
   ```

### Docker Deployment

For development:
```bash
docker-compose up
```