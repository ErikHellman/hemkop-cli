package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ErikHellman/hemkop-cli/pkg/client"
)

func main() {
	args := os.Args[1:]
	username, password := loadCredentials()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	// Parse global flags
	for len(args) > 0 {
		switch args[0] {
		case "-u", "--username":
			if len(args) < 2 {
				fatal("missing value for %s", args[0])
			}
			username = args[1]
			args = args[2:]
		case "-p", "--password":
			if len(args) < 2 {
				fatal("missing value for %s", args[0])
			}
			password = args[1]
			args = args[2:]
		default:
			goto done
		}
	}
done:

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	if username == "" || password == "" {
		fatal("credentials required: set HEMKOP_USERNAME/HEMKOP_PASSWORD env vars, use -u/-p flags, or create a .env file")
	}

	c := client.NewClient()
	fmt.Fprintf(os.Stderr, "Logging in...\n")
	if err := c.Login(username, password); err != nil {
		fatal("login failed: %v", err)
	}

	command := args[0]
	args = args[1:]

	switch command {
	case "store":
		runStore(c, args)
	case "product":
		runProduct(c, args)
	case "cart":
		runCart(c, args)
	case "help":
		printUsage()
	default:
		fatal("unknown command: %s", command)
	}
}

func runStore(c *client.Client, args []string) {
	if len(args) < 1 {
		fatal("usage: hemkop store <list|show> [args]")
	}

	switch args[0] {
	case "list":
		filter := ""
		if len(args) > 1 {
			filter = strings.ToLower(strings.Join(args[1:], " "))
		}
		stores, err := c.ListStores()
		if err != nil {
			fatal("listing stores: %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTOWN\tOPEN\tONLINE")
		for _, s := range stores {
			if filter != "" {
				match := strings.ToLower(s.Name) + " " + strings.ToLower(s.Address.Town)
				if !strings.Contains(match, filter) {
					continue
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.StoreID, s.Name, s.Address.Town,
				boolStr(s.Open), boolStr(s.OnlineStore))
		}
		w.Flush()

	case "show":
		if len(args) < 2 {
			fatal("usage: hemkop store show <store-id>")
		}
		store, err := c.GetStore(args[1])
		if err != nil {
			fatal("getting store: %v", err)
		}
		fmt.Printf("Store:    %s (%s)\n", store.Name, store.StoreID)
		fmt.Printf("Address:  %s\n", store.Address.FormattedAddress)
		fmt.Printf("Phone:    %s\n", store.Address.Phone)
		fmt.Printf("Email:    %s\n", store.Address.Email)
		fmt.Printf("Open:     %s\n", boolStr(store.Open))
		fmt.Printf("Online:   %s\n", boolStr(store.OnlineStore))
		fmt.Printf("Delivery: %s\n", store.DeliveryCost)
		fmt.Println("Hours:")
		for _, h := range store.OpeningHours {
			fmt.Printf("  %s\n", h)
		}

	default:
		fatal("unknown store command: %s", args[0])
	}
}

func runProduct(c *client.Client, args []string) {
	if len(args) < 1 {
		fatal("usage: hemkop product <search|show> [args]")
	}

	switch args[0] {
	case "search":
		if len(args) < 2 {
			fatal("usage: hemkop product search <query>")
		}
		query := strings.Join(args[1:], " ")
		result, err := c.SearchProducts(query, 0, 30)
		if err != nil {
			fatal("searching products: %v", err)
		}

		if len(result.Results) == 0 {
			fmt.Println("No products found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CODE\tNAME\tPRICE\tCOMPARE\tVOLUME\tSTOCK")
		for _, p := range result.Results {
			stock := "In stock"
			if p.OutOfStock {
				stock = "Out of stock"
			}
			if !p.Online {
				stock = "In-store only"
			}
			compare := p.ComparePrice
			if p.ComparePriceUnit != "" {
				compare += "/" + p.ComparePriceUnit
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				p.Code, p.Name, p.Price, compare, p.DisplayVolume, stock)
		}
		w.Flush()
		fmt.Fprintf(os.Stderr, "\nShowing %d of %d results\n", len(result.Results), result.Pagination.TotalResults)

	case "show":
		if len(args) < 2 {
			fatal("usage: hemkop product show <product-code>")
		}
		product, err := c.GetProduct(args[1])
		if err != nil {
			fatal("getting product: %v", err)
		}
		fmt.Printf("Code:         %s\n", product.Code)
		fmt.Printf("Name:         %s\n", product.Name)
		fmt.Printf("Description:  %s\n", product.ProductLine2)
		fmt.Printf("Manufacturer: %s\n", product.Manufacturer)
		fmt.Printf("Price:        %s (%s)\n", product.Price, product.PriceUnit)
		fmt.Printf("Compare:      %s\n", product.ComparePrice)
		fmt.Printf("Volume:       %s\n", product.DisplayVolume)
		fmt.Printf("In stock:     %s\n", boolStr(!product.OutOfStock))
		fmt.Printf("Online:       %s\n", boolStr(product.Online))
		if len(product.Labels) > 0 {
			fmt.Printf("Labels:       %s\n", strings.Join(product.Labels, ", "))
		}
		if product.Image != nil {
			fmt.Printf("Image:        %s\n", product.Image.URL)
		}

	default:
		fatal("unknown product command: %s", args[0])
	}
}

func runCart(c *client.Client, args []string) {
	if len(args) < 1 {
		fatal("usage: hemkop cart <list|add|remove|clear>")
	}

	switch args[0] {
	case "list":
		cart, err := c.GetCart()
		if err != nil {
			fatal("getting cart: %v", err)
		}
		printCart(cart)

	case "add":
		if len(args) < 2 {
			fatal("usage: hemkop cart add <product-code> [quantity]")
		}
		code := args[1]
		qty := 1
		if len(args) > 2 {
			var err error
			qty, err = strconv.Atoi(args[2])
			if err != nil {
				fatal("invalid quantity: %s", args[2])
			}
		}
		cart, err := c.AddToCart(code, qty)
		if err != nil {
			fatal("adding to cart: %v", err)
		}
		fmt.Fprintf(os.Stderr, "Added %d x %s to cart\n", qty, code)
		printCart(cart)

	case "remove":
		if len(args) < 2 {
			fatal("usage: hemkop cart remove <product-code>")
		}
		cart, err := c.RemoveFromCart(args[1])
		if err != nil {
			fatal("removing from cart: %v", err)
		}
		fmt.Fprintf(os.Stderr, "Removed %s from cart\n", args[1])
		printCart(cart)

	case "clear":
		cart, err := c.ClearCart()
		if err != nil {
			fatal("clearing cart: %v", err)
		}
		fmt.Fprintf(os.Stderr, "Cart cleared\n")
		printCart(cart)

	default:
		fatal("unknown cart command: %s", args[0])
	}
}

func printCart(cart *client.Cart) {
	if len(cart.Products) == 0 {
		fmt.Println("Cart is empty.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CODE\tNAME\tQTY\tPRICE\tCOMPARE\tTOTAL")
	for _, p := range cart.Products {
		compare := p.ComparePrice
		if p.ComparePriceUnit != "" {
			compare += "/" + p.ComparePriceUnit
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			p.Code, p.Name, p.Quantity, p.Price, compare, p.TotalDiscountedPrice)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d items, %s\n", cart.TotalUnitCount, cart.TotalPrice)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: hemkop [flags] <command> [args]

Flags:
  -u, --username   Hemköp username (or HEMKOP_USERNAME env var)
  -p, --password   Hemköp password (or HEMKOP_PASSWORD env var)

Commands:
  store list [filter]        List stores, optionally filtered by name/town
  store show <id>            Show store details
  product search <query>     Search for products
  product show <code>        Show product details
  cart list                  Show cart contents
  cart add <code> [qty]      Add product to cart (default qty: 1)
  cart remove <code>         Remove product from cart
  cart clear                 Clear entire cart
  help                       Show this help`)
}

func loadCredentials() (username, password string) {
	username = os.Getenv("HEMKOP_USERNAME")
	password = os.Getenv("HEMKOP_PASSWORD")

	if username != "" && password != "" {
		return
	}

	// Try .env file
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(value, `"'`)
		switch key {
		case "HEMKOP_USERNAME":
			if username == "" {
				username = value
			}
		case "HEMKOP_PASSWORD":
			if password == "" {
				password = value
			}
		}
	}
	return
}

func boolStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
