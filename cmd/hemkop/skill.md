---
name: hemkop
description: Use the hemkop CLI to search for grocery products, manage your cart, and browse stores on Hemköp
---
# hemkop CLI

A CLI for interacting with Hemköp's grocery store API. Use this skill when the user wants to search for groceries, manage their shopping cart, or look up store information on Hemköp.

## Setup

The CLI requires Hemköp credentials. They can be provided via:
- Environment variables: `HEMKOP_USERNAME` and `HEMKOP_PASSWORD`
- Flags: `-u <username> -p <password>`
- A `.env` file in the current directory

## Commands

### Search for products
```bash
hemkop product search <query>
```
Returns a table of matching products with code, name, price, compare price, volume, and stock status.

### Show product details
```bash
hemkop product show <product-code>
```
Product codes follow the format `{id}_{unit}` where unit is `ST` (piece) or `KG` (kilogram).

### List stores
```bash
hemkop store list [filter]
```
Optionally filter by store name or town.

### Show store details
```bash
hemkop store show <store-id>
```

### View cart
```bash
hemkop cart list
```

### Add product to cart
```bash
hemkop cart add <product-code> [quantity]
```
Default quantity is 1.

### Remove product from cart
```bash
hemkop cart remove <product-code>
```

### Clear cart
```bash
hemkop cart clear
```

## Workflow tips

- Always search for a product first to find the correct product code before adding to cart.
- Status messages go to stderr, data goes to stdout. Use pipes and redirects to process product data.
- Use `hemkop product search` with broad queries; results are paginated (max 30 per search).
