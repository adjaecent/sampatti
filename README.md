# Sampatti - Investment Portfolio Dashboard

A unified dashboard for tracking your investments across multiple platforms including Kuvera and Stockal. Sampatti securely connects to your investment accounts and provides a consolidated view of your portfolio performance.

## 🌟 Features

- **Multi-Platform Support**: Connect Kuvera and Stockal accounts
- **Secure Authentication**: Google OAuth login with invite-only access
- **Automated Data Fetching**: Scheduled updates every 24 hours
- **Consolidated Dashboard**: Unified view of all your investments
- **Privacy-First**: Encrypted credential storage with bcrypt
- **Real-time Portfolio Tracking**: Current values, gains/losses, and performance metrics
- **Clean UI**: Simple, responsive interface using DaisyUI

## 🛡️ Security & Privacy

- **Encrypted Storage**: All passwords are hashed using bcrypt before storage
- **Read-Only Access**: Only fetches data, never modifies or trades
- **Google OAuth**: Secure authentication without storing additional passwords
- **Invite-Only**: Access control for family/group plans
- **Local Database**: SQLite database stored on your server

## 🚀 Quick Start

### Prerequisites

- Go 1.25.0+
- Google OAuth2 credentials
- Valid Kuvera and/or Stockal accounts

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/adjaecent/sampatti.git
   cd sampatti
   ```

2. **Set up environment variables**
   ```bash
   export GOOGLE_CLIENT_ID="your-google-client-id"
   export GOOGLE_CLIENT_SECRET="your-google-client-secret"
   export SESSION_SECRET="your-secure-session-secret"
   export BASE_URL="http://localhost:8080"  # or your domain
   export DATABASE_PATH="./sampatti.db"
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the application**
   ```bash
   go run main.go
   ```

5. **Access the dashboard**
   Open http://localhost:8080 in your browser

## 🏗️ Architecture

### Database Schema

- **plans**: User plans and ownership
- **accounts**: Connected investment platform accounts (encrypted)
- **plan_members**: Invite-only access control
- **kuvera_fetches**: Historical Kuvera data (append-only)
- **stockal_fetches**: Historical Stockal data (append-only)

### Core Components

- **Authentication**: Google OAuth2 with session management
- **Data Fetching**: Automated background jobs using the unofficial APIs
- **Dashboard**: Server-rendered HTML with DaisyUI styling
- **Scheduler**: 24-hour intervals for data synchronization

## 📊 Supported Platforms

### Kuvera
- Portfolio summary and performance
- Detailed fund holdings
- Transaction history and SIP details
- Current gold prices

### Stockal
- Account summary and cash balances
- Portfolio details and holdings
- US stock and ETF positions

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_CLIENT_ID` | Google OAuth2 client ID | Required |
| `GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret | Required |
| `SESSION_SECRET` | Session encryption key | `your-session-secret-change-me` |
| `BASE_URL` | Application base URL | `http://localhost:8080` |
| `DATABASE_PATH` | SQLite database file path | `./sampatti.db` |
| `PORT` | Server port | `8080` |

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth2 credentials
5. Add your domain to authorized redirect URIs: `{BASE_URL}/auth/google/callback`

## 🚦 Usage

### Adding Accounts

1. Sign in with Google OAuth
2. Navigate to "Accounts" section
3. Click "Add Account"
4. Select platform (Kuvera or Stockal)
5. Enter your platform credentials
6. Credentials are encrypted and stored securely

### Dashboard Overview

- **Total Portfolio Value**: Combined value across all platforms
- **Connected Accounts**: Number and types of linked accounts
- **Recent Activity**: Latest data fetches and updates
- **Platform Breakdown**: Separate views for Kuvera and Stockal data

### Data Fetching

- **Automatic**: Every 24 hours via background scheduler
- **Manual**: Click "Fetch Data" button for immediate update
- **Historical**: All fetches are stored for trend analysis

## 🛠️ Development

### Project Structure

```
sampatti/
├── main.go                 # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── database/          # Database setup and migrations
│   ├── handlers/          # HTTP request handlers
│   ├── middleware/        # Authentication middleware
│   ├── models/           # Data models
│   ├── scheduler/        # Background job scheduler
│   └── services/         # Business logic services
├── templates/            # HTML templates
└── static/              # Static assets
```

### Running in Development

```bash
# Install dependencies
go mod tidy

# Run with live reload (if you have air installed)
air

# Or run directly
go run main.go
```

### Building for Production

```bash
# Build binary
go build -o sampatti

# Run binary
./sampatti
```

## 📝 API Integration

This application integrates with:

- [Unofficial Kuvera API](https://github.com/adjaecent/unofficial-kuvera-api)
- [Unofficial Stockal API](https://github.com/adjaecent/unofficial-stockal-api)

## ⚠️ Disclaimer

This is an unofficial application not affiliated with Kuvera or Stockal. Use at your own risk. Always verify data accuracy before making investment decisions.

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

Contributions welcome! Please read the contributing guidelines and submit pull requests.

## 🐛 Issues

Report issues on the [GitHub Issues](https://github.com/adjaecent/sampatti/issues) page.