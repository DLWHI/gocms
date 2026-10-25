# Users

These routes are responsible for users management, such as creating, sucpending, password overrides and permissions grants.

<a id="cms-users-create"></a>

### `POST /cms/users`

Create a user.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:**

  ```json
  {
    "username": "string",
    "password": "password",
    "permissions": 65536,
    "email": "example@email.com",
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if user was created successfully. Returns empty body.
  - `400 Bad Request` — if request is invalid or username already exists. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `500 Internal Server Error` — if user creation fails. Returns a plain-text message.

<a id="cms-users-search"></a>

### `POST /cms/users/search`

Search user records. Search is performed on username and email

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:** Unknown fields are ignored. The request body may include:

```json
{
  "query": "string",
  "sort_by": "column:asc|desc",  // note that not all columns are sortable
  "page": 1,
  "count": 15,  // capped at 100
  "permissions": 4294967295,  // optional, bit field as number
  "suspended": boolean,  // optional
  "suspended_after": 2141453433,  // Unix timestamp
  "suspended_before": 2141453433,  // Unix timestamp
  "last_login_after": 2141453433,  // Unix timestamp
  "last_login_before": 2141453433,  // Unix timestamp
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
        "last_login": 2141453433,  // Unix timestamp
        "suspended_at": 2141453433,  // Unix timestamp
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `500 Internal Server Error` — if search service fails. Returns a plain-text message.

<a id="cms-users"></a>

### `GET /cms/users`

List all not suspended users.

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
      "last_login": 2141453433,  // Unix timestamp
      "suspended_at": 2141453433,  // Unix timestamp
    },
    {
      "username": "string",
      "email": "email",
      "permissions": 4294967295,  // optional, bit field as number
      "last_login": 2141453433,  // Unix timestamp
      "suspended_at": 2141453433,  // Unix timestamp
    }
  ]
  ```

  - `400 Bad Request` — if request is invalid. Returns a plain-text message.
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
    "last_login": 2141453433,  // Unix timestamp
    "suspended_at": 2141453433,  // Unix timestamp
  }
  ```

  - `400 Bad Request` — if request is invalid. Returns a plain-text message.
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
- **Request body:**

```json
  {"email": "sam@example.com"}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if operation succeeds. Returns empty body.
  - `400 Bad Request` — if request is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.
  - `500 Internal Server Error` — if repository failure occurs. Returns a plain-text message.

<a id="cms-users-username-reset"></a>

### `PATCH /cms/users/{username}/password`

Override user's password. Active sessions for that user are revoked after reset.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `username` — target user.
- **Request body:**

```json
  {"password": "password"}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if password change succeeds. Returns empty body.
  - `400 Bad Request` — if request is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.
  - `500 Internal Server Error` — if password/session service failure occurs. Returns a plain-text message.

<a id="cms-users-username-permissions"></a>

### `PATCH /cms/users/{username}/permissions`

Update a user's permissions.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`.
- **Path parameters:**
  - `username` — target user.
- **Request body:**

```json
  {"permissions": 65536}
```

- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if permissions change successfully. Returns empty body.
  - `400 Bad Request` — if request is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.
  - `500 Internal Server Error` — if users service fails to change permissions. Returns a plain-text message.

<a id="cms-users-username-delete"></a>

### `DELETE /cms/users/{username}`

Suspend a user and revoke that user's sessions. Note that this soft-deletes user from the database, menaing user data will still be available.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `username` — target user.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if user suspended successfully. Empty body.
  - `400 Bad Request` — if request is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — target user was not found. Returns a plain-text message.
  - `500 Internal Server Error` — if users service fails to suspend user. Returns a plain-text message.
