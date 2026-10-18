# Admin API reference

This document covers CMS routes responsible for accessing and managing content on admin application. Return to the [route index](README.md).

## CMS — users

These routes are responsible for user management, as well as login, registration and sessions.

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

<a id="cms-register"></a>

### `POST /cms/register`

Create a user.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:**

  ```json
  {
    "username": "string",
    "passwd": "password",
    "email": "example@email.com",
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if user was created successfully. Returns empty body.
  - `400 Bad Request` — if body or username is invalid. Returns a plain-text message.
  - `401 Unauthorized` — is session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `500 Internal Server Error` — if user creation fails. Returns a plain-text message.

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

<a id="cms-users-search"></a>

### `POST /cms/users/search`

Search user records.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:**

```json
{
  "query": "string",
  "sort_by": "column:asc|desc",  // note that not all columns are sortable
  "page": 1,
  "count": 15,  // capped at 100
  "permissions": 4294967295,  // optional, bit field as number
  "suspended": boolean,  // optional
  "suspended_after": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
  "suspended_before": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
  "last_login_after": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
  "last_login_before": "2026-08-13T14:30:00Z",  // optional, must be RFC 3339 compliant
}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if search completed successfully.

  ```json
  {
    "results": [
      {
        "username": "string",
        "email": "email",
        "permissions": 4294967295,  // optional, bit field as number
        "last_login": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
        "suspended_at": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if body is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if search service fails. Returns a plain-text message.

<a id="cms-users"></a>

### `GET /cms/users`

List all users.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if user retreival succeeds

  ```json
  [
    {
      "username": "string",
      "email": "email",
      "permissions": 4294967295,  // optional, bit field as number
      "last_login": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
      "suspended_at": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
    },
    {
      "username": "string",
      "email": "email",
      "permissions": 4294967295,  // optional, bit field as number
      "last_login": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
      "suspended_at": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
    }
  ]
  ```

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if user retreival fails. Returns a plain-text message.

<a id="cms-users-username"></a>

### `GET /cms/users/{username}`

Get user by username.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `username` — target account username.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if user retreival succeeds

  ```json
  {
    "username": "string",
    "email": "email",
    "permissions": 4294967295,  // optional, bit field as number
    "last_login": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
    "suspended_at": "2026-08-13T14:30:00Z",  // RFC 3339 compliant
  }
  ```

  - `400 Bad Request` — if username is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if user not found. Returns a plain-text message.
  - `500 Internal Server Error` — if user retreival fails. Returns a plain-text message.

<a id="cms-users-username-email"></a>

### `PATCH /cms/users/{username}/email`

Change target user's email address.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `username` — target user's username.
- **Request body:** Plain JSON string

```json
  "sam@example.com"
```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if operation succeeds. Returns empty body.
  - `400 Bad Request` — if username or body is invalid. Returns a plain-text message.
  - `401 Unauthorized` — is session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.
  - `500 Internal Server Error` — if repository failure occurs. Returns a plain-text message.

<a id="cms-users-username-reset"></a>

### `PATCH /cms/users/{username}/reset`

Override user's password. Active sessions for that user are revoked after reset.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `username` — target user.
- **Request body:** Plain JSON string

```json
  "password"
```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if password change succeeds. Returns empty body.
  - `400 Bad Request` — if username or body is invalid. Returns a plain-text message.
  - `401 Unauthorized` — is session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.
  - `500 Internal Server Error` — is password/session service failure. Returns a plain-text message.

<a id="cms-users-username-permissions"></a>

### `PATCH /cms/users/{username}/permissions`

Update a user's permissions.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json` recommended.
- **Path parameters:**
  - `username` — account to update.
- **Request body:** Unsigned integer, for example `12`.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — User not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-users-username-delete"></a>

### `DELETE /cms/users/{username}`

Suspend a user and revoke that user's sessions. Requires user-delete
permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `username` — account to suspend.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Empty username. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — User not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

## CMS — editors

All editor routes require a Bearer session and `Accept-Language`. Request-body
limit: 2 MiB.

<a id="cms-editors-search"></a>

