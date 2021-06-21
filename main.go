package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/peteretelej/jsonbox"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func String(length int) string {
	return GetRandomStringWithCharset(length, charset)
}

func GetRandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func IsExists(array []Url, item string, key string) bool {
	for _, elem := range array {
		if key == "shorten" && elem.Shorten == item {
			IsExists(array, item, "shorten")
		} else if key == "custom" && elem.Shorten == item {
			return true
		}
	}

	return false
}

func IsExistsOriginalUrl(array []Url, item string) string {
	for _, elem := range array {
		if elem.Original == item {
			return elem.Shorten
		}
	}

	return ""
}

type Url struct {
	Id        string `json:"_id"`
	Original  string `json:"original"`
	Shorten   string `json:"shorten"`
	CreatedAt string `json:"_createdOn"`
}

func Shorten(c *gin.Context) {
	url := c.PostForm("url")
	if url == "" {
		c.JSON(400, gin.H{"Status": "FAILED", "Message": "No original URL found."})
		return
	}

	customUrl := c.PostForm("custom_url")
	apiKey := os.Getenv("SHORTEN_URL_AP_KEYI")
	cl, _ := jsonbox.NewClient("https://jsonbox-tyroz.herokuapp.com/")

	urlArr, _ := cl.Read(apiKey)
	urlArrByte := []byte(urlArr)
	urls := []Url{}
	_ = json.Unmarshal(urlArrByte, &urls)

	respShortenUrl := IsExistsOriginalUrl(urls, url)
	if respShortenUrl != "" {
		c.JSON(400, gin.H{"Status": "FAILED", "Message": "This URL has already exists.", "ShortenUrl": respShortenUrl})
		return
	}

	shorten := ""
	isExistShorten := false
	if customUrl == "" {
		shorten = GetRandomStringWithCharset(5, charset)

		isExistShorten = IsExists(urls, shorten, "shorten")
		if isExistShorten {
			c.JSON(400, gin.H{"Status": "ERROR", "Message": "Shorten link has Error!"})
			return
		}
	} else {
		shorten = customUrl

		isExistShorten = IsExists(urls, shorten, "custom")
		if isExistShorten {
			c.JSON(400, gin.H{"Status": "FAILED", "Message": "Custom URL has already exists."})
			return
		}
	}

	urlObj := []byte(`{"original": "` + url + `", "shorten": "` + shorten + `"}`)
	out, _ := cl.Create(apiKey, urlObj)

	if out == nil {
		c.JSON(400, gin.H{"Status": "ERROR", "Message": "Create a shorten link has error."})
		return
	} else {
		c.JSON(200, gin.H{"Status": "SUCCESS", "Message": "This URL has been shortened.", "ShortenUrl": shorten})
		return
	}
}

func RedirectOriginalUrl(c *gin.Context) {
	shortenUrl := c.Param("shorten_url")

	apiKey := os.Getenv("SHORTEN_URL_AP_KEYI")
	cl, _ := jsonbox.NewClient("https://jsonbox-tyroz.herokuapp.com/")

	urlArr, _ := cl.Read(apiKey)
	urlArrByte := []byte(urlArr)
	urls := []Url{}
	_ = json.Unmarshal(urlArrByte, &urls)

	responseOriginalUrl := ""
	for _, elem := range urls {
		if elem.Shorten == shortenUrl {
			responseOriginalUrl = elem.Original
		}
	}

	if responseOriginalUrl == "" {
		c.JSON(400, gin.H{"Status": "ERROR", "Message": "Not found."})
	} else {
		c.JSON(200, gin.H{"Status": "SUCCESS", "Message": "Success", "OriginalUrl": responseOriginalUrl})
	}
}

func CheckHealth(c *gin.Context) {
	c.JSON(200, gin.H{"Status": "Good"})
}

func ClearAll(c *gin.Context) {
	password := c.Param("password")

	apiKey := os.Getenv("SHORTEN_URL_AP_KEYI")
	cl, _ := jsonbox.NewClient("https://jsonbox-tyroz.herokuapp.com/")

	if password != "" && password == "save-sut" {
		err := cl.DeleteAll(apiKey)
		if err != nil {
			panic(err)
		}

		c.JSON(200, gin.H{"Status": "SUCCESS", "Message": "All data has been deleted"})
	} else {
		c.JSON(400, gin.H{"Status": "FAILED", "Message": "Wrong password or Password is not entered"})
	}
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	api := r.Group("/api")
	{
		api.GET("/", CheckHealth)
		api.POST("/shorten-url", Shorten)
		api.GET("/shorten/:shorten_url/to/original-url", RedirectOriginalUrl)
		api.GET("clear-all/:password", ClearAll)
	}
	r.Run()
}
