# Public API reference

This document covers CMS routes that are responsible for accessing content on end-user application.
Return to the [route index](README.md).

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/public/languages` | [List public locales](#get-publiclanguages) |
| GET | `/public/posts/{id}` | [Get a published post](#get-publicpostsid) |
| POST | `/public/posts` | [Search published posts](#post-publicposts) |
| GET | `/public/categories` | [List public categories](#get-publiccategories) |
| GET | `/public/editors` | [List public editor profiles](#get-publiceditors) |
| GET | `/public/editors/{slug}` | [Get a public editor profile](#get-publiceditorsslug) |
| GET | `/public/advertisements` | [Get public advertisement groups](#get-publicadvertisements) |
| GET | `/public/partners` | [Get partners groups](#get-publicpartners) |

## Routes

### `GET /public/languages`

List configured languages.

- **Headers:** Not required.
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if languages were retrieved successfully.

  ```json
  {
    "languages": [
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
  }
  ```

### `GET /public/posts/{id}`

Get a published post by ID.

- **Headers:**
  - `Accept-Language` — Target post language (required).
- **Path parameters:**
  - `id` — Target post ID; must be a positive integer.
- **Query parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if post was retrieved successfully.

  ```json
  {
    "post_id": 1,
    "title": "string",
    "description": "string",
    "content": "Rich text post content",
    "slug": "post-slug",
    "keywords": "string",
    "lang": "en",
    "published_at": 2141453433, // Unix timestamp
    "views": 1,
    "preview": "https://example.com/preview.jpg",
    "media": "https://example.com/media.jpg",
    "media_source": "https://example.com/source.jpg",
    "media_type": "image",
    "author": "slug",
    "categories": ["category-slug"]
  }
  ```

  - `400 Bad Request` — if language or slug is missing/invalid or the ID is invalid. Returns a plain-text message.
  - `404 Not Found` — if published post does not exist. Returns a plain-text message.
  - `406 Not Acceptable` — if requested `Accept-Language` is not supported by CMS or post does not exist it that translation. Returns a plain-text message.

### `POST /public/posts`

Search published posts. The search is performed on post's id, title, description, slug and keywords

- **Headers:**
  - `Accept-Language` — Search language (required)`.
  - `Content-Type` — `application/json`.
- **Path parameters:** None.
- **Request body:** Unknown fields are ignored. The request body may include:

  ```json
  {
    "query": "string",
    "sort_by": "column:asc|desc",  // sortable columns: post_id, title, author, published_at, views
    "page": 1,  // defaults to 1, clamped to 1
    "count": 15,  // defaults to 1, clamped to [1, 100]
    "categories": ["array", "of", "category", "slugs"], // might contain hidden categories, ignored if empty or omitted
    "authors": ["array", "of", "editor", "slugs"], // might contain soft-deleted editors, ignored if empty or omitted
    "published": boolean, // ignored if omitted
    "published_after": 2141453433,  // Unix timestamp, ignored if omitted
    "published_before": 2141453433,  // Unix timestamp, ignored if omitted
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
        "lang": "en",
        "published_at": 2141453433,  // Unix timestamp
        "views": 1,
        "preview": "https://example.com/preview.jpg",
        "author": "slug",
        "categories": ["category-slug"]
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body is malformed or a target language is invalid/missing. Returns a plain-text message.
  - `406 Not Acceptable` — if requested `Accept-Language` is not supported by CMS. Returns a plain-text message.

### `GET /public/categories`

List all non-hidden categories at the requested language.

- **Headers:**
  - `Accept-Language` — Target language (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if categories were retrieved successfully. If there are no categories, the response is an empty array.

  ```json
  {
    "categories": [
      {
        "name": "Category",
        "slug": "category",
        "lang": "en",
        "hidden": false,
        "navbar": true
      },
      {
        "name": "Another category",
        "slug": "another-category",
        "lang": "en",
        "hidden": false,
        "navbar": false
      }
    ]
  }
  ```

  - `400 Bad Request` — if language is missing or invalid. Returns a plain-text message.
  - `406 Not Acceptable` — if requested `Accept-Language` is not supported by CMS. Returns a plain-text message.

### `GET /public/editors`

List all active editor profiles for the requested language.

- **Headers:**
  - `Accept-Language` — Target language (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if editors were retrieved successfully. If there are no editors, the response is an empty array.

  ```json
  {
    "editors": [
      {
        "name": "Editor Name",
        "bio": "Editor bio in rich text",
        "slug": "string",
        "lang": "string"
      },
      {
        "name": "Editor Name",
        "bio": "Editor bio in rich text",
        "slug": "string",
        "lang": "string"
      }
    ]
  }
  ```

  - `400 Bad Request` — if language is missing or invalid. Returns a plain-text message.
  - `406 Not Acceptable` — if requested `Accept-Language` is not supported by CMS. Returns a plain-text message.

### `GET /public/editors/{slug}`

Get an editor profile by account slug for the requested language.

- **Headers:**
  - `Accept-Language` — Target language (required).
- **Path parameters:**
  - `slug` — editor slug.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if editor profile was retrieved successfully.

  ```json
  {
    "name": "Editor Name",
    "bio": "Editor bio in markdown",
    "slug": "string",
    "lang": "string"
  }
  ```

  - `400 Bad Request` — if language is missing/invalid or the slug is invalid. Returns a plain-text message.
  - `404 Not Found` — if editor does not exist. Returns a plain-text message.
  - `406 Not Acceptable` — if requested `Accept-Language` is not supported by CMS or editor is not available in that language. Returns a plain-text message.

### `GET /public/advertisements`

Return unexpired advertisements for the requested groups.

- **Headers:** Not required.
- **Path parameters:** None.
- **Query parameters:**
  - `groups` — comma-separated group names. May be repeated. Whitespace around each group name is trimmed; empty names are rejected. Example:

  ```url
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
        "id": 1,
        "name": "Advertisement",
        "url": "https://example.com",
        "priority": 1,
        "group": "group1",
        "banner": "https://example.com/banner.jpg"
      }
    ],
    "group2": [
      {
        "id": 1,
        "name": "Another advertisement",
        "url": "https://example.org",
        "priority": 2,
        "group": "group2",
        "banner": "https://example.org/banner.jpg"
      }
    ]
  }
  ```

  - `400 Bad Request` — if one of the requested groups is invalid. Returns a plain-text message.

### `GET /public/partners`

Returns advertisement that are in `partners` group

- **Headers:** Not required
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if retreival succeeds

  ```json
  {
    "partners": [
      {
        "id": 1,
        "name": "Advertisement",
        "url": "https://example.com",
        "priority": 1,
        "group": "partners",
        "banner": "https://example.com/banner.jpg"
      },
      {
        "id": 1,
        "name": "Advertisement",
        "url": "https://example.com",
        "priority": 1,
        "group": "partners",
        "banner": "https://example.com/banner.jpg"
      }
    ]
  }
  ```
