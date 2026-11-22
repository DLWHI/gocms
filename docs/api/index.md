# API documentation

This document covers general conventions applied on all routes. CMS provides 3 kinds of api endpoints to maintian and distribute content:

1. Admin API - routes responsible for CMS state management
2. Public API - interface that is designed to be used in end-user application
3. Storage API - routes responsible for media and file management of CMS.

All routes are conform with REST API. JSON Fields supplied to body that are not specified in route reference are ignored. Also fields may be omitted, in this case behaviour if specific per-route. If required field ends up omitted, route responds with `400 Bad Request`

## Common response codes

All routes may return these status codes if required conditions are met:

- `406 Not Acceptable` — if route requires `lang` path parameter and provided language is not supported by CMS or entity is not available in that language. Returns a plain-text message.
- `413 Content Too Large` — if request sends body larger than specified in **Request limit**. Some routes ignore body, so chunked data larger than limit gets ignored. Returns a plain-text message.
- `415 Unsupported Media Type` — if request sends `Content-Type` header other than specified in route reference. If reference does not specify `Content-Type`, any value is acceptable. Returns a plain-text message.
- `500 Internal Server Error` — if service failure occurs. Returns a plain-text message.

Use this jump directly to a grouped reference:

- [Admin API reference](admin.md)
- [Public API reference](public.md)
- [Storage API reference](storage.md)
