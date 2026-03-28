# Hemkop Private API Reference

Base URL: `https://www.hemkop.se`

The API is built on SAP Hybris Commerce (Axfood's commerce platform). It uses two API base paths:
- `/axfood/rest/` - Core REST API (cart, customer, store, auth)
- `/axfoodcommercewebservices/v2/hemkop/` - SAP Hybris Commerce Web Services (CMS content)
- `/search/` - Product search and campaigns

## Authentication

### Session Management
Authentication is session-based using `JSESSIONID` cookie. No Bearer tokens or API keys required.

### Login (Password)

```
POST /login
Content-Type: application/json
```

**Request Body:**
```json
{
  "j_username": "<encrypted_username>",
  "j_username_key": "<encryption_key>",
  "j_password": "<encrypted_password>",
  "j_password_key": "<encryption_key>",
  "j_remember_me": true
}
```

Note: The username and password are encrypted client-side before sending. The encryption keys are random numbers. The encrypted values appear to be base64-encoded strings containing hex-encoded data with a separator `::`.

**Response:**
```json
{"login_successful": "true"}
```

**Response Cookies (set on success):**
- `JSESSIONID` - Session cookie (30 min TTL, renewed on each request)
- `axfoodRememberMe` - Remember-me token (84 day TTL) for auto-login
- `acceleratorSecureGUID` - Secure GUID for the session

### Post-Login Flow
After login, the client performs these steps:
1. `GET /axfood/rest/cart/merge` - Merge anonymous cart with user cart
2. `POST /axfood/rest/cart/restore?action=keepSessionCart` - Restore cart (requires CSRF token)
3. `GET /axfood/rest/customer` - Fetch user profile

### CSRF Token
Required for all POST/PUT/DELETE requests after login:

```
GET /axfood/rest/csrf-token
```

**Response:** A UUID string, e.g. `"45aa4882-36f5-46af-9243-e3b3ed9b91fa"`

Use it as header: `X-Csrf-Token: <token>`

### Get Customer Profile

```
GET /axfood/rest/customer
X-Csrf-Token: <token>
```

**Response (authenticated):**
```json
{
  "uid": "0001gocd",
  "name": "Erik Hellman",
  "firstName": "Erik",
  "lastName": "Hellman",
  "email": "erik@hellman.io",
  "customerId": "0001GOCD",
  "socialSecurityNumer": "1977*******8",
  "storeId": "4003",
  "homeStoreId": "4297",
  "lastUsedLogin": "SSN",
  "isB2BCustomer": false,
  "bonusInfo": {
    "currentTierName": "steg 1",
    "currentBonusAmount": "0",
    "nextTierName": "steg 2",
    "amountToNextTier": "5000",
    "currentBonusLevelEndDate": "31 maj 2026",
    "nextVoucherValue": 15
  },
  "defaultShippingAddress": {
    "line1": "Vivelvägen 46",
    "postalCode": "12533",
    "town": "Älvsjö",
    "cellphone": "+46705777477"
  },
  "savedCards": [...],
  "linkedAccounts": [...]
}
```

**Response (anonymous):**
```json
{
  "uid": "anonymous",
  "name": "anonymous",
  "firstName": "anonymous",
  "lastName": "",
  "newCustomer": true
}
```

---

## Store Search

### List All Stores

```
GET /axfood/rest/store?online=false&clickAndCollect=false&b2bClickAndCollect=false
```

Returns ALL stores as a JSON array. Filtering by name/location is done client-side.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `online` | bool | Filter for online-enabled stores |
| `clickAndCollect` | bool | Filter for click & collect stores |
| `b2bClickAndCollect` | bool | Filter for B2B click & collect |

**Response:** Array of store objects:
```json
[
  {
    "storeId": "4798",
    "name": "Hemköp Alfta Västanågatan",
    "onlineStore": true,
    "clickAndCollect": true,
    "franchiseStore": true,
    "open": true,
    "openingHours": ["Mån 08:00-21:00", ...],
    "openingStoreMessageValue": "8-21",
    "geoPoint": {"latitude": 61.3461, "longitude": 16.0543},
    "address": {
      "line1": "Västanågatan 3",
      "town": "Alfta",
      "postalCode": "822 31",
      "phone": "0271-55230",
      "email": "info.alfta@hemkop.se",
      "formattedAddress": "Västanågatan 3, 822 31 Alfta",
      "latitude": 61.3461,
      "longitude": 16.0543
    },
    "deliveryCost": "79 kr",
    "pickingCostForDelivery": "49 kr",
    "pickingCostForCollect": "49 kr",
    "freeDeliveryThresholdFormatted": "1 900 kr",
    "freePickingCostThresholdForCollectFormatted": "700 kr",
    "customerServicePhone": "0771-55 44 00",
    "customerServiceEmail": "ehandel@hemkop.se",
    "flyerURL": "https://hemkop.eo.se/hkp/4798.pdf"
  }
]
```

### Get Active Store

```
GET /axfood/rest/store/active
```

Returns the currently selected store for the session.

### Get Store by ID

```
GET /axfood/rest/store/{storeId}
```

Returns detailed info for a specific store.

---

## Product Search

### Search Products

```
GET /search/multisearchComplete?q={query}&page={page}&size={size}&show=Page&sort={sort}
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Search query (e.g. "mjolk"). Also supports facet filters via colon syntax: `mjolk:promotionFilter:true` |
| `page` | int | Page number (0-based) |
| `size` | int | Results per page (default: 30) |
| `show` | string | Display mode ("Page") |
| `sort` | string | Sort order (empty = relevance) |

**Response:**
```json
{
  "autocompleteResultData": {
    "suggestions": [],
    "products": null
  },
  "productSearchPageData": {
    "results": [
      {
        "code": "101231504_ST",
        "name": "Mellanmjölk Laktosfri 1,5%",
        "productLine2": "VALIO, 1,5l",
        "manufacturer": "Valio",
        "priceValue": 24.95,
        "price": "24,95 kr",
        "priceUnit": "kr/st",
        "priceNoUnit": "24,95",
        "comparePrice": "16,63 kr",
        "comparePriceUnit": "l",
        "displayVolume": "1,5l",
        "online": true,
        "outOfStock": false,
        "labels": ["swedish_flag", "from_sweden"],
        "googleAnalyticsCategory": "mejeri-ost-och-agg|mjolk|standardmjolk",
        "incrementValue": 1.0,
        "productBasketType": {"code": "ST", "type": "ProductBasketTypeEnum"},
        "image": {
          "url": "https://assets.axfood.se/image/upload/f_auto,t_200/{ean}",
          "format": "product"
        },
        "thumbnail": {
          "url": "https://assets.axfood.se/image/upload/f_auto,t_100/{ean}",
          "format": "thumbnail"
        },
        "potentialPromotions": [],
        "savingsAmount": null,
        "depositPrice": ""
      }
    ],
    "pagination": { ... },
    "facets": [ ... ]
  }
}
```

**Product Code Format:** `{product_id}_{unit}` where unit is `ST` (styck/piece) or `KG` (kilogram).

### Search Campaigns/Promotions

```
GET /search/campaigns/mix?q={storeId}&type=LOYALTY&size=16&onlineSize=0&offlineSize=0&disableMimerSort=true
```

---

## Shopping Cart

### Get Cart

```
GET /axfood/rest/cart
```

Returns the current cart state with all products, prices, discounts, and delivery info.

### Add Products to Cart

```
POST /axfood/rest/cart/addProducts
Content-Type: application/json
X-Csrf-Token: <token>
```

**Request Body:**
```json
{
  "products": [
    {
      "productCodePost": "101231504_ST",
      "qty": 1,
      "pickUnit": "pieces"
    }
  ]
}
```

**Pick Units:**
- `pieces` - for items sold per piece (ST products)
- `kilogram` - for items sold by weight (KG products)

**Response:** Full cart object (same format as GET /axfood/rest/cart) with updated products, totals, etc.

### Merge Cart (after login)

```
GET /axfood/rest/cart/merge
```

Merges the anonymous session cart with the authenticated user's saved cart.

### Restore Cart (after login)

```
POST /axfood/rest/cart/restore?action=keepSessionCart
Content-Type: application/x-www-form-urlencoded
X-Csrf-Token: <token>
```

Empty body. Returns full cart state.

---

## Other Endpoints

### Feature Flags
```
GET /axfood/rest/feature
```

### Category Tree
```
GET /leftMenu/categorytree
```

### External Voucher Count
```
GET /axfood/rest/externalvoucher/count
```

### Favorite Recipes
```
GET /axfood/rest/recipe/getFavorites?page=0&size=999
```

### CMS Components
```
GET /axfoodcommercewebservices/v2/hemkop/cms/components?componentIds={id}&pageSize=1
```

### CMS Pages
```
GET /axfoodcommercewebservices/v2/hemkop/cms/pages?pageType=ContentPage&pageLabelOrId={label}&code=&fields=DEFAULT
```

---

## Common Headers

All API requests include:
- `Cookie: JSESSIONID=<session_id>` (required for auth)
- `Accept: application/json`
- `Content-Type: application/json` (for POST requests)
- `X-Csrf-Token: <token>` (required for POST/PUT/DELETE after login)

## Infrastructure Notes

- Backend: SAP Hybris Commerce (Java), proxied through CloudFront CDN
- Load balancer: AWS ALB (AWSALB/AWSALBCORS cookies)
- Session: Server-side with JSESSIONID cookie (30 min TTL, HttpOnly, Secure, SameSite=None)
- Frontend: Next.js SPA
- Media: Cloudinary (CMS images) and `assets.axfood.se` (product images)
- The API is shared across Axfood brands (Hemköp, Willys) with brand-specific paths
