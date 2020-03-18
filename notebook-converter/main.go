package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
)

type Request struct {
	BucketName string `json:"bucketName"`
}

func main() {
	bucketName := os.Args[1]
	objectKey := os.Args[2]

	if bucketName == "" {
		panic("bucket name is empty")
	}
	if objectKey == "" {
		panic("object key is empty")
	}

	endpoint := "minio-service.argo-events.svc:9000"
	accessKeyID := "minio"
	secretAccessKey := "minio123"

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, false)
	if err != nil {
		panic(err)
	}

	err = minioClient.FGetObject(bucketName, objectKey, "out/noisy.ipynb", minio.GetObjectOptions{})
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.POST("/", func(context *gin.Context) {
		var request Request
		if err := context.ShouldBindJSON(&request); err != nil {
			context.JSON(http.StatusInternalServerError, "failed to parse request")
			return
		}

		cmd := exec.Command("papermill", "out/noisy.ipynb", "out/out.ipynb", "-p", "bucketName", request.BucketName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("err: %+v\n", err)
			context.JSON(http.StatusInternalServerError, "failed run the notebook")
			return
		}

		context.JSON(http.StatusOK, fmt.Sprintf("successfully stored output image from %s bucket", request.BucketName))
	})

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
