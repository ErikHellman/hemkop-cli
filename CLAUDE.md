# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```sh
go build -o hemkop ./cmd/hemkop/          # Build the CLI binary
go test ./pkg/crypto/                      # Unit tests (encryption round-trips, padding)
go test -tags=integration ./pkg/client/    # Integration tests (requires .env with credentials)
go test -v -run TestEncryptDeterministic ./pkg/crypto/  # Run a single test
go vet ./...                               # Lint
```

Integration tests require a `.env` file in the project root with `HEMKOP_USERNAME` and `HEMKOP_PASSWORD`.

## Architecture

The project is a CLI for Hemköp's private web API (reverse-engineered from their Next.js frontend).

**Three layers:**

1. **`cmd/hemkop/main.go`** — CLI entry point. Parses args (no framework, just `os.Args` + switch), loads credentials (flags → env vars → `.env` file), logs in, dispatches to command handlers, formats output with `tabwriter`. Status messages go to stderr, data to stdout.

2. **`pkg/client/`** — HTTP client wrapping Hemköp's API. Uses `cookiejar` for automatic session management (JSESSIONID). Login encrypts credentials client-side, then fetches a CSRF token needed for all mutations (POST/DELETE). All API types (Store, Product, Cart, etc.) are defined here.

3. **`pkg/crypto/`** — Client-side credential encryption matching the frontend's JavaScript (module 37068). Algorithm: PBKDF2(random 16-digit key, random salt, 1000 iter, SHA-1) → AES-128-CBC. Output format: `base64(hex(iv)::hex(salt)::base64(ciphertext))`.

**Key API patterns:**
- Base URL: `https://www.hemkop.se`
- Core REST: `/axfood/rest/` (cart, customer, store, product detail at `/axfood/rest/p/{code}`)
- Search: `/search/multisearchComplete?q=...&page=0&size=30&show=Page&sort=`
- Cart mutations use `POST /axfood/rest/cart/addProducts` with `X-Csrf-Token` header. Setting `qty: 0` removes an item. `DELETE /axfood/rest/cart` clears the cart.
- Product codes follow the format `{id}_{unit}` where unit is `ST` (piece) or `KG` (kilogram), which determines the `pickUnit` field (`pieces` or `kilogram`).

**Single external dependency:** `golang.org/x/crypto` for PBKDF2.
