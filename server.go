package main

import (
	"html/template"
	"net/http"
	"os/exec"
	"time"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"

	photoprism "github.com/drummonds/photoprism-client-go"
	"github.com/drummonds/photoprism-client-go/api/v1"

	wifi "github.com/mark2b/wpa-connect"

)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Config struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	Logged_in        bool   `json:"logged_in"`
	Network          string `json:"network"`
	Network_password string `json:"network_password"`
	Connected        bool   `json:"connected"`
	Uri              string `json:"uri"`
	Delay            string `json:"delay"`
	Fade             string `json:"fade"`
	Configured       bool   `json:"configured"`
	Status      	 int   	`json:"status"`
	Button      	 int   	`json:"button"`
}

type FormError struct {
	Field string
	Error string
}

var (
	albumName  = "Living Frame"
	imagesDir  = "./static/images"
	configFile = "./config.json"
	wifiFile   = "/etc/wpa_supplicant/wpa_supplicant.conf"
)

func main() {
	config := getConfigData()
	config.Status = 1
	updateConfig(config)

	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = t
	e.Static("/static", "static")

	// wifiConnected := checkConnection()
	// if wifiConnected == true {
	// 	fmt.Println("Wifi connected")
	// 	config.Connected = true
	// 	updateConfig(config)
	// } else {
	// 	fmt.Println("No Wifi connection!")
	// 	config.Connected = false
	// 	updateConfig(config)
	// }

	// Renders frame html
	e.GET("/", func(c echo.Context) error {
		fmt.Println("Loaing main page")

		config = getConfigData()

		type PageData struct {
			Images     []string
			Configured bool
		}
		var pageData PageData

		images := getImagesInFolder(imagesDir)

		pageData.Images = images
		pageData.Configured = config.Configured

		config.Status = 2
		updateConfig(config)

		// return c.JSON(http.StatusOK, pageData)
		return c.Render(http.StatusOK, "index.html", pageData)
	})

	// Syncs image folder with images in album
	e.GET("/sync", func(c echo.Context) error {
		fmt.Println("Syncing")
		type Response struct {
			Images  []string
			Changed bool
		}
		var response Response

		client := getClient(config)

		albumId := getAlbumId(client, albumName)
		photos := getAlbumPhotos(client, *albumId, 1000)
		response.Changed = updateImageFolder(client, photos, imagesDir)

		response.Images = getImagesInFolder(imagesDir)

		logout(client, config)

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

		fmt.Println("All fields included")

		getClient(config)
		newConfig.Configured = true
		newConfig.Logged_in = true	

		config = updateConfig(newConfig)
		
		fmt.Println("Logged in!")

		// logout(client, config)

		return c.JSON(http.StatusOK, safeConfig(config))

	})

	e.POST("/logout", func(c echo.Context) error {
		fmt.Println("Logging out")

		client := getClient(config)

		logout(client, config)

		config.Logged_in = false

		updateConfig(config)

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
		ssid := c.FormValue("network")
		password := c.FormValue("password")
		
		err := connectToWifi(config, ssid, password)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusOK, err)
		}

		return c.JSON(http.StatusOK, safeConfig(config))
	})

	e.Logger.Fatal(e.Start(":1323"))
}

