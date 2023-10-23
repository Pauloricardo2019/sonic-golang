package main

import (
	"fmt"
	"github.com/expectedsh/go-sonic/sonic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

var mapDatabase map[string]Car

type Car struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Brand       string `json:"brand"`
}

func newSimulationDatabase() map[string]Car {
	return map[string]Car{}
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

	defer func() {
		err := sonicSearch.Quit()
		err = sonicIngester.Quit()

		if err != nil {
			panic(err)
		}
	}()

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	router.POST("/car", func(c *gin.Context) {
		car := &Car{}

		if err := c.BindJSON(car); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		car.ID = uuid.New().String()

		mapDatabase[car.ID] = *car
		//Salvar na base

		if err := sonicIngester.Push(
			"vehicles",
			"default",
			car.ID,
			fmt.Sprintf("%s %s %s %s", car.Name, car.Description, car.Color, car.Brand),
			"por"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, car)

	})

	router.GET("/car", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query param q is required"})
			return
		}

		results, err := sonicSearch.Query("vehicles", "default", query, 10, 0, "por")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		products := []Car{}

		//buscar na base
		for _, result := range results {
			v, ok := mapDatabase[result]
			if ok {
				products = append(products, v)
			}
		}

		c.JSON(http.StatusOK, products)
	})

	router.GET("/car/suggests", func(c *gin.Context) {
		suggest := c.Query("suggest")

		if suggest == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query param term is required"})
			return
		}

		results, err := sonicSearch.Suggest(
			"vehicles",
			"default",
			fmt.Sprintf("%s", suggest),
			10,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)

	})

	router.GET("/cars", func(context *gin.Context) {

		count := 0

		cars := []Car{}

		for _, value := range mapDatabase {
			cars = append(cars, value)

			countItem, err := sonicIngester.Count(
				"vehicles",
				"default",
				value.ID,
			)
			if err != nil {
				fmt.Println("error on get count: " + err.Error())
			}
			count = count + countItem

		}

		context.JSON(http.StatusOK, struct {
			Count int   `json:"count"`
			Items []Car `json:"items"`
		}{
			Count: count,
			Items: cars,
		})

	})

	router.DELETE("/car/:id", func(c *gin.Context) {

		carId := c.Param("id")

		if carId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query param term is required"})
			return
		}

		for key, value := range mapDatabase {
			if value.ID == carId {
				delete(mapDatabase, key)
			}
		}

		if err = sonicIngester.FlushObject(
			"vehicles",
			"default",
			carId,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusNoContent, gin.H{"message": "deleted successfully"})

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
