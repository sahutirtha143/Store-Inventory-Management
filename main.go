package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type InventoryItem struct {
	ID           int
	Date         string
	Item         string
	Quantity     float64
	PricePerUnit float64
	// TotalPrice        float64
	TotalPriceBuying  float64
	TotalPriceSelling float64
}

var (
	db               *sql.DB
	mu               sync.Mutex
	inventory        []InventoryItem
	sellingInventory []InventoryItem // To store selling inventory
)

func main() {
	var err error
	// Connect to the MySQL database
	dsn := "root:2112@tcp(localhost:3306)/inventory_db" // Replace with your credentials
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// Initialize tables if they don't exist
	initDB()

	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/add-item", addItemHandler)
	http.HandleFunc("/add-item-selling", addItemHandlerSelling)

	// Add delete handlers
	http.HandleFunc("/delete-item-buy", deleteItemHandlerBuy)
	http.HandleFunc("/delete-item-sell", deleteItemHandlerSell)

	//GRAPH
	http.HandleFunc("/inventory-data", inventoryDataHandler)

	// Static files(CSS and JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server started at :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func initDB() {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS inventory (id INT AUTO_INCREMENT PRIMARY KEY, date DATE NOT NULL, item VARCHAR(100) NOT NULL, quantity FLOAT NOT NULL, price_per_unit FLOAT NOT NULL, total_price FLOAT NOT NULL)")
	if err != nil {
		fmt.Println("Error initializing inventory table:", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS selling_inventory (id INT AUTO_INCREMENT PRIMARY KEY, date DATE NOT NULL, item VARCHAR(100) NOT NULL, quantity FLOAT NOT NULL, price_per_unit FLOAT NOT NULL, total_price FLOAT NOT NULL)")
	if err != nil {
		fmt.Println("Error initializing selling inventory table:", err)
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.ParseFiles("templates/index.html"))

	// Fetch inventory and selling inventory from database
	inventory, totalPriceAllItems := fetchInventory()
	sellingInventory, totalPriceAllSellingItems := fetchSellingInventory()

	// Calculate profit or loss
	profitLoss := calculateProfitLoss()

	// Pass the total price along with inventory data
	data := struct {
		Inventory            []InventoryItem
		SellingInventory     []InventoryItem
		TotalPriceAllItems   float64
		TotalPriceAllSelling float64
		ProfitLoss           float64
	}{
		Inventory:            inventory,
		SellingInventory:     sellingInventory,
		TotalPriceAllItems:   totalPriceAllItems,
		TotalPriceAllSelling: totalPriceAllSellingItems,
		ProfitLoss:           profitLoss,
	}

	t.Execute(w, data)
}

func addItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()

		// Create a new inventory item from form data
		item := InventoryItem{
			Date:         r.FormValue("date"),
			Item:         r.FormValue("item"),
			Quantity:     parsePrice(r.FormValue("quantity")),
			PricePerUnit: parsePrice(r.FormValue("price")),
		}

		// Calculate total buy price for the item
		item.TotalPriceBuying = item.Quantity * item.PricePerUnit

		mu.Lock()
		// Add the item to the inventory in the database
		_, err := db.Exec("INSERT INTO inventory (date, item, quantity, price_per_unit, total_price) VALUES (?, ?, ?, ?, ?)", item.Date, item.Item, item.Quantity, item.PricePerUnit, item.TotalPriceBuying)
		mu.Unlock()

		if err != nil {
			http.Error(w, "Error adding item to inventory", http.StatusInternalServerError)
			return
		}

		// Redirect back to the main page to show the updated inventory
		http.Redirect(w, r, "/#buy", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func addItemHandlerSelling(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()

		// Create a new inventory item from form data
		item := InventoryItem{
			Date:         r.FormValue("date"),
			Item:         r.FormValue("item"),
			Quantity:     parsePrice(r.FormValue("quantity")),
			PricePerUnit: parsePrice(r.FormValue("price")),
		}

		// Calculate total price for the item
		item.TotalPriceSelling = item.Quantity * item.PricePerUnit

		mu.Lock()
		// Add the item to the selling inventory in the database
		_, err := db.Exec("INSERT INTO selling_inventory (date, item, quantity, price_per_unit, total_price) VALUES (?, ?, ?, ?, ?)", item.Date, item.Item, item.Quantity, item.PricePerUnit, item.TotalPriceSelling)
		mu.Unlock()

		if err != nil {
			http.Error(w, "Error adding item to selling inventory", http.StatusInternalServerError)
			return
		}

		// Redirect back to the main page to show the updated selling inventory
		http.Redirect(w, r, "/#sell", http.StatusSeeOther)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Function to parse the price
func parsePrice(value string) float64 {
	var price float64
	fmt.Sscanf(value, "%f", &price)
	return price
}

// Function to fetch all items from inventory
func fetchInventory() ([]InventoryItem, float64) {
	rows, err := db.Query("SELECT id, date, item, quantity, price_per_unit, total_price FROM inventory")
	if err != nil {
		fmt.Println("Error fetching inventory:", err)
		return nil, 0
	}
	defer rows.Close()

	var inventory []InventoryItem
	var totalPriceAllItems float64

	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ID, &item.Date, &item.Item, &item.Quantity, &item.PricePerUnit, &item.TotalPriceBuying); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}
		inventory = append(inventory, item)
		totalPriceAllItems += item.TotalPriceBuying
	}

	return inventory, totalPriceAllItems
}

// Function to fetch all items from selling inventory
func fetchSellingInventory() ([]InventoryItem, float64) {
	rows, err := db.Query("SELECT id, date, item, quantity, price_per_unit, total_price FROM selling_inventory")
	if err != nil {
		fmt.Println("Error fetching selling inventory:", err)
		return nil, 0
	}
	defer rows.Close()

	var sellingInventory []InventoryItem
	var totalPriceAllSellingItems float64

	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ID, &item.Date, &item.Item, &item.Quantity, &item.PricePerUnit, &item.TotalPriceSelling); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}
		sellingInventory = append(sellingInventory, item)
		totalPriceAllSellingItems += item.TotalPriceSelling
	}

	return sellingInventory, totalPriceAllSellingItems
}

