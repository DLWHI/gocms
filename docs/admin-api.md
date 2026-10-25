# Admin API reference

This document covers CMS routes responsible for accessing and managing content on admin application. Return to the [route index](README.md).

## CMS — auth

These routes are responsible for user authentication, sessions and profile management.

<a id="cms-login"></a>

### `POST /cms/login`

Log in and create a session.

- **Headers:**
  - `Content-Type` — `application/json`.
- **Path parameters:** None.
- **Request body:**

  ```json
  {
    "username":"sam",
    "passwd":"password"
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if login succeeds

  ```json
  {
    "session_id": "dr23r9tj9pw=12",
    "user": {
      "username": "string",
      "email": "email",
      "permissions": "permissions",
      "last_login": "last_login",
    }
  }
  ```

  - `400 Bad Request` — if request body is invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — if credentials were wrong. Returns a plain-text message.
  - `500 Internal Server Error` — if login or session creation failure. Returns a plain-text message.

<a id="cms-logout"></a>

### `POST /cms/logout`

Revoke the caller's current session.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if revoke was successful. Returns empty body.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if session revocation failure occured. In this case, there is no guarantee that session was deleted and it is recommended to retry request. Returns a plain-text message.

<a id="cms-profile"></a>

### `GET /cms/profile`

Return the authenticated user's account.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if session was validated successfully

  ```json
  {
    "username": "string",
    "email": "email",
    "permissions": 4294967295,  // bit field as number
    "last_login": "last_login",
  }
  ```

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if session retreival fails. Returns a plain-text message.

<a id="cms-permissions"></a>

### `GET /cms/permissions`

Return all cms permissions, namely entities and actions that can be done on those entities. The values of actions are bit masks used to determine whic action (`"create"`, `"delete"`, `"update"`) is requested. The values of entites are bit offset used to determine on which offset is entity action bit is located. I.e., this means overall permission is given by `(action & (p >> entity))`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` —

  ```json
  {
    "entities": {
      "create": 1,
      "update": 2,
      "delete": 4,
    },
    "actions": {
      "entity": 0,
      "post": 3
    }
  }
  ```

  - `401 Unauthorized` — fi session is invalid, missing or expired identity. Returns a plain-text message.

<a id="cms-password"></a>

### `PATCH /cms/password`

