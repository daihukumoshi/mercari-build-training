package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	Imagename string `json:"image_name"`
}

type ItemsWrapper struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func sha_256(target []byte) string {
	//ハッシュ化
	h := sha256.New()
	h.Write(target)
	fmt.Printf("%x", h.Sum(nil))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func handleError(c echo.Context, err error) error {
	c.Logger().Error(err)
	return c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
}

func openDatabase() (*sql.DB, error) {
	return sql.Open("sqlite3", "./db/mercari.sqlite3")
}

func addItem(c echo.Context) error {

	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s", name)
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	//ここから画像
	// フォームのファイルを取得
	file, err := c.FormFile("image")
	if err != nil {
		return handleError(c, err)
	}
	//ファイルを開いて内容をsrcに代入
	src, err := file.Open()
	if err != nil {
		return handleError(c, err)
	}
	defer src.Close()

	fileModel := strings.Split(file.Filename, ".")
	fileName := fileModel[0]
	extension := fileModel[1]
	fileName_hash := sha_256([]byte(fileName))

	// 保存用のファイルを作成する
	dst, err := os.Create("./images/" + fmt.Sprintf("%s.%s", string(fileName_hash), extension))
	if err != nil {
		return handleError(c, err)
	}
	defer dst.Close()

	// アップロードされたファイルの内容を保存用のファイルにコピーする
	_, err = io.Copy(dst, src)
	if err != nil {
		return handleError(c, err)
	}
	c.Logger().Infof("アップロード成功!")
	// ここまで画像
	imageName := fileName_hash + "." + extension

	//dbに保存
	db, err := openDatabase()
	if err != nil {
		return handleError(c, err)
	}
	defer db.Close()

	var category_id int
	var category_name string

	//カテゴリ名が一致するcategoriesのレコード取得
	err = db.QueryRow("SELECT id, name FROM categories WHERE name = ?", category).Scan(&category_id, &category_name)
	if err != nil {
		//一致するカテゴリ名がなければ作る
		res, err := db.Exec("INSERT INTO categories (name) VALUES (?)", category)
		if err != nil {
			return handleError(c, err)
		}
		//uuidの取得
		newCategoryID, err := res.LastInsertId()
		if err != nil {
			return handleError(c, err)
		}
		category_id = int(newCategoryID)
	}

	stmt, err := db.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		return handleError(c, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, category_id, imageName)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, res)
}

func showItem(c echo.Context) error {
	db, err := openDatabase()
	if err != nil {
		return handleError(c, err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT items.name, items.image_name, categories.name  FROM items INNER JOIN categories ON items.category_id = categories.id")
	if err != nil {
		return handleError(c, err)
	}
	defer rows.Close()

	var itemsWrapper ItemsWrapper

	for rows.Next() {
		var itemName string
		var image_name string
		var categoryName string

		err = rows.Scan(&itemName, &image_name, &categoryName)
		if err != nil {
			return handleError(c, err)
		}

		newItem := Item{Name: itemName, Category: categoryName, Imagename: image_name}
		itemsWrapper.Items = append(itemsWrapper.Items, newItem)
	}

	if err = rows.Err(); err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, itemsWrapper)
}

func searchItem(c echo.Context) error {
	keyword := c.QueryParam("keyword")

	db, err := openDatabase()
	if err != nil {
		return handleError(c, err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT items.name, items.image_name, categories.name FROM items INNER JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?", "%"+keyword+"%")
	if err != nil {
		return handleError(c, err)
	}
	defer rows.Close()

	var itemsWrapper ItemsWrapper

	for rows.Next() {
		var itemName string
		var image_name string
		var categoryName string

		err = rows.Scan(&itemName, &image_name, &categoryName)
		if err != nil {
			return handleError(c, err)
		}
		newItem := Item{Name: itemName, Category: categoryName, Imagename: image_name}
		itemsWrapper.Items = append(itemsWrapper.Items, newItem)
	}
	return c.JSON(http.StatusOK, itemsWrapper)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))
	fmt.Println(imgPath)

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().SetLevel(log.DEBUG)
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}

	return c.File(imgPath)
}

func getItems(c echo.Context) error {
	db, err := openDatabase()
	if err != nil {
		return handleError(c, err)
	}
	defer db.Close()

	itemID, err := strconv.Atoi(c.Param("item_id"))
	if err != nil {
		return handleError(c, err)
	}

	rows, err := db.Query("SELECT items.name, items.image_name, categories.name FROM items INNER JOIN categories ON items.category_id = categories.id WHERE items.id=?", itemID)
	if err != nil {
		return handleError(c, err)
	}
	defer rows.Close()

	var itemsWrapper ItemsWrapper

	for rows.Next() {
		var itemName string
		var image_name string
		var categoryName string

		err = rows.Scan(&itemName, &image_name, &categoryName)
		if err != nil {
			return handleError(c, err)
		}

		newItem := Item{Name: itemName, Category: categoryName, Imagename: image_name}
		itemsWrapper.Items = append(itemsWrapper.Items, newItem)
	}

	return c.JSON(http.StatusOK, itemsWrapper)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", showItem)
	e.POST("/items", addItem)
	e.GET("/items/:item_id", getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search", searchItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
