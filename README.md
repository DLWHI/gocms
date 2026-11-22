# gocms

**gocms** is a headless, multi-language content management system written in Go. It provides authenticated APIs for managing content and a public API for serving published content to websites and applications.

## Features

- Manage posts and their translations, including publication status, authors, categories, media, and view counts.
- Configure supported languages using ISO 639-1 language codes.
- Manage categories, editor profiles, users and permissions, advertisements, and media files.
- Search CMS entities and published posts.
- Serve content through separate administration, storage, and public APIs.

## Requirements

- Go 1.26 or later
- MySQL
- Redis
- Meilisearch

The application connects to all three services at startup. Provide their connection settings through environment variables:

| Variable | Description |
| --- | --- |
| `DB_HOST` | MySQL host |
| `DB_PORT` | MySQL port |
| `DB_USER` | MySQL username |
| `DB_PASSWORD` | MySQL password |
| `DB_NAME` | MySQL database name |
| `REDIS_HOST` | Redis host |
| `REDIS_PORT` | Redis port |
| `MEILISEARCH_HOST` | Meilisearch host |
| `MEILISEARCH_PORT` | Meilisearch port |
| `MEILISEARCH_API` | Meilisearch API key; may be empty if the server does not require one |
| `IMAGE_STORAGE_PATH` | Media storage directory; defaults to `./storage` |

Redis is currently configured to use database 0 without a password.

## Quick start

Set the environment variables above in your shell, then run the following from the repository root.

1. Apply the database migrations and configure the search indexes:

   ```sh
   go run ./cmd/gocms --init
   ```

2. Create the first CMS user. The command prompts for a username, permissions, and password. Enter `all` for full permissions when appropriate:

   ```sh
   go run ./cmd/gocms --create-user
   ```

3. Start the server:

   ```sh
   go run ./cmd/gocms
   ```

The server listens on port `8080` by default. For example, to bind to a specific address and port:

```sh
go run ./cmd/gocms --address 127.0.0.1 --port 8080
```

The commands above are separate application runs; keep the same environment variables set in each shell session. You can also build the executable with `go build ./cmd/gocms`.

### Command-line options

| Option | Description |
| --- | --- |
| `--init`, `-i` | Run database migrations and configure search indexes |
| `--create-user` | Interactively create a CMS user |
| `--hard-reset` | Drop and recreate the database schema and search indexes |
| `--address`, `-a` | Server bind address; defaults to all interfaces |
| `--port`, `-p` | Server port; defaults to `8080` |
| `--storage`, `-s` | Media storage directory; overrides `IMAGE_STORAGE_PATH` |

**Warning:** `--hard-reset` deletes existing database content and search indexes. Do not use it on a database whose data you need to keep.

## API documentation

- [Admin API](docs/api/admin.md)
- [Public API](docs/api/public-api.md)
- [Storage API](docs/api/storage.md)

## Development

Run the Go test suite with:

```sh
go test ./...
```

The integration tests expect local services at MySQL `localhost:3306` (database and user both named `test`, with no password), Redis `localhost:6379`, and Meilisearch `http://localhost:7700` (no API key). The tests clear their test data, Redis database 0, and search index documents, so use disposable local services rather than production instances.

## License

This project is distributed under the [Apache License 2.0](LICENSE).
