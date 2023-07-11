package main

import (
	"context"
	"database/sql"
	"github.com/S1ckret-Labs/family-archive-web-server/features/health"
	"github.com/S1ckret-Labs/family-archive-web-server/features/uploads"
	"github.com/S1ckret-Labs/family-archive-web-server/helpers"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"os"
)

var ginLambda *ginadapter.GinLambda

func main() {
	config := loadConfig()
	dbConStr := config.GetString("database_connection_string")
	bucketName := config.GetString("file_uploads_bucket_name")

	db, _ := sql.Open("mysql", dbConStr)
	defer db.Close()

	awsSession := session.Must(session.NewSession())
	s3Client := s3.New(awsSession)

	uploadsFeature := uploads.Feature{Db: db, S3: s3Client, BucketName: bucketName}

	r := setupRouter()
	r.GET("/api/v1/users/:id/upload/requests", uploadsFeature.GetUploadRequests)
	r.POST("/api/v1/users/:id/upload/requests", uploadsFeature.CreateUploadRequests)

	Run(r)
}

func loadConfig() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix("FA")
	v.AutomaticEnv()
	return v
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(helpers.ErrorHandler())
	r.GET("/health", health.GetHealth)
	return r
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, request)
}

func Run(r *gin.Engine) {
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		ginLambda = ginadapter.New(r)
		lambda.Start(Handler)
	} else {
		r.Run(":8080")
	}
}
