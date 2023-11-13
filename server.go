package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"fmt"
	"os"
	//"io/ioutil"
	//"path"

	photoprism "github.com/drummonds/photoprism-client-go"
	// "github.com/kris-nova/logger"
	"github.com/drummonds/photoprism-client-go/api/v1"
)

func main() {
	var (
		user = os.Getenv("PP_USERNAME")
		password = os.Getenv("PP_PASSWORD")
		uri = os.Getenv("PP_URI")
	)

	// uuid := "PS416FO13JRX04IV" // This is a known ID
    client := photoprism.New(uri)
	err := client.Auth(photoprism.NewClientAuthLogin(user, password))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.V1().Index()
	if err != nil {
		fmt.Println(err)
	}

	e := echo.New()
	e.GET("/listalbums", func(c echo.Context) error {
		albums, err := client.V1().GetAlbums(nil)
		if err != nil {
			fmt.Println(err)

		}

		return c.JSON(http.StatusOK, albums)
	})

	e.GET("/album/:id", func(c echo.Context) error {
		id := c.Param("id")
		album, err := client.V1().GetAlbum(id)
		if err != nil {
			fmt.Println(err)
		}

		return c.JSON(http.StatusOK, album)
	})

	e.GET("/albumphotos/:id", func(c echo.Context) error {
		id := c.Param("id")
		photos, err := client.V1().GetPhotos(&api.PhotoOptions{
			Count: 1000,
			AlbumUID: id,
		})
		if err != nil {
			fmt.Println(err)
		}

		return c.JSON(http.StatusOK, photos)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
