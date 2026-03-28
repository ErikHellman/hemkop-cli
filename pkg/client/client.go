package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/ErikHellman/hemkop-cli/pkg/crypto"
)

const baseURL = "https://www.hemkop.se"

type Client struct {
	http      *http.Client
	csrfToken string
}

func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http: &http.Client{Jar: jar},
	}
}

// --- Authentication ---

type loginRequest struct {
	Username    string `json:"j_username"`
	UsernameKey string `json:"j_username_key"`
	Password    string `json:"j_password"`
	PasswordKey string `json:"j_password_key"`
	RememberMe  bool   `json:"j_remember_me"`
}

type loginResponse struct {
	LoginSuccessful string `json:"login_successful"`
}

func (c *Client) Login(username, password string) error {
	usernameKey, encryptedUsername, err := crypto.Encrypt(username)
	if err != nil {
		return fmt.Errorf("encrypting username: %w", err)
	}
	passwordKey, encryptedPassword, err := crypto.Encrypt(password)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}

	body := loginRequest{
		Username:    encryptedUsername,
		UsernameKey: usernameKey,
		Password:    encryptedPassword,
		PasswordKey: passwordKey,
		RememberMe:  true,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling login request: %w", err)
	}

	// Hit the home page to establish a session
	resp, err := c.http.Get(baseURL + "/")
	if err != nil {
		return fmt.Errorf("establishing session: %w", err)
	}
	resp.Body.Close()

	req, err := http.NewRequest("POST", baseURL+"/login", strings.NewReader(string(bodyJSON)))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err = c.http.Do(req)
	if err != nil {
		return fmt.Errorf("sending login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("decoding login response: %w", err)
	}

	if loginResp.LoginSuccessful != "true" {
		return fmt.Errorf("login unsuccessful: %s", loginResp.LoginSuccessful)
	}

	csrfToken, err := c.fetchCSRFToken()
	if err != nil {
		return fmt.Errorf("fetching csrf token: %w", err)
	}
	c.csrfToken = csrfToken

	return nil
}

func (c *Client) fetchCSRFToken() (string, error) {
	req, err := http.NewRequest("GET", baseURL+"/axfood/rest/csrf-token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var token string
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}
	return token, nil
}

