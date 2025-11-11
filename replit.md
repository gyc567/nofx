# NOFX AI Trading System - Replit Deployment

## Project Overview
NOFX is an AI-powered cryptocurrency trading system with support for multiple AI models (DeepSeek, Qwen) and exchanges (Binance, Hyperliquid, Aster DEX). This is a full-stack application with a Go backend and React/Vite frontend.

## Recent Changes (November 11, 2025)
- ✅ Migrated from Vercel to Replit
- ✅ Configured Vite to run on port 5000 (required for Replit webview)
- ✅ Built Go backend binary for faster startup
- ✅ Created unified workflow running both backend and frontend
- ✅ Frontend runs on port 5000, backend API on port 8080
- ✅ Admin mode enabled by default (no login required for testing)

## Architecture

### Backend (Go)
- **Port**: 8080
- **Binary**: `nofx-backend` (pre-compiled)
- **Database**: SQLite (`config.db`)
- **Features**:
  - REST API for trader management
  - WebSocket for real-time crypto market data
  - Support for Binance, Hyperliquid, Aster exchanges
  - AI integration with DeepSeek, Qwen, and custom APIs

### Frontend (React + Vite)
- **Port**: 5000 (exposed for Replit webview)
- **Location**: `web/` directory
- **Package Manager**: npm
- **Build Tool**: Vite 6.x

## Project Structure
```
.
├── nofx-backend          # Compiled Go binary
├── main.go               # Go backend entry point
├── config.json           # System configuration
├── config.db             # SQLite database
├── api/                  # API server code
├── trader/               # Trading logic
├── market/               # Market data handlers
├── web/                  # Frontend application
│   ├── src/             # React source code
│   ├── package.json     # Frontend dependencies
│   └── vite.config.ts   # Vite configuration
└── prompts/             # AI prompt templates
```

## Running the Application

The application uses a single workflow that starts both services:
- Backend: `./nofx-backend` (pre-compiled Go binary)
- Frontend: `cd web && npm run dev`

Both services start automatically via the `fullstack-app` workflow.

## Configuration

### System Config (`config.json`)
- **admin_mode**: `true` - Admin mode enabled (no login required)
- **beta_mode**: `false` - Beta access disabled
- **api_server_port**: `8080` - Backend API port
- **jwt_secret**: Pre-configured (should be changed in production)
- **leverage**: BTC/ETH 5x, Altcoins 5x
- **default_coins**: BTCUSDT, ETHUSDT, SOLUSDT, etc.

### Environment Variables
API keys for exchanges and AI models are stored in the SQLite database and configured via the web interface. For trading functionality, users need to configure:
- Exchange credentials (Binance API keys, Hyperliquid private key, etc.)
- AI model API keys (DeepSeek, Qwen, or custom API)

## Development Notes

### Rebuilding the Backend
If you modify the Go code, rebuild the backend:
```bash
go build -o nofx-backend main.go
```

### Installing Dependencies
- **Frontend**: `cd web && npm install`
- **Backend**: `go mod download`

### Database
The SQLite database (`config.db`) is created automatically on first run. It stores:
- User accounts
- Trader configurations
- AI model settings
- Exchange credentials
- Trading history

## Security Notes
- Admin mode is enabled for easy testing (bypasses authentication)
- JWT secret should be changed in production
- API keys and secrets are stored in the database (consider using Replit Secrets for production)
- Default admin user: `admin@localhost`

## Key Features
- Multi-agent AI trading with DeepSeek and Qwen
- Support for Binance Futures, Hyperliquid, and Aster DEX
- Real-time market data via WebSocket
- Community-driven competition system
- Self-hosted, full control over trading logic

## Troubleshooting

### Backend not starting
- Check if `nofx-backend` binary exists: `ls -lh nofx-backend`
- Rebuild if needed: `go build -o nofx-backend main.go`
- Check logs in the workflow console

### Frontend proxy errors
- Ensure backend is running on port 8080
- Check Vite config proxy settings in `web/vite.config.ts`

### Database issues
- Delete `config.db` to reset (will lose all data)
- Check file permissions: `ls -la config.db`

## Next Steps for Production
1. Change JWT secret in `config.json`
2. Configure real exchange API keys via the web interface
3. Add AI model API keys (DeepSeek or Qwen)
4. Set up proper authentication (disable admin mode)
5. Consider using Replit Secrets for sensitive credentials
6. Test with small amounts before live trading

## Links
- GitHub: https://github.com/your-repo/nofx
- Documentation: See `docs/` directory
- Quick Start: See `QUICK_START.md`