Change current user's password.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`.
- **Path parameters:** Not required
- **Request body:**

  ```json
  {
    "old_password": "old",
    "new_password": "new"
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if change succeeds. Returns empty body.
  - `400 Bad Request` — if username is empty or invalid JSON provided. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity, or if old credentials is invalid. Returns a plain-text message.
  - `500 Internal Server Error` — if change fails. Returns a plain-text message.

## CMS — editors

These routes are responsible for editor management.

<a id="cms-editors-search"></a>

### `POST /cms/editors/search`

Search editors. Search is performed on editor name, email and target username

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — Search language; required if the request body omits `locale`.
- **Path parameters:** None.
- **Request body:** Unknown fields are ignored. The request body may include:

```json
{
  "query": "string",
  "sort_by": "column:asc|desc",  // note that not all columns are sortable
  "page": 1,
  "count": 15,  // capped at 100
  "locale": "string",  // optional, request requires `Accept-Language` header if this field is missing
  "suspended": boolean,  // optional
  "suspended_after": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
  "suspended_before": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
}
```

- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if search succeeds

  ```json
  {
    "results": [
      {
        "username": "string",
        "email": "email@example.com",
        "bio": "Editor's bio in markdown",
        "name": "string",
        "locale": "strings",
        "suspended_at": 2141453433,  // Unix timestamp
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if search failure occurs. Returns a plain-text message.

<a id="cms-editors"></a>

### `GET /cms/editors`

Get all editors of specified locale.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if editor retreival succeeds

  ```json
  [
    {
      "username": "string",
      "email": "email@example.com",
      "bio": "Editor's bio in markdown",
      "name": "string",
      "locale": "strings",
      "suspended_at": 2141453433,  // Unix timestamp
    },
    {
      "username": "string",
      "email": "email@example.com",
      "bio": "Editor's bio in markdown",
      "name": "string",
      "locale": "strings",
      "suspended_at": 2141453433,  // Unix timestamp
    }
  ]
  ```

  - `400 Bad Request` — if target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if editor service fails. Returns a plain-text message.

<a id="cms-editors-create"></a>

### `POST /cms/editors`

Create an editor profile.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language, required if the request body omits `locale`.
- **Path parameters:** None.
- **Request body:**

```json
{
  "username": "string",
  "email": "email@example.com",
  "bio": "Editor's bio in markdown",
  "name": "string",
  "locale": "strings",  // optional, request requires `Accept-Language` header if this field is missing
}
```

- **Request limit:** 2 MiB.
- **Response:**
  - `201 Created` — if editor created successfully. Returns empty body.
  - `400 Bad Request` — if request body is malformed or a target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `500 Internal Server Error` — if editor service fails. Returns a plain-text message.

<a id="cms-editors-username"></a>

### `GET /cms/editors/{username}`

Get target username's editor profile.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language, required.
- **Path parameters:**
  - `username` — account username.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if user retreival succeeds

  ```json
  {
    "username": "string",
    "email": "<email@example.com>",
    "bio": "Editor's bio in markdown",
    "name": "string",
    "locale": "strings",  // optional, request requires `Accept-Language` header if this field is missing
    "suspended_at": 2141453433,  // Unix timestamp
  }
  ```

  - `400 Bad Request` — if username is invalid or target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — Target editor not found. Returns a plain-text message.
  - `500 Internal Server Error` — if editor service fails to retrieve data. Returns a plain-text message.

<a id="cms-editors-username-patch"></a>

### `PATCH /cms/editors/{username}`

Update an editor profile. Currently, only editor himself has rights to update info.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` —  target language, required if the request body omits `locale`.
- **Path parameters:**
  - `username` — target editor username.
- **Request body:**

```json
{
  "username": "string",
  "email": "email@example.com",
  "bio": "Editor's bio in markdown",
  "name": "string",
  "locale": "strings",  // optional, request requires `Accept-Language` header if this field is missing
}
```

- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if update succeeds. Returns empty body.
  - `400 Bad Request` — if request body is malformed or a target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target editor was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if editor service fails to commit update. Returns a plain-text message.

<a id="cms-editors-username-delete"></a>

### `DELETE /cms/editors/{username}`

Suspend an editor profile.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `username` — account username.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if delete succeeds. Returns empty body.
  - `400 Bad Request` — if username is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target editor was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if editor service fails delete. Returns a plain-text message.

## CMS — locales

These routes are responsible for content locale management.

<a id="cms-locales"></a>

### `GET /cms/locales`

Get all current locales supported by CMS.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` —

  ```json
  [
    {
    "name": "English",
    "code":"en"
    },
    {
      "name": "Deutsch",
      "code":"de"
    },
    {
      "name": "Русский",
      "code":"ru"
    }
  ]
  ```

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if locale service fails. Returns a plain-text message.

<a id="cms-locales-create"></a>

### `POST /cms/locales`

Create locale. Entites will be available to be edited in that locale

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:**

```json
{
  "name": "string",
  "code": "string", // ISO 639-1 compliant code
}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if creation succeeds. Returns empty body.
  - `400 Bad Request` — if request JSON is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `500 Internal Server Error` — if locales service fails to create locale. Returns a plain-text message.

<a id="cms-locales-code-delete"></a>

### `DELETE /cms/locales/{code}`

Create locale. Entites will be available to be edited in that locale

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `code` — target locale code.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if deletion succeeds. Returns empty body.
  - `400 Bad Request` — if code is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target locale was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if locales service fails to delete locale. Returns a plain-text message.

## CMS — posts

<a id="cms-posts-create"></a>

### `POST /cms/posts`

Create a post along with translation. The authenticated user becomes the author.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target post language, required if the request body omits `locale`.
- **Path parameters:** None.
- **Request body:** Unknown fields are ignored. Optional fields replaced by zero-values. The request body may include:

```json
  {
    "title": "string",
    "description": "string",
    "markdown": "Post content in markdown",
    "slug": "non-empty-string-without-space",
    "keywords": "non-empty-string-without-space",
    "locale": "string", // optional, request requires `Accept-Language` header if this field is missing
    "published_at": 2141453433,  // Unix timestamp, provide, if post needs to be published as created
    "preview": "https://example.com/preview.jpg",
    "media": "https://example.com/post_main_media.jpg",
    "media_source": "string",
    "media_type": "video|image",
    "categories": ["string", "array"],
  }
```

- **Request limit:** 2 MiB.
- **Response:**
  - `201 Created` — if creation was successful. Returns empty body
  - `400 Bad Request` — if request body is malformed or a target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `500 Internal Server Error` — if posts service fails to create. Returns a plain-text message.

<a id="cms-posts-search"></a>

### `POST /cms/posts/search`

Performs post search. The search is performed on post's title, description and keywords

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
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
        "locale": "en",
        "created_at": 2141453433,  // Unix timestamp
        "published_at": 2141453433,  // Unix timestamp
        "views": 1,
        "author": "username",
        "categories": ["category-slug"]
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body is malformed or a target locale is missing. Returns a plain-text message.
  - `500 Internal Server Error` — if search operation failed. Returns a plain-text message.

<a id="cms-posts-id"></a>

### `GET /cms/posts/{id}`

Get particular post by id

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language, required if the request body omits `locale`.
- **Path parameters:**
  - `id` — unsigned decimal post ID.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if post retreival succeeds

  ```json
  {
    "post_id": 1,
    "title": "string",
    "description": "string",
    "markdown": "Post content in Markdown format",
    "slug": "post-slug",
    "keywords": "string",
    "locale": "en",
    "created_at": 2141453433, // Unix timestamp
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

  - `400 Bad Request` — if id is invalid or locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post or post translation were not found. Returns a plain-text message.
  - `500 Internal Server Error` — if post service fails to retrieve post. Returns a plain-text message.

<a id="cms-posts-id-patch"></a>

### `PATCH /cms/posts/{id}`

Update a post translation. The author or a user with post-edit permission may
perform the update.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language, required.
- **Path parameters:**
  - `id` — unsigned decimal post ID
- **Request body:**

```json
{
  "title": "string",
  "description": "string",
  "markdown": "Post content in Markdown format",
  "slug": "post-slug",
  "keywords": "string",
  "locale": "en",
  "published_at": 2141453433, // Unix timestamp
  "preview": "https://example.com/preview.jpg",
  "media": "https://example.com/media.jpg",
  "media_source": "https://example.com/source.jpg",
  "media_type": "image",
  "author": "username",
  "categories": ["category-slug"]
}
```

- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if update succeeds. Returns empty body.
  - `400 Bad Request` — if id, request body or post data is invalid, or if target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if posts service fails to update. Returns a plain-text message.

<a id="cms-posts-id-views"></a>

### `PATCH /cms/posts/{id}/views`

Set the post's view count. Requires special priveleges

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal post ID.
- **Request body:** Plain unsigned integer

```json
25
```

- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if update succeeds. Returns empty body.
  - `400 Bad Request` — if id or request body. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if posts service fails to update. Returns a plain-text message.
  
<a id="cms-posts-id-delete"></a>

### `DELETE /cms/posts/{id}`

Delete a post and all its translations. The author or a user with post-delete
permission may perform the operation.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal post ID.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if delete succeeds. Returns empty body.
  - `400 Bad Request` — if id is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if posts service fails to delete. Returns a plain-text message.

<a id="cms-posts-id-locale-delete"></a>

### `DELETE /cms/posts/{id}/{locale}`

Delete one translation. The author or a user with post-delete permission may
perform the operation.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal post ID.
  - `locale` — target language.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if delete succeeds. Returns empty body.
  - `400 Bad Request` — if id or locale is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post translation was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if posts service fails to delete. Returns a plain-text message.

## CMS — categories

This section decribes routes that are responsible for category management.

<a id="cms-categories-search"></a>

### `POST /cms/categories/search`

Searches categories. Search is performed on category name and slug.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — Search language, required if the request body omits `locale`..
- **Path parameters:** None.
- **Request body:** Unknown parameters are ignored. Request may include:

```json
{
  "query": "string",
  "sort_by": "column:asc|desc",  // note that not all columns are sortable
  "page": 1,
  "count": 15,  // capped at 100
  "locale": "string",  // optional, request requires `Accept-Language` header if this field is missing
  "hidden": boolean,  // optional
  "navbar": boolean,  // optional
}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if seach succeeds succeeds

  ```json
  {
    "results": [
      {
        "slug": "string",
        "name": "string",
        "navbar": boolean,
        "hidden": boolean,
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if id, request body or post data is invalid, or if target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if search service fails. Returns a plain-text message.

<a id="cms-categories"></a>

### `GET /cms/categories`

Get all categories for specified locale

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**

  - `200 OK` — if retreival succeeds

  ```json
  [
    {
      "slug": "string",
      "name": "string",
      "locale": "string",
      "navbar": boolean,
      "hidden": boolean,
    },
    {
      "slug": "string",
      "name": "string",
      "locale": "string",
      "navbar": boolean,
      "hidden": boolean,
    }
  ]
  ```

  - `400 Bad Request` — if locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if category service fails to retrieve data. Returns a plain-text message.
  
<a id="cms-categories-create"></a>

### `POST /cms/categories`

Create a category.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language, required if body omits `locale`.
- **Path parameters:** None.
- **Request body:**

  ```json
  {
    "slug": "string",
    "name": "string",
    "locale": "string", // optional, request requires `Accept-Language` header if this field is missing
    "navbar": boolean,
    "hidden": boolean,
  }
  ```

- **Response:**
  - `201 Created` — if creation succeeds. Returns empty body.
  - `400 Bad Request` — if request or request data is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post translation was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if posts service fails to delete. Returns a plain-text message.

<a id="cms-categories-slug"></a>

### `GET /cms/categories/{slug}`

retrieve particular category.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language (required).
- **Path parameters:**
  - `slug` — category slug.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if retreival succeeds

  ```json
  {
    "slug": "string",
    "name": "string",
    "locale": "string",
    "navbar": boolean,
    "hidden": boolean,
  }
  ```

  - `400 Bad Request` — if slug is invalid or target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target category translation was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if category service fails to retrieve data. Returns a plain-text message.

<a id="cms-categories-slug-patch"></a>

### `PATCH /cms/categories/{slug}`

Update a category. Note that `slug` can be updated, despite it is being primary key.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — target language, required if body omits `locale`.
- **Path parameters:**
  - `slug` — target slug.
- **Request body:**

```json
  {
    "slug": "string",
    "name": "string",
    "locale": "string",  // optional, request requires `Accept-Language` header if this field is missing
    "navbar": boolean,
    "hidden": boolean,
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if update succeeds. Returns empty body.
  - `400 Bad Request` — if slug is invalid or target locale is missing. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target category translation was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if category service fails to update. Returns a plain-text message.

<a id="cms-categories-slug-delete"></a>

### `DELETE /cms/categories/{slug}`

Delete a category.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `slug` — target category slug.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if deletion succeeds. Returns empty body.
  - `400 Bad Request` — if slug is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target category was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if category service fails to delete. Returns a plain-text message.

## CMS — advertisements

This section decribes routes that are responsible for advertisement management.

<a id="cms-ads-search"></a>

### `POST /cms/ads/search`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** Unknown parameters are ignored. Request may include:

```json
{
  "query": "string",
  "sort_by": "column:asc|desc",  // note that not all columns are sortable
  "page": 1,
  "count": 15,  // capped at 100
  "expired": boolean,  // optional
  "expires_after": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
  "expires_before": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
  "groups": ["string", "array"],
}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if retreival succeeds

  ```json
  {
    "results": [
      {
        "id": 1,
        "name": "string",
        "description": "string",
        "url": "https://example.com/preview.jpg",
        "priority": 1,
        "group": "string",
        "expires_at": 2141453433,  // Unix timestamp, provide, if advertisement in temporary
        "banner": "https://example.com/post_main_media.jpg",
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if search failure occurs. Returns a plain-text message.

<a id="cms-ads-create"></a>

### `POST /cms/ads`

Create an advertisement.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:**

```json
{
  "name": "string",
  "description": "string",
  "url": "https://example.com/preview.jpg",
  "priority": 1,
  "group": "string",
  "expires_at": 2141453433,  // Unix timestamp, provide, if advertisement in temporary
  "banner": "https://example.com/post_main_media.jpg",
}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if advertisement created successfully. Returns empty body.
  - `400 Bad Request` — if request body is malformed or group is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `500 Internal Server Error` — if advertisement service fails. Returns a plain-text message.

<a id="cms-ads-id"></a>

### `GET /cms/ads/{id}`

retrieve specific advertisement.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal advertisement ID.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if advertisement retreival succeeds

  ```json
  {
    "id": 1,
    "name": "string",
    "description": "string",
    "url": "https://example.com/preview.jpg",
    "priority": 1,
    "group": "string",
    "expires_at": 2141453433,  // Unix timestamp, provide, if advertisement is temporary
    "banner": "https://example.com/post_main_media.jpg",
  }
  ```

  - `400 Bad Request` — if id is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if target advertisement was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if advertisement service fails to retrieve data. Returns a plain-text message.

<a id="cms-ads-id-patch"></a>

### `PATCH /cms/ads/{id}`

Update an advertisement

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal advertisement ID.
- **Request body:**

  ```json
  {
    "name": "string",
    "description": "string",
    "url": "https://example.com/preview.jpg",
    "priority": 1,
    "group": "string",
    "expires_at": 2141453433,  // Unix timestamp, provide, if advertisement is temporary
    "banner": "https://example.com/post_main_media.jpg",
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if update succeeds. Returns empty body.
  - `400 Bad Request` — if request body is malformed or id is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target advertisement was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if advertisement service fails to commit update. Returns a plain-text message.

<a id="cms-ads-id-delete"></a>

### `DELETE /cms/ads/{id}`

Delete advertisement

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal advertisement ID.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if delete succeeds. Returns empty body.
  - `400 Bad Request` — if id  is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target advertisement was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if advertisement service fails delete. Returns a plain-text message.

<a id="cms-ads-groups"></a>

### `GET /cms/ads/groups`

Get list of all `group` values that are currently in advertisements table.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if retreival succeeds.

  ```json
  [
    "string",
    "array",
    "of",
    "groups"
  ]
  ```

  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if advertisement service fails. Returns a plain-text message.