// Function to calculate profit or loss
func calculateProfitLoss() float64 {
	_, totalPriceAllItems := fetchInventory()

	_, totalPriceAllSellingItems := fetchSellingInventory()

	// Calculate profit or loss
	profitLoss := totalPriceAllSellingItems - totalPriceAllItems

	return profitLoss
}

// DELETE RECORDE FOR THE TABLE:
func deleteItemHandlerBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		itemID := r.FormValue("id")

		mu.Lock()
		// Delete the item from the inventory based on the ID
		_, err := db.Exec("DELETE FROM inventory WHERE id = ?", itemID)
		mu.Unlock()

		if err != nil {
			http.Error(w, "Error deleting item from inventory", http.StatusInternalServerError)
			return
		}

		// Redirect back to the main page to show the updated inventory
		http.Redirect(w, r, "/#buy", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func deleteItemHandlerSell(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		itemID := r.FormValue("id")

		mu.Lock()
		// Delete the item from the selling inventory based on the ID
		_, err := db.Exec("DELETE FROM selling_inventory WHERE id = ?", itemID)
		mu.Unlock()

		if err != nil {
			http.Error(w, "Error deleting item from selling inventory", http.StatusInternalServerError)
			return
		}

		// Redirect back to the main page to show the updated selling inventory
		http.Redirect(w, r, "/#sell", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler to serve inventory data for the chart(GRAPH)
func inventoryDataHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Fetch inventory and selling inventory data
	inventory, totalPriceAllItems := fetchInventory()
	sellingInventory, totalPriceAllSellingItems := fetchSellingInventory()

	profitLoss := calculateProfitLoss()

	// Create a response structure
	data := struct {
		InventoryItems        []InventoryItem `json:"inventory"`
		SellingInventoryItems []InventoryItem `json:"selling_inventory"`
		TotalBuyPrice         float64         `json:"total_buy_price"`
		TotalSellPrice        float64         `json:"total_sell_price"`
		ProfitLoss            float64         `json:"profit_loss"`
	}{
		InventoryItems:        inventory,
		SellingInventoryItems: sellingInventory,
		TotalBuyPrice:         totalPriceAllItems,
		TotalSellPrice:        totalPriceAllSellingItems,
		ProfitLoss:            profitLoss,
	}

	// Convert the data to JSON and send it
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}
