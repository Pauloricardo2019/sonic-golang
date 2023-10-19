package main

import (
	"fmt"
	"github.com/expectedsh/go-sonic/sonic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

var mapDatabase map[string]Product

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func newSimulationDatabase() map[string]Product {
	return map[string]Product{}
}

func init() {
	mapDatabase = newSimulationDatabase()
}

func main() {
	host := "localhost"
	port := 1491
	password := "SecretPassword"

	sonicIngester, err := newIngester(host, port, password)
	if err != nil {
		panic(err)
	}

	sonicSearch, err := newSearcher(host, port, password)
	if err != nil {
		panic(err)
	}

	if err = sonicIngester.Ping(); err != nil {
		fmt.Println(err.Error())
	}

	if err = sonicSearch.Ping(); err != nil {
		fmt.Println(err.Error())
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	router.POST("/product", func(c *gin.Context) {
		product := &Product{}

		if err := c.BindJSON(product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		product.ID = uuid.New().String()

		mapDatabase[product.ID] = *product
		//Salvar na base

		if err := sonicIngester.Push(
			"products",
			"default",
			product.ID,
			fmt.Sprintf("%s %s", product.Name, product.Description),
			"por"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, product)

	})

	router.GET("/product", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query param q is required"})
			return
		}

		results, err := sonicSearch.Query("products", "default", query, 10, 0, "por")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		products := []Product{}

		//buscar na base
		for _, result := range results {
			v, ok := mapDatabase[result]
			if ok {
				products = append(products, v)
			}
		}

		c.JSON(http.StatusOK, products)
	})

	router.Run(":3333")
}

func newIngester(host string, port int, pass string) (sonic.Ingestable, error) {

	ingester, err := sonic.NewIngester(host, port, pass)
	if err != nil {
		return nil, err
	}

	return ingester, nil
}

func newSearcher(host string, port int, pass string) (sonic.Searchable, error) {
	searcher, err := sonic.NewSearch(host, port, pass)
	if err != nil {
		return nil, err
	}

	return searcher, nil
}
