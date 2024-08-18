package middleware

import (
	"be/config"
	"be/utils"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
		})
		return
	}

	// connection firebasse storage
	storageClient, err := config.ConfigFirebaseStorage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	folder, _ := utils.CheckFolder(file.Header.Get("Content-Type"))
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	objectName := path.Join(folder, fileName)

	// buckett
	bucket := storageClient.Bucket(os.Getenv("URL_BUCKET_FIREBASE"))
	obj := bucket.Object(objectName)
	wc := obj.NewWriter(context.Background())
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer f.Close()
	if _, err := io.Copy(wc, f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := wc.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
    valiToken,err := GetIdAuthorFromToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	fileURL := strings.ReplaceAll(folder, "/", "%2F")
	fileLink :=  os.Getenv("URL_BUCKET_FIREBASE") + "/o/" + fileURL + fileName + "?alt=media"
	c.Set("file_link", fileLink)
	c.Set("user_id", valiToken)
	c.Next()
}
// folder, typeFile := utils.CheckFolder(file.Header.Get("Content-Type"))

// c.JSON(http.StatusOK, gin.H{
// 	"imageURL": imageURL,
// })
// fileName := fmt.Sprintf("%s_%d.%s", file.Filename, time.Now().UnixNano(), typeFile)
// if err := c.SaveUploadedFile(file, folder+fileName); err != nil {
// 	c.JSON(http.StatusInternalServerError, gin.H{
// 		"err":    err.Error(),
// 		"msg":    "Can't save uploaded file " + fileName,
// 		"status": http.StatusInternalServerError,
// 	})
// 	c.Abort()
// 	return
// }
// c.JSON(201, gin.H{
// 	// "file": "FIle : " + fileName,
// 	"folder":folder,
// 	"data": folder + fileName,
// 	"req":file.Header.Get("Content-Type"),
// })
// c.Set("file_link", folder + fileName )
// c.Next()
