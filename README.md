# Simple Bank API

A RESTful API service for a simple banking system built with Go, providing secure account management and money transfer functionality.

## 🚀 Features

- **User Management**: User registration and authentication
- **Account Management**: Create and manage bank accounts
- **Money Transfers**: Secure transfers between accounts with transaction support
- **JWT/PASETO Authentication**: Token-based authentication system
- **Database Transactions**: ACID compliance for financial operations
- **Input Validation**: Comprehensive request validation
- **Docker Support**: Containerized deployment

## 🛠 Tech Stack

- **Language**: Go 1.23
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: SQLC (SQL code generation)
- **Authentication**: JWT & PASETO tokens
- **Testing**: Testify, GoMock
- **Containerization**: Docker & Docker Compose
- **Database Migration**: golang-migrate
- **Linting**: golangci-lint

## 📋 Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL (if running locally)
- golang-migrate CLI tool

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**

   ```bash
   git clone https://github.com/mt5718214/simple-bank.git
   cd simplebank
   ```

2. **Start the services**

   ```bash
   docker-compose up -d
   ```

3. **The API will be available at**: `http://localhost:8080`

### Local Development

1. **Start PostgreSQL**

   ```bash
   make postgres
   ```

2. **Create database**

   ```bash
   make createdb
   ```

3. **Run database migrations**

   ```bash
   make migrateup
   ```

4. **Start the server**
   ```bash
   make server
   ```

## 📚 API Endpoints

### Authentication

- `POST /users` - Register a new user
- `POST /users/login` - User login

### Accounts (Authenticated)

- `POST /accounts` - Create a new account
- `GET /accounts/:id` - Get account by ID
- `GET /accounts` - List user's accounts

### Transfers (Authenticated)

- `POST /transfers` - Create a money transfer

## 🔧 Configuration

Copy `app.env` and modify the values as needed:

```env
DB_DRIVER="postgres"
DB_SOURCE="postgres://root:secret@localhost:5432/simple_bank?sslmode=disable"
SERVER_ADDRESS="0.0.0.0:8080"
TOKEN_SYMMETRIC_KEY=your-32-character-secret-key
ACCESS_TOKEN_DURATION=15m
```

## 🧪 Testing

Run all tests:

```bash
make test
```

## 🔨 Development Commands

```bash
# Database operations
make postgres          # Start PostgreSQL container
make createdb          # Create database
make dropdb            # Drop database
make migrateup         # Run all migrations
make migratedown       # Rollback all migrations

# Code generation
make sqlc              # Generate SQL code
make mock              # Generate mocks for testing

# Development
make server            # Start the server
make test              # Run tests
```

## 🐳 Docker

### Build and run with Docker Compose

```bash
docker-compose up --build
```

### Health Check

The application includes health checks to ensure PostgreSQL is ready before starting the API server.

## 📝 Code Quality

### Linting

```bash
# Install golangci-lint
brew install golangci-lint

# Run linter
golangci-lint run
```

### Pre-commit Hooks

Set up pre-commit hooks to run linting automatically:

```bash
git config core.hooksPath .githooks
```

## 🏗 Project Structure

```
.
├── api/                # HTTP handlers and middleware
├── db/
│   ├── migration/      # Database migration files
│   ├── mock/           # Generated mocks
│   ├── query/          # SQL queries
│   └── sqlc/           # Generated SQL code
├── token/              # JWT/PASETO token implementation
├── util/               # Utility functions and config
├── docs/               # Documentation
├── .github/workflows/  # CI/CD pipelines
└── docker-compose.yml  # Docker services configuration
```

## 🚀 Deployment

The project includes GitHub Actions workflows for:

- Automated testing
- Docker image building
- Push Docker image to ECR

### Commit Convention

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code formatting
- `refactor:` Code refactoring
- `test:` Adding tests
- `chore:` Maintenance tasks
