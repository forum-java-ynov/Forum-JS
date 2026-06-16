# Forum-JS

A web forum built in Go with SQLite — supports posts, comments, likes/dislikes, image uploads, and OAuth authentication (Google & GitHub).

---

## Features

- Registration / Login / Logout
- OAuth login via Google and GitHub
- Create and delete posts with optional image upload (JPEG, PNG, GIF, WebP)
- Comment on posts, edit and delete your own comments
- Like / Dislike posts and comments
- Filter posts by theme/category
- Session-based authentication with cookies
- Rate limiting on login and registration
- HTTP error handling (400, 401, 403, 404, 500)
- Unit tests for core modules

---

## Requirements

- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) (recommended)
- **Or** Go 1.21+ for local run

---

## Run with Docker

```bash
docker compose up --build
```

Then open [http://localhost:8082](http://localhost:8082)

To stop:

```bash
docker compose down
```

---

## Run Locally (without Docker)

```bash
go mod tidy
go run .
```

The server starts on [http://localhost:8082](http://localhost:8082).  
The SQLite database is created automatically at `database/database.db`.

---

## Run Tests

```bash
go test ./backend/...
```

---

## Project Structure

```
Forum-JS/
├── .github/                    # GitHub Actions workflows
├── backend/
│   ├── admin.go                # Admin handlers
│   ├── adminDB.go              # Admin DB queries
│   ├── auth.go                 # Login, register, OAuth (Google & GitHub)
│   ├── auth_test.go            # Auth unit tests
│   ├── comments.go             # Comment handlers
│   ├── Database.go             # DB init and all SQL queries
│   ├── Database_test.go        # DB unit tests
│   ├── filter.go               # Filter posts by theme
│   ├── likes.go                # Like/dislike logic
│   ├── posts.go                # Post handlers and image upload
│   ├── ratelimit.go            # Rate limiter middleware
│   ├── ratelimit_test.go       # Rate limiter unit tests
│   ├── Server.go               # HTTP server, routes, templates
│   ├── Server_test.go          # Server unit tests
│   ├── session.go              # Session management
│   └── validation.go           # Input validation
├── frontend/
│   ├── css/                    # Stylesheets
│   ├── html/                   # HTML templates
│   └── js/                     # JavaScript
├── database/                   # SQLite DB (auto-created at runtime)
├── uploads/                    # Uploaded images (auto-created at runtime)
├── .dockerignore
├── .env                        # Environment variables (not committed)
├── .gitignore
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

---

## API Routes

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/` | No | Home — list all posts |
| POST | `/db/register` | No | Register a new user |
| POST | `/db/login` | No | Login |
| GET | `/auth/logout` | No | Logout |
| GET | `/api/me` | Yes | Get current user info |
| GET | `/auth/google/login` | No | OAuth Google |
| GET | `/auth/github/login` | No | OAuth GitHub |
| POST | `/db/create_post` | Yes | Create a post |
| GET | `/db/posts` | No | Get all posts (JSON) |
| POST | `/db/delete_post` | Yes | Delete a post |
| POST | `/db/create_comment` | Yes | Add a comment |
| GET | `/db/comments?post_id=` | No | Get comments for a post |
| POST | `/db/edit_comment` | Yes | Edit a comment |
| POST | `/db/delete_comment` | Yes | Delete a comment |
| GET | `/db/toggle_like?id=` | Yes | Like a post |
| GET | `/db/toggle_dislike?id=` | Yes | Dislike a post |
| GET | `/db/toggle_comment_like?id=` | Yes | Like a comment |
| GET | `/db/toggle_comment_dislike?id=` | Yes | Dislike a comment |

---

## Notes

- Passwords are hashed with **bcrypt**
- Uploaded images are automatically resized to 800×800px max
- Rate limiting: 5 requests/minute on `/db/login` and `/db/register`
- The `database/` and `uploads/` folders are created automatically on first run
- OAuth credentials must be set in the `.env` file (see `.env.example` if provided)
