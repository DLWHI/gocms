# Public API reference

This document covers CMS routes that are responsible for accessing content on end-user application
Return to the [route index](README.md).

## Routes

<a id="public-locales"></a>

### `GET /public/locales`

List configured locales.

- **Headers:** Not required.
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if locales were retrieved successfully.

  ```json
  [
    {
      "name": "English",
      "code": "en"
    },
    {
      "name": "Deutsch",
      "code": "de"
    },
    {
      "name": "Русский",
      "code": "ru"
    }
  ]
  ```

  - `500 Internal Server Error` — if locale service could not list locales. Returns a plain-text message.

<a id="public-posts-search"></a>

### `POST /public/posts`

Search published posts.

- **Headers:**
  - `Accept-Language` — Search language; required if the request body omits `locale`.
  - `Content-Type` — `application/json`.
- **Path parameters:** None.
- **Request body:** Unknown fields are ignored. The request body may include:

  ```json
  {
    "query": "string",
    "sort_by": "column:asc|desc",  // note that not all columns are sortable
    "page": 1,
    "count": 15,  // capped at 100
    "locale": "string", // optional, request requires `Accept-Language` header if this field is missing
    "categories": ["string", "array"], // optional
    "exclude_categories": ["string", "array"], // optional
    "authors": ["string", "array"], // optional
    "published_after": "2026-08-13T14:30:00Z", // optional, must be RFC 3339 compliant
    "published_before": "2026-08-13T14:30:00Z", // optional, must be RFC 3339 compliant
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if search completed successfully.

  ```json
  {
    "results": [
      {
        "post_id": 1,
        "title": "string",
        "description": "string",
        "slug": "string",
        "locale": "en",
        "published_at": 2141453433,  // Unix timestamp
        "views": 1,
        "preview": "https://example.com/preview.jpg",
        "author": "username",
        "categories": ["category-slug"]
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body is malformed or a required locale is missing.
  - `500 Internal Server Error` — if search operation failed.

<a id="public-posts-id"></a>

### `GET /public/posts/{id}`

Get a published post by ID and slug.

- **Headers:**
  - `Accept-Language` — Target post language (required).
- **Path parameters:**
  - `id` — Target post ID; must be an unsigned integer.
- **Query parameters:**
  - `slug` — Target post slug; must match the post's slug.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if post was retrieved successfully.

  ```json
  {
    "post_id": 1,
    "title": "string",
    "description": "string",
    "markdown": "Post content in Markdown format",
    "slug": "post-slug",
    "keywords": "string",
    "locale": "en",
    "published_at": 2141453433, // Unix timestamp
    "views": 1,
    "preview": "https://example.com/preview.jpg",
    "media": "https://example.com/media.jpg",
    "media_source": "https://example.com/source.jpg",
    "media_type": "image",
    "author": "username",
    "categories": ["category-slug"]
  }
  ```

  - `400 Bad Request` — if locale or slug is missing, or the ID is invalid.
  - `404 Not Found` — if published post does not exist.
  - `500 Internal Server Error` — if retrieving the post, its categories, or its view count failed.

<a id="public-categories"></a>

### `GET /public/categories`

List all non-hidden categories at the requested locale.

- **Headers:**
  - `Accept-Language` — Target language (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if categories were retrieved successfully. If there are no categories, the response is an empty array.

  ```json
  [
    {
      "name": "Category",
      "slug": "category",
      "locale": "en",
      "hidden": false,
      "navbar": true
    },
    {
      "name": "Another category",
      "slug": "another-category",
      "locale": "en",
      "hidden": false,
      "navbar": false
    }
  ]
  ```

  - `400 Bad Request` — if locale is missing.
  - `500 Internal Server Error` — if retrieving categories failed.

<a id="public-editors"></a>

### `GET /public/editors`

List all active editor profiles for the requested locale.

- **Headers:**
  - `Accept-Language` — Target language (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if editors were retrieved successfully. If there are no editors, the response is an empty array.

  ```json
  [
    {
      "name": "Editor Name",
      "username": "username",
      "email": "editor@example.com",
      "locale": "en"
    },
    {
      "name": "Another Editor",
      "username": "another-user",
      "email": "another@example.com",
      "locale": "en"
    }
  ]
  ```

  - `400 Bad Request` — if locale is missing.
  - `500 Internal Server Error` — if retrieving editors failed.

<a id="public-editors-username"></a>

### `GET /public/editors/{username}`

Get an editor profile by account username for the requested locale.

- **Headers:**
  - `Accept-Language` — Target language (required).
- **Path parameters:**
  - `username` — editor username.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if editor profile was retrieved successfully.

  ```json
  {
    "name": "Editor Name",
    "bio": "Editor bio in markdown",
    "username": "string",
    "email": "string",
    "locale": "string"
  }
  ```

  - `400 Bad Request` — if locale is missing or the username is invalid.
  - `404 Not Found` — if editor does not exist.
  - `500 Internal Server Error` — if retrieving the editor failed.

<a id="public-advertisements"></a>

### `GET /public/advertisements`

Return unexpired advertisements for the requested groups.

- **Headers:** Not required.
- **Path parameters:** None.
- **Query parameters:**
  - `groups` — comma-separated group names. May be repeated. Whitespace around each group name is trimmed; empty names are ignored. Example:

  ```
  /public/advertisements?groups=partners,sidebar&groups=footer
  ```

- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if advertisements were retrieved successfully. A requested group with no advertisements maps to an empty array. If no group names are supplied, the response is an empty object.

  ```json
  {
    "group1": [
      {
        "name": "Advertisement",
        "url": "https://example.com",
        "priority": 1,
        "group": "group1",
        "banner": "https://example.com/banner.jpg"
      }
    ],
    "group2": [
      {
        "name": "Another advertisement",
        "url": "https://example.org",
        "priority": 2,
        "group": "group2",
        "banner": "https://example.org/banner.jpg"
      }
    ]
  }
  ```

  - `500 Internal Server Error` — if retrieving a requested advertisement group failed.
