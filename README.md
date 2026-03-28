# hemkop-cli

[![Release](https://github.com/ErikHellman/hemkop-cli/actions/workflows/release.yml/badge.svg)](https://github.com/ErikHellman/hemkop-cli/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/ErikHellman/hemkop-cli)](https://github.com/ErikHellman/hemkop-cli/releases/latest)
[![License: MIT](https://img.shields.io/github/license/ErikHellman/hemkop-cli)](LICENSE)

A command-line interface for interacting with [Hemköp](https://www.hemkop.se/), the Swedish grocery store. Search for stores and products, manage your shopping cart, and more — all from the terminal.

> **Note:** This tool uses Hemköp's private web API. You need a Klubb Hemköp account with a **password** configured. BankID login is not supported.
>
> If you don't have an account, create one at [hemkop.se](https://www.hemkop.se/registrera/privat/identifiera). If you already have an account but only use BankID, you need to set a password: log in with BankID on the website, go to [Mina Sidor](https://www.hemkop.se/mina-sidor), and set a password under your account settings.

## Install

### macOS and Linux

```sh
curl -fsSL https://raw.githubusercontent.com/ErikHellman/hemkop-cli/main/install.sh | sh
```

This installs the binary to `~/.local/bin`. If that directory isn't in your PATH, the script will tell you how to add it.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ErikHellman/hemkop-cli/main/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\hemkop` and adds it to your user PATH.

### From source

```sh
go install github.com/ErikHellman/hemkop-cli/cmd/hemkop@latest
```

## Configuration

The CLI needs your Hemköp credentials. You can provide them in three ways (checked in this order):

1. **Flags:** `-u <username> -p <password>`
2. **Environment variables:** `HEMKOP_USERNAME` and `HEMKOP_PASSWORD`
3. **`.env` file** in the current directory:
   ```
   HEMKOP_USERNAME="197702065538"
   HEMKOP_PASSWORD="your-password"
   ```

The username is your Swedish personal identity number (personnummer) or Hemköp member number.

## Usage

```
hemkop [flags] <command> [args]
```

### Store commands

**List all stores** (optionally filter by name or town):

```
$ hemkop store list stockholm
ID    NAME                              TOWN         OPEN  ONLINE
4297  Hemköp Stockholm City             Stockholm    Yes   No
4196  Hemköp Stockholm Djurgårdsstaden  Stockholm    Yes   Yes
4127  Hemköp Stockholm Skanstull        Stockholm    Yes   Yes
4938  Hemköp Stockholm Älvsjö           Älvsjö       Yes   No
...
```

**Show store details:**

```
$ hemkop store show 4938
Store:    Hemköp Stockholm Älvsjö (4938)
Address:  Sjättenovembervägen 208, 125 34 Älvsjö
Phone:    08-6471230
Email:    INFO.STOCKHOLM.ALVSJOHALLEN@HEMKOP.SE
Open:     Yes
Online:   No
Delivery: 79 kr
Hours:
  Mån 08:00-21:00
  Tis 08:00-21:00
  ...
```

### Product commands

**Search for products:**

```
$ hemkop product search mjölk
CODE          NAME                                PRICE     COMPARE     VOLUME  STOCK
101231504_ST  Mellanmjölk Laktosfri 1,5%          24,95 kr  16,63 kr/l  1,5l    In stock
101233933_ST  Mellanmjölk Längre Hållbarhet 1,5%  19,50 kr  13,00 kr/l  1,5l    In stock
101205823_ST  Mellanmjölk 1,5%                    13,95 kr  13,95 kr/l  1l      In stock
...
```

**Show product details:**

```
$ hemkop product show 101205823_ST
Code:         101205823_ST
Name:         Mellanmjölk 1,5%
Description:  GARANT, 1l
Manufacturer: Garant
Price:        13,95 kr (kr/st)
Compare:      13,95 kr
Volume:       1l
In stock:     Yes
Online:       Yes
Labels:       swedish_flag, from_sweden
Image:        https://assets.axfood.se/image/upload/...
```

### Cart commands

**Add a product to your cart** (default quantity is 1):

```
$ hemkop cart add 101205823_ST 2
Added 2 x 101205823_ST to cart
CODE          NAME              QTY  PRICE     COMPARE     TOTAL
101205823_ST  Mellanmjölk 1,5%  2    13,95 kr  13,95 kr/l  27,90 kr

Total: 2 items, 27,90 kr
```

**View your cart:**

```
$ hemkop cart list
CODE          NAME              QTY  PRICE     COMPARE     TOTAL
101205823_ST  Mellanmjölk 1,5%  2    13,95 kr  13,95 kr/l  27,90 kr

Total: 2 items, 27,90 kr
```

**Remove a product:**

```
$ hemkop cart remove 101205823_ST
```

**Clear the entire cart:**

```
$ hemkop cart clear
```

## Development

### Prerequisites

- Go 1.25 or later

### Building

```sh
git clone https://github.com/ErikHellman/hemkop-cli.git
cd hemkop-cli
go build -o hemkop ./cmd/hemkop/
```

### Running tests

Unit tests (no credentials needed):

```sh
go test ./pkg/crypto/
```

Integration tests (requires a `.env` file with valid credentials):

```sh
go test -tags=integration ./pkg/client/
```

### Project structure

```
cmd/hemkop/          CLI entry point
pkg/crypto/          Client-side credential encryption (AES-128-CBC + PBKDF2)
pkg/client/          HTTP client for the Hemköp API
API.md               Reverse-engineered API documentation
```

### Creating a release

Releases are automated via GitHub Actions. To create a new release:

```sh
git tag v1.0.0
git push origin v1.0.0
```

This triggers the release workflow which:
1. Runs the test suite
2. Cross-compiles binaries for Linux, macOS, and Windows (amd64 + arm64)
3. Creates a GitHub release with all binaries attached

## License

MIT License. See [LICENSE](LICENSE) for details.
