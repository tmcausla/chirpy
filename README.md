# Chirpy

Chirpy is a RESTful API backend for a Twitter-inspired microblogging platform, built with Go and PostgreSQL. Users can register accounts, authenticate, publish short messages (chirps), and interact with other users through a JSON API.

The project also includes JWT-based authentication, refresh token management, and webhook integration for premium account upgrades.

## Motivation

I built Chirpy as part of the [Boot.dev](https://www.boot.dev/) backend development curriculum to develop my understanding of HTTP servers, REST API design, authentication, and database integration in Go.

The project provided an opportunity to work with several important backend concepts, including:

- Designing and implementing RESTful API endpoints.
- User authentication and authorization.
- Secure password storage and token management.
- Relational database design and SQL query generation.
- HTTP middleware and request handling.
- Integrating external services through webhooks.

## Quick Start

### Prerequisites

- [Go](https://go.dev/dl/) 1.27.2 or later.
- [PostgreSQL](https://www.postgresql.org/download/).
- [Goose](https://github.com/pressly/goose) for database migrations.

### Installation

Clone the repository:

```bash
git clone https://github.com/tmcausla/chirpy.git
cd chirpy
```

Download dependencies:

```bash
go mod download
```

### Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE chirpy;
```

### Environment Configuration

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
TOKEN_SECRET=your_secret_key
POLKA_KEY=your_polka_api_key
PLATFORM=dev
```

Replace the database credentials and secret values with your own.

- `DB_URL`: PostgreSQL connection string.
- `TOKEN_SECRET`: Secret used to sign and validate JWTs.
- `POLKA_KEY`: API key used to authenticate incoming Polka webhook requests.
- `PLATFORM`: Application environment. Set to `dev` to enable development-only administrative functionality.

### Run Migrations

Apply the database migrations using Goose:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

### Build and Run

Build the executable:

```bash
go build -o chirpy .
```

Start the server:

```bash
./chirpy
```

The server listens on port `8080`.

## Usage

Chirpy exposes a JSON REST API for managing users, authentication, and chirps.

### Health Check

Check whether the server is running:

```http
GET /api/healthz
```

### User Management

Register a new user:

```http
POST /api/users
Content-Type: application/json
```

```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

Update an existing user's email and password:

```http
PUT /api/users
Authorization: Bearer <access_token>
Content-Type: application/json
```

### Authentication

Log in to receive an access token and refresh token:

```http
POST /api/login
Content-Type: application/json
```

```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

Refresh an expired access token:

```http
POST /api/refresh
Authorization: Bearer <refresh_token>
```

Revoke a refresh token:

```http
POST /api/revoke
Authorization: Bearer <refresh_token>
```

Access tokens are valid for one hour, while refresh tokens remain valid for 60 days unless revoked.

### Chirps

Create a new chirp:

```http
POST /api/chirps
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
    "body": "Hello, Chirpy!"
}
```

Chirps are limited to 140 characters and undergo profanity filtering before being stored.

Retrieve all chirps:

```http
GET /api/chirps
```

Filter chirps by author:

```http
GET /api/chirps?author_id=<user_id>
```

Change the sorting order:

```http
GET /api/chirps?sort=desc
```

Retrieve an individual chirp:

```http
GET /api/chirps/{chirpID}
```

Delete a chirp:

```http
DELETE /api/chirps/{chirpID}
Authorization: Bearer <access_token>
```

Only the original author can delete a chirp.

### Chirpy Red

Chirpy supports premium account upgrades through an external Polka webhook integration.

The webhook endpoint accepts authenticated upgrade events:

```http
POST /api/polka/webhooks
Authorization: ApiKey <polka_api_key>
Content-Type: application/json
```

Successful upgrade events update the associated user's Chirpy Red membership status.

### Administration

View the number of requests made to the static file server:

```http
GET /admin/metrics
```

Reset the application:

```http
POST /admin/reset
```

**Note:** The reset endpoint is restricted to development mode (`PLATFORM=dev`) and deletes registered users and their associated data.

## Database

Chirpy uses PostgreSQL for persistent storage, with [Goose](https://github.com/pressly/goose) managing schema migrations and [sqlc](https://sqlc.dev/) generating type-safe Go code from SQL queries.

The database is organized around three primary entities:

- **Users:** Registered accounts containing user information and hashed passwords.
- **Chirps:** Short messages associated with the user who created them.
- **Refresh Tokens:** Authentication tokens associated with individual users, including their expiration and revocation status.

Users have a one-to-many relationship with both chirps and refresh tokens. Each user can create multiple chirps and maintain multiple refresh tokens.

## Contributing

This project was built as a learning exercise, but suggestions and improvements are welcome!

If you'd like to contribute:

1. Fork the repository.
2. Create a branch for your changes.
3. Make your changes and test them.
4. Submit a pull request describing your changes.
