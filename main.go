package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	dockerapi "github.com/oribon/malbinon/pkg/dockerApi"
	dogsapi "github.com/oribon/malbinon/pkg/dogsApi"
)

type FileInfo struct {
	os.FileInfo
}

func (f FileInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Name":    f.Name(),
		"Size":    f.Size(),
		"Mode":    f.Mode(),
		"ModTime": f.ModTime().Format("2-Jan-06 15:04:05"),
	})
}

type ByTime []FileInfo

func (f ByTime) Len() int { return len(f) }
func (f ByTime) Less(i, j int) bool {
	return (f[i].ModTime()).String() > (f[j].ModTime()).String()
}
func (f ByTime) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type FileInfoList struct {
	fileInfoList []os.FileInfo
}

type DirInfoList struct {
	dirInfoList []os.FileInfo
}

type ImageList struct {
	Images []string `json:"images"`
}

func (f FileInfoList) MarshalJSON() ([]byte, error) {
	fileInfoList := make([]FileInfo, 0, len(f.fileInfoList))
	for _, val := range f.fileInfoList {
		if !val.IsDir() {
			fileInfoList = append(fileInfoList, FileInfo{val})
		}
	}

	return json.Marshal(fileInfoList)
}

func (f DirInfoList) MarshalJSON() ([]byte, error) {
	dirInfoList := make([]FileInfo, 0, len(f.dirInfoList))
	for _, val := range f.dirInfoList {
		if val.IsDir() {
			dirInfoList = append(dirInfoList, FileInfo{val})
		}
	}
	sort.Sort(ByTime(dirInfoList))
	return json.Marshal(dirInfoList)
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("./public/html/*.html")
	r.Static("/public/", "./public/")
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})

	r.GET("/images", func(c *gin.Context) {
		c.HTML(200, "images.html", gin.H{})
	})

	r.GET("/dirs", func(c *gin.Context) {
		c.HTML(200, "dirs.html", gin.H{})
	})

	r.GET("/dirs/:dirname", func(c *gin.Context) {
		c.HTML(200, "dirImages.html", gin.H{})
	})

	r.GET("/api/dirs", func(c *gin.Context) {
		dirsPath := "/dirs"
		files, err := ioutil.ReadDir(dirsPath)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(200, DirInfoList{files})
	})

	r.GET("/api/images", func(c *gin.Context) {
		imagesPath := "/images"
		files, err := ioutil.ReadDir(imagesPath)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(200, FileInfoList{files})
	})

	r.GET("/api/dirs/:dirname", func(c *gin.Context) {
		dirsPath := "/dirs"
		dirName := c.Param("dirname")
		dirsPath = filepath.Join(dirsPath, dirName)
		files, err := ioutil.ReadDir(dirsPath)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(200, FileInfoList{files})
	})

	r.GET("/download/:filename", func(c *gin.Context) {
		imagesPath := "/images"
		fileName := c.Param("filename")
		targetPath := filepath.Join(imagesPath, fileName)
		//This check is for example, I am not sure it can prevent all possible filename attacks - will be much better if real filename will not come from user side.
		if !strings.HasPrefix(filepath.Clean(targetPath), imagesPath) {
			c.String(403, "Looks like you are attacking me")
			return
		}
		//Seems this headers needed for some browsers (for example without this headers Chrome will download files as txt)
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		c.File(targetPath)
	})

	r.POST("/imageslist", func(c *gin.Context) {
		var images ImageList
		err := c.BindJSON(&images)
		if err != nil {
			fmt.Println(err)
		}
		session := sessions.Default(c)
		session.Set("images", images.Images)
		session.Save()
		c.JSON(200, gin.H{"images": session.Get("images")})
	})

	r.GET("/ws", func(c *gin.Context) {
		wshandler(c)
	})

	r.GET("/malbinon", func(c *gin.Context) {
		c.HTML(200, "malbinon.html", gin.H{})
	})

	r.RunTLS(":443", "/etc/crts/cert.pem", "/etc/crts/privkey.pem")
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wshandler(c *gin.Context) {
	sess := sessions.Default(c)
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade:", err)
		return
	}

	images, ok := sess.Get("images").([]string)
	if !ok {
		conn.WriteMessage(websocket.TextMessage, []byte("No images were previously sent, aborting..."))
		return
	}

	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}
	msgs := make(chan string)
	imagesFilesChan := make(chan string)
	var imagesFileNames []string

	for _, image := range images {
		wg.Add(1)
		go pullAndSaveImage(image, imagesFilesChan, msgs, &wg)
	}

	done := make(chan bool)

	go func() {
		wg.Wait()
		done <- true
		close(done)
	}()

readChannel:
	for {
		select {
		case msg := <-msgs:
			sendStringToWebSocket(msg, mutex, conn)
		case image := <-imagesFilesChan:
			imagesFileNames = append(imagesFileNames, image)
		case <-done:
			close(msgs)
			close(imagesFilesChan)
			for msg := range msgs {
				sendStringToWebSocket(msg, mutex, conn)
			}
			for image := range imagesFilesChan {
				imagesFileNames = append(imagesFileNames, image)
			}
			break readChannel
		}
	}

	if len(imagesFileNames) > 0 {
		dirName, err := createDogDir("/dirs")
		if err != nil {
			sendStringToWebSocket("err: "+err.Error(), mutex, conn)
		} else {
			for _, image := range imagesFileNames {
				err = os.Symlink(filepath.Join("/images", image), filepath.Join("/dirs", dirName, image))
			}
			sendStringToWebSocket("Your dog dir name is: "+dirName, mutex, conn)
		}
	} else {
		sendStringToWebSocket("err: No images were pulled, no dog for you :(", mutex, conn)
	}

	sendStringToWebSocket("Malbinon ended! Have a lovely day!", mutex, conn)
	conn.Close()
}

func pullAndSaveImage(imageName string, imagesFiles chan string, msgs chan string, wg *sync.WaitGroup) error {
	defer wg.Done()
	msgs <- "Started pulling: " + imageName
	imagePulled, err := dockerapi.PullImage(imageName)
	if err != nil {
		msgs <- "err: " + err.Error()
		return err
	}

	msgs <- "Sucessfully pulled: " + imagePulled

	imageFileName, err := dockerapi.SaveImage(imagePulled, "/images")
	if err != nil {
		msgs <- "err: " + err.Error()
		return err
	}

	msgs <- "Sucessfully saved: " + imagePulled + " as: " + imageFileName

	imagesFiles <- imageFileName

	return nil
}

func sendStringToWebSocket(str string, mutex *sync.Mutex, conn *websocket.Conn) {
	mutex.Lock()
	conn.WriteMessage(websocket.TextMessage, []byte(str))
	mutex.Unlock()
}

func createDogDir(baseDir string) (string, error) {
	dirName, err := dogsapi.GenerateDogName()
	if err != nil {
		return "", err
	}
	err = os.Mkdir(filepath.Join(baseDir, dirName), os.ModePerm)
	if err != nil {
		return "", err
	}
	return dirName, nil
}
