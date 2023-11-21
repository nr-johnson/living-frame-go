package myFunctions

func getClient(config Config) *photoprism.Client {
	var thisClient *photoprism.Client

	thisClient = photoprism.New(config.Uri)
	err := thisClient.Auth(photoprism.NewClientAuthLogin(config.Username, config.Password))
	if err != nil {
		fmt.Println(fmt.Sprintf("Login Err: %s", err))
	}

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
		Count: count,
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
	if (len(image.Files) < 1) {
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
		imageNames[(imgCount - 1) - i] = image.Name()
	}

	return imageNames
}

func updateConfig(config Config) Config {
	file, _ := json.MarshalIndent(config, "", "")
	_ = ioutil.WriteFile(configFile, file, 0644)
	fmt.Println("Config file updated.")
	return config
}
func getConfigData() Config {
	exists := fileExists(configFile)
	if (!exists) {
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

func login(config Config) *photoprism.Client {

	thisClient := getClient(config)

	err := thisClient.V1().Index()
	if err != nil {
		fmt.Println(fmt.Sprintf("Index Err: %s", err))
	}

	return thisClient
}
func logout(client *photoprism.Client, config Config) {
	newConfig := config

	err := client.V1().CancelIndex()
	if err != nil {
		fmt.Println(err)
	}

	newConfig.Configured = false

	newConfig = safeConfig(newConfig)

	updateConfig(newConfig)
}

