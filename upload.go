package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

func FileUploadRoute(c * gin.Context) {
	fmt.Println("Here")

	shasum := c.PostForm("shasum")
	/*channelID := c.PostForm("channelID")
	size := c.PostForm("size")
	name := c.PostForm("name")
	extension := c.PostForm("extension")
	author := c.PostForm("author")*/

	file, err := c.FormFile("file")

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	filename := "./files/" + filepath.Base(shasum)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("File %s uploaded", file.Filename))
}

func ServeExecutable(c * gin.Context) {
	http.ServeFile(c.Writer, c.Request, "./SecureChat Setup 0.5.8.exe")
}
