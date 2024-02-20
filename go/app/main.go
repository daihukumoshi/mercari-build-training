package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"io"
	"crypto/sha256"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name string `json:"name"`
	Category string `json:"category"`
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

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	//ここからStep2
	//jsonファイルの読み込み
	raw, err := os.ReadFile("./items.json")
	//エラーハンドリング
	if err != nil {
        fmt.Println("Read error")
    }
	//jsonに対応した構造体のインスタンスを生成
	var itemsWrapper ItemsWrapper
	//jsonをデコード
	if err := json.Unmarshal(raw, &itemsWrapper); err != nil {
		fmt.Println("Decode error")
		fmt.Println(err)
	}

	//ここから画像
	// フォームのファイルを取得
	file, err := c.FormFile("image")
	if err != nil {
		fmt.Println("Image Get error")
	}
	//ファイルを開いて内容をsrcに代入
	src, err := file.Open()
	if err != nil {
		fmt.Println("Image Open error")
	}
	defer src.Close()

	fileModel := strings.Split(file.Filename, ".")
	fileName := fileModel[0]
	extension := fileModel[1]
	fileName_hash := sha_256([]byte(fileName))

	// 保存用のディレクトリを作成する（存在していなければ、保存用のディレクトリを新規作成）
	err = os.MkdirAll("./images", os.ModePerm)
	if err != nil {
		fmt.Println("Mkdir error")
	}

	// 保存用のファイルを作成する
	dst, err := os.Create("./images/"+fmt.Sprintf("%s.%s",string(fileName_hash),extension))
	if err != nil {
		fmt.Println("Mkfile error")
		fmt.Println(err)
	}
	defer dst.Close()

	// アップロードされたファイルの内容を保存用のファイルにコピーする
	_, err = io.Copy(dst, src)
	if err != nil {
		fmt.Println("Copyfile error")
	}
	fmt.Println("アップロード成功!")
	// ここまで画像
	var imageName string = fileName_hash + "." + extension
	//新しい商品（ジャンル）の読み取り
	category := c.FormValue("category")
	//新しい商品を構造体に
	newItem := Item{Name: name, Category: category, Imagename: imageName}
	//新しい商品を商品一覧配列に追加
	itemsWrapper.Items = append(itemsWrapper.Items, newItem)

	//商品一覧配列をjson化
	ans, err := json.Marshal(itemsWrapper)
	if err != nil {
		fmt.Println("Encode error")
	}
	os.WriteFile("./items.json", []byte(ans), 0664)
	if err != nil {
		fmt.Println("WriteFile error")
	}

	return c.JSON(http.StatusOK, res)
}

func showItem (c echo.Context) error {
	//jsonファイルの読み込み
	raw, err := os.ReadFile("./items.json")
	//エラーハンドリング
	if err != nil {
        fmt.Println("Road error")
    }
	var itemsWrapper ItemsWrapper
	if err := json.Unmarshal(raw, &itemsWrapper); err != nil {
		fmt.Println("Decode error")
		fmt.Println(err)
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
	image_id := c.Param("image_id")

	//jsonファイルの読み込み
	raw, err := os.ReadFile("./items.json")
	//エラーハンドリング
	if err != nil {
        fmt.Println("Road error")
    }
	//jsonに対応した構造体のインスタンスを生成
	var itemsWrapper ItemsWrapper
	//jsonをデコード
	if err := json.Unmarshal(raw, &itemsWrapper); err != nil {
		fmt.Println("Decode error")
		fmt.Println(err)
	}

	items := itemsWrapper.Items
	index, _:= strconv.Atoi(image_id)
	if len(items) < index {
		message := fmt.Sprintf("noImage")
		res := Response{Message: message}
		return c.JSON(http.StatusOK, res)
	}else {
		item := items[index]
		return c.JSON(http.StatusOK, item)
	}
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
	e.GET("/items/:image_id", getItems)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}