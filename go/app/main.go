package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	// "encoding/json"
	// "io/ioutil"
	"database/sql"
	"strconv"

	// "image"
	"crypto/sha256"
	// "bufio"
	"encoding/hex"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	_ "github.com/mattn/go-sqlite3"
	_ "gorm.io/driver/sqlite"
)

const (
	ImgDir = "image"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
    Name  string `json:"name"`
    Category  string `json:"category"`
}

type ItemJson struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}
var j = []byte(`{"foo":1,"bar":2,"baz":[3,4]}`)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	rawImage := c.FormValue("image")
	openedImage,_ := os.ReadFile("./images/" + rawImage)
	hash := sha256.New()
	hash.Write([]byte(openedImage))
	hashSum := hash.Sum(nil)
	ext := hex.EncodeToString(hashSum[:]) + ".jpg"
	fmt.Printf("%x", hash.Sum(nil))

	db, err := sql.Open("sqlite3", "/Users/karenlaw/mercari/mercari-build-training-2022/db/mercari.sqlite3")
	checkErr(err)
	stmt, _ := db.Prepare("INSERT INTO items (id, name, category, image_filename) VALUES (?, ?, ?, ?)")
	stmt.Exec(nil, name, category, ext)
	defer stmt.Close()
	defer db.Close()

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	// data, err := ioutil.ReadFile("items.json")
    // if err != nil {
    //     fmt.Println(err)
    // }
	// result := fmt.Sprintf(`{"items": [%s, ...]}`, string(data))

	db, err := sql.Open("sqlite3", "/Users/karenlaw/mercari/mercari-build-training-2022/db/mercari.sqlite3")
	checkErr(err)
	rows, _ := db.Query("SELECT * FROM items")
	var id int
	var name string
	var category string
	var result string
	for rows.Next() {
		rows.Scan(&id, &name, &category)
		result = result + fmt.Sprintf(strconv.Itoa(id) + ": " + name + "(" + category +")" + "\n")
	}
	defer db.Close()
	return c.JSON(http.StatusOK, result)
}

func getItemWithItemId(c echo.Context) error {
	inputId := c.Param("item_id")
	db, err := sql.Open("sqlite3", "/Users/karenlaw/mercari/mercari-build-training-2022/db/mercari.sqlite3")
	checkErr(err)
	rows, _ := db.Query("SELECT * FROM items WHERE rowid= ?", inputId)
	var id int
	var name string
	var category string
	var image_filename string
	var message string
	for rows.Next() {
		rows.Scan(&id, &name, &category, &image_filename)
		message = fmt.Sprintf("{'name': %s, 'category': %s, 'image': %s}", name, category, image_filename)
	}
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func searchItem(c echo.Context) error {
	keyword := c.QueryParam("keyword")
	db, err := sql.Open("sqlite3", "/Users/karenlaw/mercari/mercari-build-training-2022/db/mercari.sqlite3")
	checkErr(err)
	rows, _ := db.Query("SELECT * FROM items WHERE name= ?", keyword)
	var id int
	var name string
	var category string
	var result string
	for rows.Next() {
		rows.Scan(&id, &name, &category)
		result = result + fmt.Sprintf(strconv.Itoa(id) + ": " + name + "(" + category +")" + "\n")
	}
	return c.JSON(http.StatusOK, result)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.GET("/items/:item_id", getItemWithItemId)
	e.GET("/search", searchItem)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