// Photoprism functions
func login(config Config) (*photoprism.Client, error) {
	var respErr error

	thisClient := getClient(config)

	fmt.Println("Got client")

	return thisClient, respErr
}
func logout(client *photoprism.Client, config Config) {
	err := client.V1().CancelIndex()
	if err != nil {
		fmt.Println(err)
	}
}
func getClient(config Config) *photoprism.Client {
	var thisClient *photoprism.Client

	thisClient = photoprism.New(config.Uri)
	err := thisClient.Auth(photoprism.NewClientAuthLogin(config.Username, config.Password))
	if err != nil {
		fmt.Println(fmt.Sprintf("Login Err: %s", err))
	}

	fmt.Println("Got client!")

	// err = thisClient.V1().Index()
	// if err != nil {
	// 	fmt.Println(fmt.Sprintf("Index Err: %s", err))
	// }

	return thisClient
}
func getAlbumId(client *photoprism.Client, target string) *string {
	albums, err := client.V1().GetAlbums(nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	album := "nil"
	for _, elem := range albums {
		if elem.AlbumTitle == target {
			album = elem.AlbumUID
			break
		}
	}
	if album == "nil" {
		return nil
	}

	return &album
}
func getAlbumPhotos(client *photoprism.Client, target string, count int) []api.Photo {
	photos, err := client.V1().GetPhotos(&api.PhotoOptions{
		Count:    count,
		AlbumUID: target,
	})
	if err != nil {
		fmt.Println(err)
	}
	// Ensure photo contains files
	for i, photo := range photos {
		if len(photo.Files) < 1 {
			photos[i] = getPhoto(client, photo.PhotoUID)
		}
	}

	return photos
}
func getPhoto(client *photoprism.Client, target string) api.Photo {
	photo, err := client.V1().GetPhoto(target)
	if err != nil {
		fmt.Println(err)
	}

	return photo
}
func updateImageFolder(client *photoprism.Client, photos []api.Photo, dir string) bool {
	fmt.Println("Updating images...")

	images, _ := ioutil.ReadDir(dir)
	deleted := false
	downloaded := false

	// Downloads images not already in folder
	for _, photo := range photos {

		photoName := photo.Files[0].FileName[strings.LastIndex(photo.Files[0].FileName, "/")+1:]
		found := false
		for _, img := range images {
			fileName := img.Name()

			if fileName == photoName {
				found = true
				break
			}
		}
		if found != true {
			downloadPhoto(client, photo, dir)
			downloaded = true
		}
	}
	if downloaded != true {
		fmt.Println("No images downloaded.")
	}

	// Remove images not in list from photoprism
	for _, img := range images {
		if !img.IsDir() {
			fileName := img.Name()
			found := false

			for _, photo := range photos {
				imgName := photo.Files[0].FileName[strings.LastIndex(photo.Files[0].FileName, "/")+1:]
				if fileName == imgName {
					found = true
					break
				}
			}

			if found != true {
				deleted = true
				os.Remove(fmt.Sprintf("%s/%s", dir, fileName))
				fmt.Println(fmt.Sprintf("Deleted %s", fileName))
			}
		}
	}
	if deleted != true {
		fmt.Println("No images deleted.")
	}

	fmt.Println("Images updated!")

	if downloaded == true || deleted == true {
		return true
	}
	return false
}
func downloadPhoto(client *photoprism.Client, image api.Photo, dir string) bool {
	// Ensures image object contains array of files
	if len(image.Files) < 1 {
		properImage := getPhoto(client, image.PhotoUID)
		return downloadPhoto(client, properImage, dir)
	}

	file, err := client.V1().GetPhotoDownload(image.Files[0].PhotoUID)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fileName := fmt.Sprintf("%s/%s", dir, path.Base(image.Files[0].FileName))
	ioutil.WriteFile(fileName, file, 0666)
	fmt.Println(fmt.Sprintf("Downloaded %s", fileName))

	return true
}

// Returns array of names from images in /static/images folder.
func getImagesInFolder(dir string) []string {
	images, _ := ioutil.ReadDir(dir)
	imgCount := len(images)
	imageNames := make([]string, imgCount)

	for i, image := range images {
		imageNames[(imgCount-1)-i] = image.Name()
	}

	return imageNames
}

// Config file functions
func updateConfig(config Config) Config {
	file, _ := json.MarshalIndent(config, "", "")
	_ = ioutil.WriteFile(configFile, file, 0644)
	fmt.Println("Config file updated.")
	return config
}
func getConfigData() Config {
	exists := fileExists(configFile)
	if !exists {
		return generateNewConfigFile()
	}

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error opening config file: %s", err))
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error reading config file: %s", err))
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error parsing config file: %s", err))
	}

	return config
}
func generateNewConfigFile() Config {
	var config Config
	config.Logged_in = false
	config.Connected = false
	config.Configured = false
	config.Delay = "10"
	config.Fade = "2"

	updateConfig(config)

	return config
}
func fileExists(path string) bool {
	_, error := os.Stat(path)
	return !errors.Is(error, os.ErrNotExist)
}
func safeConfig(config Config) Config {
	safeData := config

	safeData.Password = ""
	safeData.Network_password = ""

	return safeData
}

// Wifi functions
func connectToWifi(config Config, ssid string, password string) error {
	var conErr error

	if conn, err := wifi.ConnectManager.Connect(ssid, password, time.Second * 60); err == nil {
		fmt.Println("Connected", conn.NetInterface, conn.SSID, conn.IP4.String(), conn.IP6.String())
		config.Network = ssid
		config.Network_password = password
		config.Connected = true
		updateConfig(config)
	} else {
		conErr = err
	}
	return conErr
}
func checkConnection() bool {
	_, err := http.Get("http://clients3.google.com/generate_204")
    if err != nil {
        return false
    }
    return true
}