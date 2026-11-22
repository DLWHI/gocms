# Users

These routes are responsible for users management, such as creating, sucpending, password overrides and permissions grants.

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/users` | [List active users](#get-cmsusers) |
| GET | `/cms/users/{username}` | [Get a user](#get-cmsusersusername) |
| POST | `/cms/users` | [Create a user](#post-cmsusers) |
| POST | `/cms/users/search` | [Search users](#post-cmsuserssearch) |
| PATCH | `/cms/users/{username}/email` | [Change email](#patch-cmsusersusernameemail) |
| PATCH | `/cms/users/{username}/password` | [Reset a password](#patch-cmsusersusernamepassword) |
| PATCH | `/cms/users/{username}/permissions` | [Change permissions](#patch-cmsusersusernamepermissions) |
| DELETE | `/cms/users/{username}` | [Suspend a user](#delete-cmsusersusername) |

## Routes

### `GET /cms/users`

List all not suspended users. Request body is ignored if sent.

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

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `GET /cms/users/{username}`

Get user by username. Request body is ignored if sent.

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

  - `400 Bad Request` — if username is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if user not found. Returns a plain-text message.

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
    "permissions": 65536, // optional, defaults to editor role
    "email": "example@email.com", // optional
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if user was created successfully.

  ```json
  {
    "username": "string", // provided username in canonical form
    "location": "/cms/users/{username}"
  }
  ```

  - `400 Bad Request` — if request body or request parameters are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `409 Conflict` — if username already exists. Returns a plain-text message.

### `POST /cms/users/search`

Search user records. Search is performed on username and email

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:** Omitted/null fields are defaulted to zero-values unless stated otherwise.

```json
{
  "query": "string",
  "sort_by": "column:asc|desc",  // sortable columns: username, last_login
  "page": 1, // defaults to 1, clamped to 1
  "count": 15,  // defaults to 1, clamped to [1, 100]
  "permissions": 65535,  // bit field as number, ignored if omitted
  "suspended": boolean,  // ignored if omitted
  "suspended_after": 2141453433,  // Unix timestamp, ignored if omitted
  "suspended_before": 2141453433,  // Unix timestamp, ignored if omitted
  "last_login_after": 2141453433,  //  Unix timestamp, ignored if omitted
  "last_login_before": 2141453433,  // Unix timestamp, ignored if omitted
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
        "permissions": 65535,  // optional, bit field as number
        "last_login": 2141453433,  // Unix timestamp
        "suspended_at": 2141453433,  // Unix timestamp
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body or request parameters are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

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
  - `204 No Content` — if email change succeeds. Returns empty body.
  - `400 Bad Request` — if request body, parameters or username are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.

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
  - `400 Bad Request` — if request body, parameters or username are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.

### `PATCH /cms/users/{username}/permissions`

Update a user's permissions. Does not revoke user's sessions. Changes apply immedeatly

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
  - `400 Bad Request` — if request body, parameters or username are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target user not found. Returns a plain-text message.

### `DELETE /cms/users/{username}`

Suspend a user and revoke that user's sessions. Note that this soft-deletes user from the database, menaing user data will still be available. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `username` — target user.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if user suspended successfully. Returns empty body.
  - `400 Bad Request` — if username is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — target user was not found. Returns a plain-text message.
