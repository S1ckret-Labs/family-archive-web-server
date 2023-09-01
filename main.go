package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	cors "github.com/rs/cors/wrapper/gin"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"github.com/S1ckret-Labs/family-archive-web-server/docs"
	"github.com/S1ckret-Labs/family-archive-web-server/features/health"
	"github.com/S1ckret-Labs/family-archive-web-server/features/tree"
	"github.com/S1ckret-Labs/family-archive-web-server/features/uploads"
	"github.com/S1ckret-Labs/family-archive-web-server/helpers"
)

var ginLambda *ginadapter.GinLambda

// @contact.name   s1ckret-labs
// @contact.url    https://github.com/S1ckret-Labs
// @contact.email  support@some_email.com

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  General Docs
// @externalDocs.url          https://github.com/S1ckret-Labs/family-archive-docs
func main() {
	docs.SwaggerInfo.Title = "Family Archive API Docs"
	docs.SwaggerInfo.Description = "Additional Info:"
	docs.SwaggerInfo.Version = "0.0.1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	config := loadConfig()
	dbConStr := config.GetString("database_connection_string")
	bucketName := config.GetString("file_uploads_bucket_name")

	db, _ := sql.Open("mysql", dbConStr)
	defer db.Close()

	awsSession := session.Must(session.NewSession())
	s3Client := s3.New(awsSession)

	uploadsFeature := uploads.Feature{Db: db, S3: s3Client, BucketName: bucketName}
	treeFeature := tree.Feature{Db: db}

	r := setupRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/api/v1/users/:id/upload/requests", uploadsFeature.GetUploadRequests)
	r.POST("/api/v1/users/:id/upload/requests", uploadsFeature.CreateUploadRequests)

	r.GET("/api/v1/users/:id/tree", treeFeature.GetTree)
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

	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173/"},
	}))

	r.GET("/health", health.GetHealth)
	return r
}

func Handler(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
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