### `POST /cms/editors/search`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:** None.
- **Request body:** [Search request](#search-request) with editor fields.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Search envelope of editor documents.
  - `400 Bad Request` — Missing locale or invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 2 MiB. Returns a plain-text message.

<a id="cms-editors"></a>

### `GET /cms/editors`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — JSON array of active editor documents in the requested locale.
  - `400 Bad Request` — Missing locale. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-editors-create"></a>

### `POST /cms/editors`

Create an editor profile. Requires editor-create permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:** None.
- **Request body:** [Editor document](#editor-document). If `locale` is empty,
  the `Accept-Language` value is used.
- **Request limit:** 2 MiB.
- **Response:**
  - `201 Created` — Empty body.
  - `400 Bad Request` — Missing locale or invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `500 Internal Server Error` — Repository failure or request body exceeds 2 MiB. Returns a plain-text message.

<a id="cms-editors-username"></a>

### `GET /cms/editors/{username}`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `username` — account username.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — JSON editor document for the requested locale.
  - `400 Bad Request` — Missing locale. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Editor not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-editors-username-patch"></a>

### `PATCH /cms/editors/{username}`

Update an editor profile. The authenticated user must match the username in
the path.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `username` — account username.
- **Request body:** [Editor document](#editor-document); body `username` must
  equal the path parameter. Empty `locale` falls back to `Accept-Language`.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid JSON, missing locale, or path/body username mismatch. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or caller mismatch. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 2 MiB. Returns a plain-text message.

<a id="cms-editors-username-delete"></a>

### `DELETE /cms/editors/{username}`

Suspend an editor profile. Requires editor-delete permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `username` — account username.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Empty username. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

## CMS — locales

All locale routes require a Bearer session. Request-body limit: 4 KiB.

<a id="cms-locales"></a>

### `GET /cms/locales`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON array of `{"name":"English","code":"en"}` locale objects.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-locales-create"></a>

### `POST /cms/locales`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** `{"name":"English","code":"en"}`.
- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — Empty body.
  - `400 Bad Request` — Invalid JSON or locale. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-locales-code-delete"></a>

### `DELETE /cms/locales/{code}`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `code` — locale code.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Missing code. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `404 Not Found` — Locale not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

## CMS — posts

All post routes require a Bearer session and `Accept-Language`. Request-body
limit: 2 MiB.

<a id="cms-posts-create"></a>

### `POST /cms/posts`

Create a post translation. Requires post-create permission; the authenticated
user becomes the author.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:** None.
- **Request body:** [Post document](#post-document). `slug` must be non-empty
  and contain no whitespace. If `locale` is empty, `Accept-Language` is used.
- **Request limit:** 2 MiB.
- **Response:**
  - `201 Created` — JSON post document, including its assigned `post_id`.
  - `400 Bad Request` — Missing locale, invalid slug, or invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `500 Internal Server Error` — Repository failure or request body exceeds 2 MiB. Returns a plain-text message.

<a id="cms-posts-search"></a>

### `POST /cms/posts/search`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:** None.
- **Request body:** [Search request](#search-request) with post fields.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Search envelope of post documents.
  - `400 Bad Request` — Missing locale or invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 2 MiB. Returns a plain-text message.

<a id="cms-posts-id"></a>

### `GET /cms/posts/{id}`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `id` — unsigned decimal post ID.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — JSON post translation with categories.
  - `400 Bad Request` — Invalid ID or missing locale. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, insufficient permission, or unauthorized author. Returns a plain-text message.
  - `404 Not Found` — Post not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-posts-id-patch"></a>

### `PATCH /cms/posts/{id}`

Update a post translation. The author or a user with post-edit permission may
perform the update.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `id` — unsigned decimal post ID; it overrides `post_id` supplied in the body.
- **Request body:** [Post document](#post-document); `slug` must be non-empty
  and contain no whitespace. Empty `locale` uses `Accept-Language`.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid ID, invalid JSON, missing locale, or invalid slug. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, insufficient permission, or unauthorized author. Returns a plain-text message.
  - `404 Not Found` — Post not found. Returns a plain-text message.
  - `500 Internal Server Error` — Repository failure or request body exceeds 2 MiB. Returns a plain-text message.

<a id="cms-posts-id-views"></a>

### `PATCH /cms/posts/{id}/views`

Set the post's view count. Requires the post to exist; route middleware also
requires `Accept-Language`, although the handler does not use it.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `id` — unsigned decimal post ID.
- **Request body:** Unsigned integer, for example `25`.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid ID or invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `404 Not Found` — Post not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 2 MiB. Returns a plain-text message.

<a id="cms-posts-id-delete"></a>

### `DELETE /cms/posts/{id}`

Delete a post and all its translations. The author or a user with post-delete
permission may perform the operation.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `id` — unsigned decimal post ID.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid ID or missing locale header. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, insufficient permission, or unauthorized author. Returns a plain-text message.
  - `404 Not Found` — Post not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-posts-id-locale-delete"></a>

### `DELETE /cms/posts/{id}/{locale}`

Delete one translation. The author or a user with post-delete permission may
perform the operation.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `id` — unsigned decimal post ID.
  - `locale` — translation locale code.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid ID. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, insufficient permission, or unauthorized author. Returns a plain-text message.
  - `404 Not Found` — Post not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

## CMS — categories

All category routes require a Bearer session and `Accept-Language`. Request
body limit: 4 KiB.

<a id="cms-categories-search"></a>

### `POST /cms/categories/search`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:** None.
- **Request body:** [Search request](#search-request) with category fields.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Search envelope of category documents.
  - `400 Bad Request` — Missing locale or invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-categories"></a>

### `GET /cms/categories`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON array of categories in the requested locale.
  - `400 Bad Request` — Missing locale. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-categories-create"></a>

### `POST /cms/categories`

Create a category; requires category-create permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** [Category document](#category-document). `slug` must be
  non-empty and contain no whitespace. Empty `locale` uses `Accept-Language`.
- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — Empty body (the handler sets JSON content type but writes no JSON payload).
  - `400 Bad Request` — Invalid JSON, missing locale, or invalid slug. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-categories-slug"></a>

### `GET /cms/categories/{slug}`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `slug` — category slug.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON category document.
  - `400 Bad Request` — Missing locale or slug. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Category not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-categories-slug-patch"></a>

### `PATCH /cms/categories/{slug}`

Update a category; requires category-edit permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `slug` — route selector. The handler updates from the
  request document; it does not copy the path slug into the body.
- **Request body:** [Category document](#category-document). Supply the target
  `slug`; empty `locale` uses `Accept-Language`.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON category document.
  - `400 Bad Request` — Invalid JSON, missing locale, or invalid slug. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Category not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-categories-slug-delete"></a>

### `DELETE /cms/categories/{slug}`

Delete a category; requires category-delete permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Accept-Language` — required.
- **Path parameters:**
  - `slug` — category slug.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Missing slug or locale header. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Category not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

## CMS — advertisements

All advertisement routes require a Bearer session. Request-body limit: 4 KiB.

<a id="cms-ads-search"></a>

### `POST /cms/ads/search`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** [Search request](#search-request) with advertisement fields.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Search envelope of [advertisement documents](#advertisement-document).
  - `400 Bad Request` — Invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-ads-create"></a>

### `POST /cms/ads`

Create an advertisement; requires advertisement-create permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** [Advertisement document](#advertisement-document). `url`
  must be an absolute HTTP or HTTPS URL. `group` is lowercased and must not
  contain whitespace.
- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — Empty body.
  - `400 Bad Request` — Invalid JSON, URL, or group. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-ads-id"></a>

### `GET /cms/ads/{id}`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal advertisement ID.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON advertisement document.
  - `400 Bad Request` — Invalid ID. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Advertisement not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-ads-id-patch"></a>

### `PATCH /cms/ads/{id}`

Update an advertisement; requires advertisement-edit permission.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal advertisement ID.
- **Request body:** [Advertisement document](#advertisement-document). Its
  `id` is expected to match the path ID. `url` and `group` follow the same
  validation as create.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid ID or JSON, path/body ID mismatch, or invalid URL or group. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Advertisement not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.
- **Implementation note:** on a path/body ID mismatch, the handler writes
  `400` but currently does not return before calling the update service. Do
  not rely on this mismatch check to prevent a mutation.

<a id="cms-ads-id-delete"></a>

### `DELETE /cms/ads/{id}`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — unsigned decimal advertisement ID.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — Invalid ID. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Advertisement not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-ads-groups"></a>

### `GET /cms/ads/groups`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON array of group names.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-ads-groups-group"></a>

### `GET /cms/ads/groups/{group}`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `group` — group name.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON array of advertisement documents in the group.
  - `400 Bad Request` — Empty group. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Advertisement group not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.

<a id="cms-groups-search"></a>

### `POST /cms/groups/search`

Fetch advertisements for a list of groups. Note that the registered route is
`/cms/groups/search`, not `/cms/ads/groups/search`.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** JSON string array, for example `["partners","sidebar"]`.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — For non-empty input, JSON object mapping group names to
  advertisement arrays. For empty input, the handler returns an empty JSON
  array (`[]`) instead of an object.
  - `400 Bad Request` — Invalid JSON. Returns a plain-text message.
  - `401 Unauthorized` — Missing or invalid session, or insufficient permission. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure or request body exceeds 4 KiB. Returns a plain-text message.

<a id="cms-partners"></a>

### `GET /cms/partners`

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — JSON array of advertisements in the `partners` group.
  - `401 Unauthorized` — Missing or invalid session. Returns a plain-text message.
  - `404 Not Found` — Entity not found. Returns a plain-text message.
  - `500 Internal Server Error` — Service failure. Returns a plain-text message.
