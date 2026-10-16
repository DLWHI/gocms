# API documentation

Use this page to jump directly to a route reference:

- [Admin API](admin-api.md) — CMS routes under `/cms`
- [Storage API](storage-api.md) — routes under `/storage`
- [Public API](public-api.md) — routes under `/public`

## CMS — users

| Method | Route | Reference |
| --- | --- | --- |
| POST | `/cms/login` | [Log in](admin-api.md#cms-login) |
| POST | `/cms/register` | [Register a user](admin-api.md#cms-register) |
| POST | `/cms/logout` | [Log out](admin-api.md#cms-logout) |
| GET | `/cms/profile` | [Get current profile](admin-api.md#cms-profile) |
| GET | `/cms/permissions` | [List permission names and values](admin-api.md#cms-permissions) |
| POST | `/cms/users/search` | [Search users](admin-api.md#cms-users-search) |
| GET | `/cms/users` | [List users](admin-api.md#cms-users) |
| GET | `/cms/users/{username}` | [Get a user](admin-api.md#cms-users-username) |
| PATCH | `/cms/users/{username}/email` | [Change email](admin-api.md#cms-users-username-email) |
| PATCH | `/cms/users/{username}/change` | [Change password](admin-api.md#cms-users-username-change) |
| PATCH | `/cms/users/{username}/reset` | [Reset a password](admin-api.md#cms-users-username-reset) |
| PATCH | `/cms/users/{username}/permissions` | [Change permissions](admin-api.md#cms-users-username-permissions) |
| DELETE | `/cms/users/{username}` | [Suspend a user](admin-api.md#cms-users-username-delete) |

## CMS — editors

| Method | Route | Reference |
| --- | --- | --- |
| POST | `/cms/editors/search` | [Search editors](admin-api.md#cms-editors-search) |
| GET | `/cms/editors` | [List editors](admin-api.md#cms-editors) |
| POST | `/cms/editors` | [Create an editor profile](admin-api.md#cms-editors-create) |
| GET | `/cms/editors/{username}` | [Get an editor profile](admin-api.md#cms-editors-username) |
| PATCH | `/cms/editors/{username}` | [Update an editor profile](admin-api.md#cms-editors-username-patch) |
| DELETE | `/cms/editors/{username}` | [Suspend an editor profile](admin-api.md#cms-editors-username-delete) |

## CMS — locales

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/locales` | [List locales](admin-api.md#cms-locales) |
| POST | `/cms/locales` | [Create a locale](admin-api.md#cms-locales-create) |
| DELETE | `/cms/locales/{code}` | [Delete a locale](admin-api.md#cms-locales-code-delete) |

## CMS — posts

| Method | Route | Reference |
| --- | --- | --- |
| POST | `/cms/posts` | [Create a post translation](admin-api.md#cms-posts-create) |
| POST | `/cms/posts/search` | [Search posts](admin-api.md#cms-posts-search) |
| GET | `/cms/posts/{id}` | [Get a post translation](admin-api.md#cms-posts-id) |
| PATCH | `/cms/posts/{id}` | [Update a post translation](admin-api.md#cms-posts-id-patch) |
| PATCH | `/cms/posts/{id}/views` | [Set post views](admin-api.md#cms-posts-id-views) |
| DELETE | `/cms/posts/{id}` | [Delete a post and its translations](admin-api.md#cms-posts-id-delete) |
| DELETE | `/cms/posts/{id}/{locale}` | [Delete one post translation](admin-api.md#cms-posts-id-locale-delete) |

## CMS — categories

| Method | Route | Reference |
| --- | --- | --- |
| POST | `/cms/categories/search` | [Search categories](admin-api.md#cms-categories-search) |
| GET | `/cms/categories` | [List categories](admin-api.md#cms-categories) |
| POST | `/cms/categories` | [Create a category](admin-api.md#cms-categories-create) |
| GET | `/cms/categories/{slug}` | [Get a category](admin-api.md#cms-categories-slug) |
| PATCH | `/cms/categories/{slug}` | [Update a category](admin-api.md#cms-categories-slug-patch) |
| DELETE | `/cms/categories/{slug}` | [Delete a category](admin-api.md#cms-categories-slug-delete) |

## CMS — advertisements

| Method | Route | Reference |
| --- | --- | --- |
| POST | `/cms/ads/search` | [Search advertisements](admin-api.md#cms-ads-search) |
| POST | `/cms/ads` | [Create an advertisement](admin-api.md#cms-ads-create) |
| GET | `/cms/ads/{id}` | [Get an advertisement](admin-api.md#cms-ads-id) |
| PATCH | `/cms/ads/{id}` | [Update an advertisement](admin-api.md#cms-ads-id-patch) |
| DELETE | `/cms/ads/{id}` | [Delete an advertisement](admin-api.md#cms-ads-id-delete) |
| GET | `/cms/ads/groups` | [List advertisement groups](admin-api.md#cms-ads-groups) |
| GET | `/cms/ads/groups/{group}` | [Get a group's advertisements](admin-api.md#cms-ads-groups-group) |
| POST | `/cms/groups/search` | [Get several advertisement groups](admin-api.md#cms-groups-search) |
| GET | `/cms/partners` | [Get partner advertisements](admin-api.md#cms-partners) |

## Storage

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/storage/tree/{path...}` | [List a storage directory](storage-api.md#storage-tree) |
| GET | `/storage/media/{path...}` | [Read a storage file](storage-api.md#storage-media-get) |
| PUT | `/storage/media/{path...}` | [Upload a storage file](storage-api.md#storage-media-put) |
| DELETE | `/storage/media/{path...}` | [Delete a storage path](storage-api.md#storage-media-delete) |
| PUT | `/storage/folders/{path...}` | [Create a storage folder](storage-api.md#storage-folders) |
| GET | `/storage/preview/{path...}` | [Read a storage preview path](storage-api.md#storage-preview) |

## Public API

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/public/locales` | [List public locales](public-api.md#public-locales) |
| POST | `/public/posts` | [Search published posts](public-api.md#public-posts-search) |
| GET | `/public/posts/{id}` | [Get a published post](public-api.md#public-posts-id) |
| GET | `/public/categories` | [List public categories](public-api.md#public-categories) |
| GET | `/public/editors` | [List public editor profiles](public-api.md#public-editors) |
| GET | `/public/editors/{username}` | [Get a public editor profile](public-api.md#public-editors-username) |
| GET | `/public/advertisements` | [Get public advertisement groups](public-api.md#public-advertisements) |
