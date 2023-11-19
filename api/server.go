package api

import (
	"cdn/storage"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router            *gin.Engine
	s3Service         *storage.S3Service
	cloudfrontService *storage.CloudFrontService
	port              string
}

func NewServer(s3 *storage.S3Service, cloudFront *storage.CloudFrontService, port string) *Server {
	server := &Server{
		s3Service:         s3,
		cloudfrontService: cloudFront,
		port:              port,
	}
	server.SetupRouter()
	return server
}

func (s *Server) SetupRouter() {
	router := gin.New()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	router.Use(cors.New(corsConfig))
	router.MaxMultipartMemory = 8 << 20

	router.POST("/", s.HandleUploadFile)
	router.PUT("/:key", s.HandleUpdateFile)
	router.DELETE("/:key", s.HandleDeleteFile)
	router.GET("/:key", s.HandleGetFile)

	s.router = router
}

func (s *Server) Run() {
	s.router.Run(fmt.Sprintf(":%s", s.port))
}