type Customer struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func (c *Client) GetCustomer() (*Customer, error) {
	req, err := http.NewRequest("GET", baseURL+"/axfood/rest/customer", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.csrfToken != "" {
		req.Header.Set("X-Csrf-Token", c.csrfToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var customer Customer
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// --- Stores ---

type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Address struct {
	Line1            string `json:"line1"`
	Town             string `json:"town"`
	PostalCode       string `json:"postalCode"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	FormattedAddress string `json:"formattedAddress"`
}

type Store struct {
	StoreID       string   `json:"storeId"`
	Name          string   `json:"name"`
	Open          bool     `json:"open"`
	OnlineStore   bool     `json:"onlineStore"`
	OpeningHours  []string `json:"openingHours"`
	GeoPoint      GeoPoint `json:"geoPoint"`
	Address       Address  `json:"address"`
	DeliveryCost  string   `json:"deliveryCost"`
	FranchiseStore bool    `json:"franchiseStore"`
}

func (c *Client) ListStores() ([]Store, error) {
	req, err := http.NewRequest("GET", baseURL+"/axfood/rest/store?online=false&clickAndCollect=false&b2bClickAndCollect=false", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stores []Store
	if err := json.NewDecoder(resp.Body).Decode(&stores); err != nil {
		return nil, err
	}
	return stores, nil
}

func (c *Client) GetStore(storeID string) (*Store, error) {
	req, err := http.NewRequest("GET", baseURL+"/axfood/rest/store/"+storeID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var store Store
	if err := json.NewDecoder(resp.Body).Decode(&store); err != nil {
		return nil, err
	}
	return &store, nil
}

// --- Products ---

type Product struct {
	Code           string   `json:"code"`
	Name           string   `json:"name"`
	ProductLine2   string   `json:"productLine2"`
	Manufacturer   string   `json:"manufacturer"`
	PriceValue     float64  `json:"priceValue"`
	Price          string   `json:"price"`
	PriceUnit      string   `json:"priceUnit"`
	ComparePrice     string   `json:"comparePrice"`
	ComparePriceUnit string  `json:"comparePriceUnit"`
	DisplayVolume    string  `json:"displayVolume"`
	Online         bool     `json:"online"`
	OutOfStock     bool     `json:"outOfStock"`
	Labels         []string `json:"labels"`
	DepositPrice   string   `json:"depositPrice"`
	Image          *Image   `json:"image"`
	ProductBasketType *BasketType `json:"productBasketType"`
}

type Image struct {
	URL    string `json:"url"`
	Format string `json:"format"`
}

type BasketType struct {
	Code string `json:"code"`
}

type Pagination struct {
	CurrentPage  int `json:"currentPage"`
	PageSize     int `json:"pageSize"`
	TotalResults int `json:"totalResults"`
	TotalPages   int `json:"numberOfPages"`
}

type SearchResult struct {
	Results    []Product  `json:"results"`
	Pagination Pagination `json:"pagination"`
}

type searchResponse struct {
	ProductSearchPageData SearchResult `json:"productSearchPageData"`
}

func (c *Client) GetProduct(code string) (*Product, error) {
	req, err := http.NewRequest("GET", baseURL+"/axfood/rest/p/"+code, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product not found: %s (status %d)", code, resp.StatusCode)
	}

	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (c *Client) SearchProducts(query string, page, size int) (*SearchResult, error) {
	url := fmt.Sprintf("%s/search/multisearchComplete?q=%s&page=%d&size=%d&show=Page&sort=", baseURL, query, page, size)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	return &sr.ProductSearchPageData, nil
}

// --- Cart ---

type CartProduct struct {
	Code                string      `json:"code"`
	Name                string      `json:"name"`
	ProductLine2        string      `json:"productLine2"`
	Price               string      `json:"price"`
	PriceValue          float64     `json:"priceValue"`
	Quantity            int         `json:"quantity"`
	TotalPrice          string      `json:"totalPrice"`
	TotalDiscountedPrice string     `json:"totalDiscountedPrice"`
	ComparePrice         string     `json:"comparePrice"`
	ComparePriceUnit     string     `json:"comparePriceUnit"`
	Image               *Image      `json:"image"`
	ProductBasketType   *BasketType `json:"productBasketType"`
}

type Cart struct {
	Code           string        `json:"code"`
	TotalItems     int           `json:"totalItems"`
	TotalUnitCount int           `json:"totalUnitCount"`
	TotalPrice     string        `json:"totalPrice"`
	SubTotalPrice  string        `json:"subTotalPrice"`
	TotalDiscount  string        `json:"totalDiscount"`
	Products       []CartProduct `json:"products"`
}

func (c *Client) GetCart() (*Cart, error) {
	req, err := http.NewRequest("GET", baseURL+"/axfood/rest/cart", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cart Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

type addProductItem struct {
	ProductCodePost string `json:"productCodePost"`
	Qty             int    `json:"qty"`
	PickUnit        string `json:"pickUnit"`
}

type addProductsRequest struct {
	Products []addProductItem `json:"products"`
}

func pickUnitForCode(code string) string {
	if strings.HasSuffix(code, "_KG") {
		return "kilogram"
	}
	return "pieces"
}

func (c *Client) AddToCart(productCode string, qty int) (*Cart, error) {
	body := addProductsRequest{
		Products: []addProductItem{{
			ProductCodePost: productCode,
			Qty:             qty,
			PickUnit:        pickUnitForCode(productCode),
		}},
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/axfood/rest/cart/addProducts", strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.csrfToken != "" {
		req.Header.Set("X-Csrf-Token", c.csrfToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("add to cart failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var cart Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func (c *Client) RemoveFromCart(productCode string) (*Cart, error) {
	return c.AddToCart(productCode, 0)
}

func (c *Client) ClearCart() (*Cart, error) {
	req, err := http.NewRequest("DELETE", baseURL+"/axfood/rest/cart?cancelPossibleContinueCart=false", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.csrfToken != "" {
		req.Header.Set("X-Csrf-Token", c.csrfToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("clear cart failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var cart Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, err
	}
	return &cart, nil
}
