package uploads

import (
	"database/sql"
	"github.com/S1ckret-Labs/family-archive-web-server/helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Feature struct {
	Db         *sql.DB
	S3         *s3.S3
	BucketName string
}

func (f Feature) GetUploadRequests(c *gin.Context) {
	userId, err := helpers.ParamInt64(c, "id")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	uploadFiles, err := FindUploadRequests(f.Db, userId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, uploadFiles)
}

func (f Feature) CreateUploadRequests(c *gin.Context) {
	userId, err := helpers.ParamInt64(c, "id")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// TODO: validate file names
	var fileNames []string
	if err := c.BindJSON(&fileNames); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	urls, err := f.createS3SignedUrls(strconv.FormatInt(userId, 10)+"/", fileNames)
	if err != nil {
		log.Println("Error while creating S3 signed URLs!")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ids, err := InsertUploadRequests(f.Db, userId, fileNames)
	if err != nil {
		log.Println("Error while inserting upload files to a database!")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Println("Inserted UploadFiles: ", ids)

	filesToUploadUrls := helpers.ZipToMap(fileNames, urls)

	// Result
	c.JSON(http.StatusOK, filesToUploadUrls)
}

func (f Feature) createS3SignedUrls(prefix string, fileNames []string) ([]string, error) {
	urls := make([]string, 0, 20) // 20 is minimum user upload size

	for _, fileName := range fileNames {
		putObjectInput := &s3.PutObjectInput{
			Bucket: aws.String(f.BucketName),
			Key:    aws.String(prefix + fileName),
		}
		req, _ := f.S3.PutObjectRequest(putObjectInput)
		url, err := req.Presign(10 * time.Minute)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}
