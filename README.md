# Expense Tracking API

A RESTful API service built in Go for tracking personal expenses and managing financial categories. This application helps users monitor and organize their spending by providing endpoints to manage expense records and categories.

## Docker Setup

### Building the Docker Image

```bash
docker build -t expense-tracker .
```

### Running with Docker Compose

```bash
# Build and start the container
docker-compose up --build

# Run in background
docker-compose up -d --build

# View logs
docker-compose logs -f app

# Stop containers
docker-compose down
```

### Environment Variables

The application supports the following environment variables:

- `DB_HOST` - Database host (default: localhost)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password 
- `DB_NAME` - Database name (default: postgres)
- `DB_PORT` - Database port (default: 5432)
- `PORT` - API server port (default: 8080)

You can override these by creating a `.env` file in the project root or passing them via docker-compose environment variables.

## Getting Started

### Prerequisites
- Git
- Docker and Docker Desktop (recommended)
- OR: Go 1.21+, PostgreSQL 15+

### Running with Docker (Recommended)

1. Clone the repository
```bash
git clone https://github.com/yanaatere/expense_tracking.git
cd expense_tracking
```

2. Set up environment variables (optional)
```bash
cat > .env << EOF
DB_HOST=db.btcqmtnjujfkasfkffwo.supabase.co
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=postgres
DB_PORT=5432
EOF
```

3. Build and run with Docker
```bash
docker-compose up --build
```

The API will be available at `http://localhost:8080`

### Running Locally

1. Clone the repository
```bash
git clone https://github.com/yanaatere/expense_tracking.git
```

2. Set up the database
- Install PostgreSQL
- Run the migration scripts in the migrations folder
```bash
go run cmd/migrate/main.go
```

3. Install dependencies
```bash
go mod download
```


4. Run the application
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

## Contributing

Feel free to submit issues, fork the repository and create pull requests for any improvements.

## License

This project is licensed under the MIT License.