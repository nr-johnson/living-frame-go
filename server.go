package main

import (
	"html/template"
	"net/http"

	"encoding/json"
	"github.com/labstack/echo/v4"
	"fmt"
	"os"
	"os/exec"
	"errors"
	"io"
	"io/ioutil"
	"path"
	"strings"

	photoprism "github.com/drummonds/photoprism-client-go"
	// "github.com/kris-nova/logger"
	"github.com/drummonds/photoprism-client-go/api/v1"

	"github.com/nr-johnson/living-frame-go/myFunctions"
)

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Logged_in bool `json:"logged_in"`
	Network string `json:"network"`
	Network_password string `json:"network_password"`
	Connected bool `json:"connected"`
	Uri string `json:"uri"`
	Delay string `json:"delay"`
	Fade string `json:"fade"`
	Configured bool `json:"configured"`
}

type FormError struct {
	Field string
	Error string
}

var (
	albumName = "Living Frame"
	imagesDir = "./static/images"
	configFile = "./config.json"
	wifiFile = "/etc/wpa_supplicant/wpa_supplicant.conf"
)


func main() {
	config := getConfigData()

	var client *photoprism.Client

	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = t
	e.Static("/static", "static")

	// Renders frame html
	e.GET("/", func(c echo.Context) error {
		fmt.Println("Loaing main page")

		config = getConfigData()

		type PageData struct {
			Images []string
			Configured bool
		}
		var pageData PageData

		images := getImagesInFolder(imagesDir)

		pageData.Images = images
		pageData.Configured = config.Configured

		// return c.JSON(http.StatusOK, pageData)
		return c.Render(http.StatusOK, "index.html", pageData)
	})

	// Syncs image folder with images in album
	e.GET("/sync", func(c echo.Context) error {
		fmt.Println("Syncing")

		if client == nil {
			client = getClient(config)
		}

		type Response struct {
			Images []string
			Changed bool
		}
		var response Response

		albumId := getAlbumId(client, albumName)
		photos := getAlbumPhotos(client, *albumId, 1000)
		response.Changed = updateImageFolder(client, photos, imagesDir)

		response.Images = getImagesInFolder(imagesDir)
		
		return c.JSON(http.StatusOK, response)
	})

	e.POST("/updateconfig", func(c echo.Context) error {
		fmt.Println("Updating config file")
		newConfig := getConfigData()

		if c.FormValue("delay") != "" {
			newConfig.Delay = c.FormValue("delay")
		}

		if c.FormValue("fade") != "" {
			newConfig.Fade = c.FormValue("fade")
		}

		config = updateConfig(newConfig)

		return c.JSON(http.StatusOK, safeConfig(newConfig))
	})

	e.GET("/getconfig", func(c echo.Context) error {
		fmt.Println("Getting config data")
		thisConfig := safeConfig(config)

		return c.JSON(http.StatusOK, thisConfig)
	})

	e.POST("/login", func(c echo.Context) error {
		fmt.Println("Logging in")
		var errors []FormError

		if c.FormValue("username") == "" {
			var err FormError
			err.Field = "username"
			err.Error = "Missing username"

			errors = append(errors, err)
		}

		if c.FormValue("password") == "" {
			var err FormError
			err.Field = "password"
			err.Error = "Missing password"

			errors = append(errors, err)
		}

		if c.FormValue("uri") == "" {
			var err FormError
			err.Field = "uri"
			err.Error = "Missing Photoprism URL"

			errors = append(errors, err)
		}

		if len(errors) > 0 {
			return c.JSON(http.StatusOK, errors)
		}

		newConfig := getConfigData()

		newConfig.Username = c.FormValue("username")
		newConfig.Password = c.FormValue("password")
		newConfig.Uri = c.FormValue("uri")

		newConfig.Configured = true
		config = updateConfig(newConfig)

		return c.JSON(http.StatusOK, safeConfig(config))
		
	})

	e.GET("/logout", func(c echo.Context) error {
		if client == nil {
			client = getClient(config)
		}

		logout(client, config)

		return c.JSON(http.StatusOK, safeConfig(config))
	})

	e.GET("/networks", func(c echo.Context) error {
		grep := exec.Command("grep", "ESSID")
		cmd := exec.Command("sudo", "iwlist", "wlan0", "scanning")

		pipe, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println(err)
		}

		grep.Stdin = pipe

		cmd.Start()

		out, err := grep.Output()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(string(out))

		return c.String(http.StatusOK, string(out))
	})

	e.POST("/wifi", func(c echo.Context) error {
		cmdStruct := exec.Command("lsblk")

		out,err := cmdStruct.Output()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(string(out))

		return c.String(http.StatusOK, string(out))
	})
	
	// e.GET("/checkconnection", func(c echo.Context) error {
		
	// })

	// e.GET("/shutdown", func(c echo.Context) error {
		// cmdStruct := exec.Command("sudo", "shutdown", "now")

		// out,err := cmdStruct.Output()
		// if err != nil {
		// 	fmt.Println(err)
		// }

		// fmt.Println(string(out))

		// return c.String(http.StatusOK, string(out))

	// })



	e.Logger.Fatal(e.Start(":1323"))
}