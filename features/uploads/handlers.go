package uploads

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4"

	"github.com/S1ckret-Labs/family-archive-web-server/helpers"
)

// feature represents a feature with a database connection S3 | client and bucket name
type Feature struct {
	Db         *sql.DB
	S3         *s3.S3
	BucketName string
}

// CreateUploadRequest represents the request structure for creating an upload
type CreateUploadRequest struct {
	ObjectKey  string   `json:"ObjectKey"  example:"example_object_key"`
	SizeBytes  uint64   `json:"SizeBytes"  example:"102400"`
	TakenAtSec null.Int `json:"TakenAtSec"`
}

// CreateUploadRequestResult represents the result structure after creating an upload request
type CreateUploadRequestResult struct {
	ObjectId  uint64 `json:"ObjectId"  example:"123"`
	ObjectKey string `json:"ObjectKey" example:"example_object_key"`
	UploadUrl string `json:"UploadUrl" example:"https://s3.example.com/uploads/uploaded_file"`
}

// @Summary Get user's upload requests
// @Description Get the upload requests for a user
// @ID get-upload-requests
// @Produce json
// @Param id path uint64 true "User ID"
// @Success 200 "Successful response"
// @Failure 400 "Bad Request"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/users/{id}/upload/requests [get]
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

// @Summary Create upload requests
// @Description Create upload requests for a user
// @ID create-upload-requests
// @Accept json
// @Produce json
// @Param id path uint64 true "User ID"
// @Success 200 "Successful response"
// @Failure 400 "Bad Request"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/users/{id}/upload/requests [post]
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

// @Summary Delete user's upload requests
// @Description Delete the upload requests for a user
// @ID delete-upload-requests
// @Produce json
// @Param id path uint64 true "User ID"
// @Success 204 "No Content"
// @Failure 400 "Bad Request"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/users/{id}/upload/requests [delete]
func (f Feature) DeleteUploadRequests(c *gin.Context) {
	userId, err := helpers.ParamUint64(c, "id")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = DeleteUploadRequests(f.Db, userId)
	if err != nil {
		log.Println("Error while deleting upload requests from the database!")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func composeCreateUploadRequestResults(
	ids []uint64,
	keys []string,
	urls []string,
) []CreateUploadRequestResult {
	if len(ids) != len(keys) || len(keys) != len(urls) {
		log.Panicf(
			"Can't compose final result. Array sizes doesn't match! %d, %d, %d\n",
			len(ids),
			len(keys),
			len(urls),
		)
	}

	var results []CreateUploadRequestResult
	for i := range ids {
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
