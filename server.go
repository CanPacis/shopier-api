package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gorilla/handlers"
	mux "github.com/gorilla/mux"
)

type Price struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

type Seller struct {
	ID    string `json:"id"`
	Query string `json:"query"`
	Link  string `json:"link"`
}

type ShortProduct struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Query string `json:"query"`
	Image string `json:"image"`
	Link  string `json:"link"`
	Price Price  `json:"price"`
}

type Product struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Shipping    string   `json:"shipping"`
	Images      []string `json:"images"`
	Price       Price    `json:"price"`
	Seller      Seller   `json:"seller"`
}

type Products struct {
	Content []ShortProduct `json:"content"`
	Length  int            `json:"length"`
}

func main() {
	r := mux.NewRouter()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	methods := handlers.AllowedMethods([]string{"GET", "POST"})
	origins := handlers.AllowedOrigins([]string{"*"})

	r.HandleFunc("/products/{shop}", ProductsHandler)
	r.HandleFunc("/product/{id}", ProductHandler)

	http.Handle("/", r)

	fmt.Println("Server is up and listening on port 8000")
	loggedRouter := handlers.LoggingHandler(os.Stdout, handlers.CORS(headers, methods, origins)(r))
	http.ListenAndServe(":8000", loggedRouter)
}

func ProductHandler(w http.ResponseWriter, r *http.Request) {

	var product Product
	var images []string
	var c = colly.NewCollector()
	idRaw := mux.Vars(r)["id"]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	c.OnHTML("div.product-page-container", func(e *colly.HTMLElement) {
		id, err := strconv.Atoi(idRaw)

		if err != nil {
			panic(err)
		}

		price, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(strings.Split(e.ChildText(".product-info__price"), "TL")[0]), ",", "."), 32)

		if err != nil {
			panic(err)
		}

		sellerID := strings.Split(e.ChildAttr(".cart-seller__link", "href"), "shop=")[1]

		product = Product{
			ID:          id,
			Title:       e.ChildText(".product-info__title"),
			Description: e.ChildText("#tab-description p"),
			Shipping:    e.ChildText("#tab-shipping p"),
			Seller: Seller{
				ID:    sellerID,
				Query: "http://localhost:8000/products/" + sellerID,
				Link:  "https://shopier.com/storefront.php?shop=" + sellerID,
			},
			Price: Price{Type: "TL", Amount: math.Floor(price*100) / 100},
		}
	})

	c.OnHTML(".product-swiper-thumbs .product__image", func(e *colly.HTMLElement) {
		images = append(images, e.Attr("src"))
	})

	c.Visit("https://www.shopier.com/ShowProductNew/products.php?id=" + idRaw + "&sid=aWVHQ1dqd1o1RGFUYlZ1NjBfMF8gXyA=")

	product.Images = images
	data, err := json.Marshal(product)

	if err != nil {
		panic(err)
	}

	w.Write(data)
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	var products []ShortProduct
	var c = colly.NewCollector()
	shop := mux.Vars(r)["shop"]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	c.OnHTML("div.product", func(e *colly.HTMLElement) {
		rawId := strings.Split(e.ChildAttr(".product_name_url", "href"), "=")[1]
		id, err := strconv.Atoi(rawId)

		if err != nil {
			panic(err)
		}

		price, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(strings.Split(e.ChildText(".product__price"), "TL")[0]), ",", "."), 32)

		if err != nil {
			panic(err)
		}

		products = append(products, ShortProduct{
			ID:    id,
			Image: e.ChildAttr("img", "src"),
			Title: e.ChildText(".product__title"),
			Query: "http://localhost:8000/product/" + rawId,
			Link:  "https://www.shopier.com/" + rawId,
			Price: Price{Type: "TL", Amount: math.Floor(price*100) / 100},
		})
	})

	c.Visit("https://www.shopier.com/ShowProductNew/storefront.php?shop=" + shop + "&sid=VUpoM2Z1Rzl6a0ZuTGM1ZzExXy0xXyBfIA==")

	response := Products{
		Content: products,
		Length:  len(products),
	}

	data, err := json.Marshal(response)

	if err != nil {
		panic(err)
	}

	w.Write(data)
}
