# Expense Tracking API

A RESTful API service built in Go for tracking personal expenses and managing financial categories. This application helps users monitor and organize their spending by providing endpoints to manage expense records and categories.

## Features

- Category Management
  - Create, read, update, and delete expense categories
  - Categorize expenses for better organization
  - Track category-specific spending

- RESTful API Endpoints
  - Categories API
    - GET /api/categories - List all categories
    - GET /api/categories/{id} - Get category details
    - POST /api/categories - Create new category
    - PUT /api/categories/{id} - Update category
    - DELETE /api/categories/{id} - Delete category

## Technology Stack

- Go (Golang)
- PostgreSQL Database
- Gorilla Mux Router
- RESTful API Architecture

## Project Structure

```
expense_tracking/
├── controllers/
│   └── category_controller.go
├── models/
│   └── category.go
├── migrations/
│   └── 002_create_categories_table.sql
└── main.go
```

## Getting Started

1. Clone the repository
```bash
git clone https://github.com/yanaatere/expense_tracking.git
```

2. Set up the database
- Install PostgreSQL
- Create a new database
- Run the migration scripts in the migrations folder

3. Run the application
```bash
go run main.go
```

## API Documentation

### Categories

#### List Categories
```
GET /api/categories
```

#### Get Category
```
GET /api/categories/{id}
```

#### Create Category
```
POST /api/categories
Content-Type: application/json

{
    "name": "string",
    "description": "string"
}
```

#### Update Category
```
PUT /api/categories/{id}
Content-Type: application/json

{
    "name": "string",
    "description": "string"
}
```

#### Delete Category
```
DELETE /api/categories/{id}
```

## Contributing

Feel free to submit issues, fork the repository and create pull requests for any improvements.

## License

This project is licensed under the MIT License.