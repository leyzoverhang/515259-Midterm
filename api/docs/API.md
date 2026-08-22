# Recipe API contract

This hand-authored contract covers recipe, user, and auth endpoints (all fully implemented) plus reference-data endpoints (contract only — not yet implemented, see [Reference data](#reference-data)). Generated Swagger files intentionally describe only implemented routes.

## Conventions

- Base path: `/api/v1`.
- All endpoints require `Authorization: Bearer <access-token>`.
- Request bodies use `Content-Type: application/json`; responses are JSON.
- Fields use camelCase; timestamps are RFC 3339 strings.
- Recipe, ingredient, and instruction IDs are JSON integers, never strings.
- Errors are `{ "message": "..." }`. The response type reserves a `code` field, but no endpoint populates it yet — treat the HTTP status as authoritative.

Unless otherwise specified: `400` is invalid input, `401` is a missing/invalid token, and `500` is an unexpected failure.

## Representations

Reference data:

```json
{ "id": "easy", "name": "Easy" }
```

Recipe:

```json
{
  "id": 42,
  "name": "Tom yum soup",
  "description": "A bright, spicy Thai soup.",
  "imageUrl": "https://images.example.com/tom-yum.jpg",
  "difficulty": { "id": "medium", "name": "Medium" },
  "duration": { "id": "30m", "name": "10 - 30 mins" },
  "ingredients": [{ "id": 101, "description": "2 cups stock" }],
  "instructions": [{ "id": 201, "description": "Bring the stock to a simmer." }],
  "creator": { "id": "3f0c1a7e-2b19-4c5e-9f3a-000000000000", "name": "Somchai" },
  "isFavorite": false,
  "rating": { "average": 4.9, "total": 128 },
  "createdAt": "2026-08-20T10:00:00Z",
  "updatedAt": "2026-08-20T10:00:00Z"
}
```

`imageUrl` is `null` when absent. `creator.name` is a string. Only active ingredients and instructions are returned; no `deletedAt` field is exposed. `isFavorite` is `true` when the authenticated caller has favorited this recipe (a matching row exists in `user_favorites`), `false` otherwise — see [Favorites](#favorites). `rating.average` is `recipes.average_rating` rounded to 1 decimal place; `rating.total` is the count of ratings for this recipe. Both are `0` when the recipe has no ratings yet — see [Ratings](#ratings).

Create and replace use a complete write body. `name`, `description`, `difficultyId`, `durationId`, `ingredients`, and `instructions` are required; `imageUrl` is optional. Ingredient/instruction arrays may be empty, and each item has a required non-empty `description`.

```json
{
  "name": "Tom yum soup",
  "description": "A bright, spicy Thai soup.",
  "imageUrl": "https://images.example.com/tom-yum.jpg",
  "difficultyId": "medium",
  "durationId": "30m",
  "ingredients": [{ "description": "2 cups stock" }],
  "instructions": [{ "description": "Bring the stock to a simmer." }]
}
```

`difficultyId` and `durationId` must reference active master-data records; otherwise the response is `400 invalid_request` (not `404`, despite the name). The `ingredients`/`instructions` keys must be present in the body — an explicit `[]` is fine, but omitting the key or sending `null` fails validation. When present, `imageUrl` must be a valid URL.

User:

```json
{
  "id": "3f0c1a7e-2b19-4c5e-9f3a-000000000000",
  "name": "Somchai",
  "email": "somchai@example.com",
  "imageUrl": "https://images.example.com/avatar.jpg",
  "bio": "Home cook who loves spicy food."
}
```

`name` and `email` are sourced from Keycloak (refreshed on every login from the ID token's claims) and are read-only through this API. `id` is an internally generated UUID, not the Keycloak subject — the Keycloak subject (`sub`) is stored server-side as a separate, unexposed field used only to look up the local user record. `imageUrl` and `bio` are `null` when absent.

## Reference data

**Not implemented yet.** No route, handler, service, or repository exists for `/difficulties` or `/durations` — calling them currently 404s at the router level (no matching route), not via the handler's own 404 logic. `difficulties` and `durations` tables exist (with the seed rows below) and are already read internally by the recipe package (`Repository.HasActiveReferences`, `Repository.DifficultyExists`) to validate `difficultyId`/`durationId` on recipe writes and the `difficulty` list filter — only the standalone list endpoints described below are missing. The contract below is the target shape for when they're built.

### `GET /difficulties`

Returns active difficulties as `{ "total": 3, "results": [<reference-data>, ...] }`.

| Status   | Meaning            |
| -------- | ------------------ |
| 200      | Active master data |
| 401, 500 | Error body         |

Fixed values: `easy` (Easy), `medium` (Medium), `hard` (Hard).

### `GET /durations`

Returns active durations as `{ "total": 4, "results": [<reference-data>, ...] }`.

| Status   | Meaning            |
| -------- | ------------------ |
| 200      | Active master data |
| 401, 500 | Error body         |

Fixed values: `10m` (5 - 10 mins), `30m` (10 - 30 mins), `60m` (~1 Hour), and `long` (More than 1 hour).

## Recipes

### `POST /recipes`

Send the complete write body. The authenticated user becomes the creator. On success, the response body is **`{ "id": <int> }` only** — not the full recipe.

| Status | Meaning                                                                          |
| ------ | -------------------------------------------------------------------------------- |
| 201    | `{ "id": <int> }`                                                                |
| 400    | Missing/malformed body, or referenced `difficultyId`/`durationId` does not exist |
| 401    | Error body                                                                       |
| 500    | Error body                                                                       |

An invalid/malformed body and an unresolvable `difficultyId`/`durationId` both currently return `400 { "message": "invalid request" }` — there is no distinct `404` case for this endpoint despite reference-data lookups happening server-side.

### `GET /recipes/{recipeId}`

`recipeId` is a required integer.

| Status | Meaning                           |
| ------ | --------------------------------- |
| 200    | Complete recipe response          |
| 400    | Invalid `recipeId`                |
| 401    | Error body                        |
| 404    | Recipe is missing or soft-deleted |
| 500    | Error body                        |

### `GET /recipes`

Requires `Authorization: Bearer <access-token>` like all recipe routes — this list endpoint is not public. Returns only active recipes.

| Parameter    | Type            | Meaning                                                              |
| ------------ | --------------- | -------------------------------------------------------------------- |
| `name`       | string          | Optional case-insensitive substring filter on recipe name            |
| `difficulty` | string          | Optional difficulty ID filter; must reference an existing difficulty |
| `favorite`   | boolean         | Optional tri-state filter on the caller's favorites — see below      |
| `sort`       | `ASC` or `DESC` | Case-sensitive; orders by `createdAt`; default `DESC`                |
| `page`       | int             | Optional, minimum 1, default 1                                       |
| `limit`      | int             | Optional, 1-100, default 12                                          |

`favorite` is a tri-state filter on `user_favorites` for the authenticated caller: omitted applies no filter, `true` returns only recipes the caller favorited, `false` returns only recipes the caller has **not** favorited.

The success body is `{ "total": 1, "results": [<complete-recipe>, ...] }`; `total` is the count after all filters. Each recipe in `results` includes `isFavorite` and `rating` (see [Representations](#representations)); when `favorite=true` every returned recipe has `isFavorite: true`, and when `favorite=false` every returned recipe has `isFavorite: false`.

| Status | Meaning                                                                                        |
| ------ | ---------------------------------------------------------------------------------------------- |
| 200    | Filtered results                                                                               |
| 400    | Invalid query — malformed `page`/`limit`, unsupported `sort` value, or an unknown `difficulty` |
| 401    | Error body                                                                                     |
| 500    | Error body                                                                                     |

An unknown `difficulty` filter currently returns `400`, not `404`.

Known current limitation: `page` has no effect on the returned rows (a server-side pagination bug bypasses offsetting); every page currently returns the same first `limit` rows ordered by `sort`.

### `PUT /recipes/{recipeId}`

`recipeId` is a required integer. Send a complete write body: omitted required fields are not preserved. The response is the complete replacement recipe, including integer `id`.

| Status | Meaning                                                               |
| ------ | --------------------------------------------------------------------- |
| 200    | Complete recipe response                                              |
| 400    | Invalid path parameter or body                                        |
| 401    | Error body                                                            |
| 403    | Another user owns the active recipe                                   |
| 404    | Recipe is missing/soft-deleted, or difficulty/duration is unavailable |
| 500    | Error body                                                            |

### `DELETE /recipes/{recipeId}`

`recipeId` is a required integer. Deletion sets only `recipes.deleted_at`; that recipe is then omitted from detail and list results.

| Status | Meaning                             |
| ------ | ----------------------------------- |
| 204    | No response body                    |
| 400    | Invalid path parameter              |
| 401    | Error body                          |
| 403    | Another user owns the active recipe |
| 404    | Recipe is missing or soft-deleted   |
| 500    | Error body                          |

For update and delete, ownership is checked after locating an active recipe: another user's active recipe returns `403`; missing or soft-deleted recipes return `404`.

## Favorites

Favoriting is per-user: it marks a `(user_id, recipe_id)` row in `user_favorites` for the authenticated caller. There is no database uniqueness constraint on that pair — duplicate prevention is an application-level check, not a schema-level one. See [`GET /recipes`](#get-recipes) for reading favorite status (the `favorite` filter and the `isFavorite` field on each recipe).

### `POST /recipes/{recipeId}/favorite`

`recipeId` is a required integer. Adds the recipe to the authenticated caller's favorites.

| Status | Meaning                     |
| ------ | --------------------------- |
| 204    | Favorited; no response body |
| 400    | Invalid `recipeId`          |
| 401    | Error body                  |
| 500    | Error body                  |

Calling this repeatedly for the same recipe is a no-op after the first call: if a `user_favorites` row for this `(user_id, recipe_id)` pair already exists, no duplicate row is inserted and the response is still `204`.

### `DELETE /recipes/{recipeId}/favorite`

`recipeId` is a required integer. Removes the recipe from the authenticated caller's favorites by deleting the matching `user_favorites` row outright — a hard delete, not the soft-delete pattern used for recipes (`deleted_at` is never set; the row is gone).

| Status | Meaning                                                  |
| ------ | -------------------------------------------------------- |
| 204    | Unfavorited (or already not favorited); no response body |
| 400    | Invalid `recipeId`                                       |
| 401    | Error body                                               |
| 500    | Error body                                               |

If no matching `user_favorites` row exists, the response is still `204` — not `404` — and no row is deleted.

## Ratings

Rating is per-user: it records a `(user_id, recipe_id)` row in `recipe_ratings` for the authenticated caller. Duplicate prevention is an application-level check (not a database uniqueness constraint) — the service looks for an existing `recipe_ratings` row for that pair before inserting. See [`GET /recipes`](#get-recipes) and [Representations](#representations) for reading the aggregated rating (the `rating.average`/`rating.total` fields on each recipe).

### `POST /recipes/{recipeId}/rating`

`recipeId` is a required integer.

```json
{ "rating": 5 }
```

| Status | Meaning                    |
| ------ | -------------------------- |
| 204    | Rated; no response body    |
| 400    | Invalid `recipeId` or body |
| 401    | Error body                 |
| 500    | Error body                 |

Calling this repeatedly for the same recipe is a no-op after the first call: if a `recipe_ratings` row for this `(user_id, recipe_id)` pair already exists, the new rating is not recorded and the response is still `204` — a caller can only rate a given recipe once.

On a successful (non-duplicate) rating, `recipes.average_rating` for that recipe is recomputed from all its `recipe_ratings` rows and persisted immediately, so the next read of that recipe reflects the new average.

## Users

`{id}` only accepts the literal `me`; the handler substitutes the authenticated user's ID from the middleware-attached context. Any other value returns `404` — there is no lookup-by-arbitrary-ID capability.

### `GET /users/{id}`

| Status | Meaning                |
| ------ | ---------------------- |
| 200    | Complete user response |
| 401    | Error body             |
| 404    | `{id}` is not `me`     |
| 500    | Error body             |

### `PUT /users/{id}`

Send `imageUrl` and `bio`; both are optional and either may be `null` or omitted to clear the field — the two are indistinguishable server-side, so omitting a key has the same effect as sending it as `null`. `name` and `email` are not accepted here — they are sourced from Keycloak and cannot be changed through this API. The response is the complete, updated user representation.

```json
{
  "imageUrl": "https://images.example.com/avatar.jpg",
  "bio": "Home cook who loves spicy food."
}
```

| Status | Meaning                |
| ------ | ---------------------- |
| 200    | Complete user response |
| 400    | Malformed body         |
| 401    | Error body             |
| 404    | `{id}` is not `me`     |
| 500    | Error body             |

## Auth

`/auth/*` routes are exempt from the `Authorization: Bearer <access-token>` requirement in [Conventions](#conventions) — none of them sit behind the JWT middleware, including `POST /auth/logout` despite its generated Swagger doc carrying a `BearerAuth` security annotation (a doc artifact, not an enforced check). These routes drive the Keycloak login handshake and token lifecycle; `GET /auth/login` and `GET /auth/callback` are browser redirects, not JSON endpoints.

State and tickets are single-use values stored in Redis (`oauth_state:<value>` key prefix) and atomically consumed with `GETDEL` — a second use of the same state or ticket always misses and fails with the "invalid or expired" case below, indistinguishable from actual expiry.

Credential:

```json
{
  "accessToken": "eyJhbGci...",
  "refreshToken": "eyJhbGci...",
  "tokenType": "Bearer",
  "expiresAt": "2026-08-22T11:00:00Z"
}
```

### `GET /auth/login`

Generates a random state token, saves it to Redis with a 5-minute TTL, and redirects the browser to Keycloak's authorization endpoint with that state. Not called via `fetch`/`axios` — the frontend navigates the browser here directly.

| Status | Meaning                                    |
| ------ | ------------------------------------------ |
| 302    | Redirect to Keycloak's login page          |
| 500    | Error body (failed to generate/save state) |

### `GET /auth/callback`

Keycloak redirects here itself after the user authenticates — this is not an endpoint the frontend calls directly. Requires `code` and `state` query parameters.

| Parameter | Type   | Meaning                                     |
| --------- | ------ | ------------------------------------------- |
| `code`    | string | Authorization code issued by Keycloak       |
| `state`   | string | Must match the state saved by `/auth/login` |

On success: consumes (`GETDEL`) the saved state, exchanges `code` for tokens with Keycloak, verifies the returned `id_token`, and upserts a local `user` row keyed by the Keycloak subject (`sub`) — creating it on first login or refreshing `email`/`name`/`preferredUsername`/`lastSignedInAt` on subsequent ones. It then mints a one-time ticket, stores the resulting `Credential` in Redis under that ticket for 30 seconds, and redirects to `{FRONTEND_URL}/auth/callback?ticket=<ticket>`.

| Status | Meaning                                                                 |
| ------ | ----------------------------------------------------------------------- |
| 302    | Redirect to `{FRONTEND_URL}/auth/callback?ticket=<ticket>`              |
| 400    | Missing `code` or `state` query parameter                               |
| 401    | `state` not found/expired/already consumed                              |
| 502    | Any other failure — code exchange, `id_token` verify, user upsert, etc. |

### `POST /auth/exchange`

Frontend calls this via axios immediately after being redirected back with `?ticket=...`, trading the short-lived ticket for the real `Credential`.

```json
{ "ticket": "pQx7...base64url..." }
```

Consumes (`GETDEL`) the ticket saved by `/auth/callback`; since that ticket has a 30-second TTL and is single-use, a delayed or repeated call fails with 401.

| Status | Meaning                                   |
| ------ | ----------------------------------------- |
| 200    | `Credential` JSON                         |
| 400    | Missing `ticket` in body                  |
| 401    | Ticket not found/expired/already consumed |
| 500    | Error body                                |

### `POST /auth/logout`

Revokes the given refresh token directly against Keycloak (`POST {realm}/protocol/openid-connect/logout`, form-encoded, with `client_id`/`client_secret`/`refresh_token`). No `Authorization` header is read or required — only the refresh token in the body identifies the session.

```json
{ "refreshToken": "eyJhbGci..." }
```

| Status | Meaning                                                          |
| ------ | ---------------------------------------------------------------- |
| 204    | Logged out; no response body                                     |
| 400    | Missing `refreshToken` in body                                   |
| 502    | Keycloak did not return `204` (invalid token, unreachable, etc.) |

### `POST /auth/refresh-token`

Exchanges a still-valid Keycloak refresh token for a new credential pair (`grant_type=refresh_token` against `POST {realm}/protocol/openid-connect/token`, form-encoded). No `Authorization` header is sent or required — refreshing happens precisely when the caller's access token is gone or expired.

```json
{ "refreshToken": "eyJhbGci..." }
```

Success response is the `Credential` representation above, with `expiresAt` computed as `now + expires_in` from Keycloak's token response.

| Status | Meaning                                                                                                                                                   |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 200    | New `Credential` JSON                                                                                                                                     |
| 400    | Missing `refreshToken` in body                                                                                                                            |
| 401    | Keycloak's token endpoint returned any non-`200` status (invalid/expired/revoked token, or a Keycloak-side failure) — the two cases are not distinguished |
| 500    | Error body (request build/transport failure, or malformed response body from Keycloak)                                                                    |
