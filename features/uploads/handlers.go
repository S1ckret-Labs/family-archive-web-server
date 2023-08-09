package uploads

import (
	"database/sql"
	"github.com/S1ckret-Labs/family-archive-web-server/helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4"
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

type CreateUploadRequest struct {
	ObjectKey  string
	SizeBytes  uint64
	TakenAtSec null.Int
}

type CreateUploadRequestResult struct {
	ObjectId  uint64
	ObjectKey string
	UploadUrl string
}

func (f Feature) GetUploadRequests(c *gin.Context) {
	userId, err := helpers.ParamUint64(c, "id")
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
	userId, err := helpers.ParamUint64(c, "id")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// TODO: Validate if file names are not colliding (must be unique)
	// TODO: Validate file extensions (must be acceptable)
	var uploadRequests []CreateUploadRequest
	if err := c.BindJSON(&uploadRequests); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var objectKeys []string
	for _, request := range uploadRequests {
		objectKeys = append(objectKeys, request.ObjectKey)
	}

	urls, err := f.createS3SignedUrls(strconv.FormatUint(userId, 10)+"/", objectKeys)
	if err != nil {
		log.Println("Error while creating S3 signed URLs!")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ids, err := InsertUploadRequests(f.Db, userId, uploadRequests)
	if err != nil {
		log.Println("Error while inserting upload files to a database!")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	results := composeCreateUploadRequestResults(ids, objectKeys, urls)
	c.JSON(http.StatusOK, results)
}

func composeCreateUploadRequestResults(ids []uint64, keys []string, urls []string) []CreateUploadRequestResult {
	if len(ids) != len(keys) || len(keys) != len(urls) {
		log.Panicf("Can't compose final result. Array sizes doesn't match! %d, %d, %d\n", len(ids), len(keys), len(urls))
	}

	var results []CreateUploadRequestResult
	for i, _ := range ids {
		results = append(results, CreateUploadRequestResult{
			ObjectId:  ids[i],
			ObjectKey: keys[i],
			UploadUrl: urls[i],
		})
	}
	return results
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
