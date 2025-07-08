# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an email account management system (邮箱账户管理系统) built with Go (Gin) backend, Vue.js frontend, and Chrome extension. The system manages email accounts, platform registrations, service subscriptions, and provides OAuth2 authentication with browser extension integration.

## Architecture

- **Backend**: Go with Gin framework, SQLite + GORM, JWT auth, OAuth2 (LinuxDo/Google/Microsoft)
- **Frontend**: Vue 3 + Composition API, Element Plus UI, Pinia state management, Vue Router
- **Extension**: Chrome manifest v3 with content scripts and background service worker
- **Database**: SQLite with encrypted sensitive data storage
- **Deployment**: Docker Compose with volume persistence

## Key Commands

### Development
```bash
# Backend
go run .                          # Run backend locally
go test ./...                     # Run tests
go build -o email_server main.go # Build binary

# Frontend
npm install                       # Install dependencies
npm run serve                     # Development server (proxies to backend)
npm run build                     # Production build
npm run lint                      # Code linting
npm run test:unit                 # Unit tests with Vitest

# Full stack
./src/build-and-deploy.sh        # Build both frontend and backend
```

### Production
```bash
docker-compose up -d             # Start services
docker-compose down              # Stop services
docker-compose build             # Build images
./deploy.sh                      # Full production deployment
./backup.sh                      # Database backup
```

## Code Structure

### Backend (`src/backend/`)
- `main.go` - HTTP server entry point, middleware setup
- `handlers/` - HTTP route handlers organized by feature
- `models/` - GORM data models (EmailAccount, Platform, ServiceSubscription, User)
- `middleware/` - Auth, CORS, static file serving
- `integrations/` - External APIs (Gmail, IMAP client)
- `utils/` - Encryption, JWT, password utilities

### Frontend (`src/frontend/`)
- `src/main.js` - Vue app entry with Element Plus
- `src/views/` - Route-level components (Login, Dashboard, EmailManagement, etc.)
- `src/components/` - Reusable components with Element Plus styling
- `src/stores/` - Pinia stores for state management
- `src/utils/` - API client, OAuth helpers, request interceptors

### Extension (`browser-extension/`)
- `manifest.json` - Chrome extension manifest v3
- `background.js` - Service worker for credential management
- `content.js` - Form detection and auto-fill functionality

## API Architecture

- Base URL: `/api/v1`
- Authentication: JWT Bearer tokens
- Public routes: `/auth/*`, `/oauth2/*`, `/health`
- Protected routes: `/email-accounts/*`, `/platforms/*`, `/subscriptions/*`, `/inbox/*`
- File upload: CSV import support (Bitwarden format)

## Key Features

1. **Email Account Management** - CRUD with IMAP integration and encrypted storage
2. **Platform Registration Tracking** - Website/service credential management
3. **Service Subscription Management** - Paid service tracking with cost management
4. **OAuth2 Authentication** - Multi-provider support (LinuxDo, Google, Microsoft)
5. **Browser Extension** - Auto-fill with form detection and credential dropdown

## Development Notes

- Frontend dev server proxies API calls to backend on port 8080
- Database migrations handled by GORM auto-migrate
- Sensitive data encrypted using custom encryption utility
- OAuth2 tokens stored securely with refresh capability
- Extension uses Chrome storage API for credential caching
- All user-facing text is in Chinese with English code comments

## Testing

- Backend: Go testing framework with `go test ./...`
- Frontend: Vitest with Vue Test Utils for unit tests
- Extension: Manual testing with Chrome developer tools