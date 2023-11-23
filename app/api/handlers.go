package api

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (server *Server) HandleUploadFile(ctx *gin.Context) {

	file, err := ctx.FormFile("file")
	if err != nil {
		log.Error().Err(err).Msg("failed to get file")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "File with key 'file' not found."})
		return
	}

	openFile, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload file."})
		return
	}

	defer openFile.Close()

	contentType := getContentType(openFile)

	err = server.s3Service.UploadFile(file.Filename, contentType, openFile)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload file")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload file"})
		return
	}

	log.Info().Str("key", file.Filename).Msg("file uploaded succesfully")
	ctx.JSON(http.StatusCreated, gin.H{"message": "File uploaded succesfully."})
}

func (server *Server) HandleUpdateFile(ctx *gin.Context) {

	file, err := ctx.FormFile("file")
	if err != nil {
		log.Error().Err(err).Msg("failed to get file")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "File with key 'file' not found."})
		return
	}

	key := ctx.Param("key")
	if !server.s3Service.DoesKeyExist(key) {
		log.Warn().Str("key", key).Str("bucket", server.s3Service.Bucket).Msg("given key not exist in bucket")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Given key not exist in bucket"})
		return
	}

	openFile, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update file."})
		return
	}

	defer openFile.Close()

	contentType := getContentType(openFile)

	err = server.s3Service.UploadFile(key, contentType, openFile)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload file")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update file"})
		return
	}

	log.Info().Str("key", file.Filename).Msg("file updated succesfully")

	err = server.cloudfrontService.CreateInvalidation(fmt.Sprintf("/%s", key))
	if err != nil {
		log.Error().Err(err).Msg("failed to create invalidation")
	}

	ctx.Status(http.StatusNoContent)
}

func (server *Server) HandleDeleteFile(ctx *gin.Context) {

	key := ctx.Param("key")

	err := server.s3Service.DeleteFile(key)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete file")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update file"})
		return
	}

	err = server.cloudfrontService.CreateInvalidation(fmt.Sprintf("/%s", key))
	if err != nil {
		log.Error().Err(err).Msg("failed to create invalidation")
	}

	log.Info().Str("key", key).Msg("file deleted succesfully")
	ctx.Status(http.StatusNoContent)
}

func (server *Server) HandleGetFile(ctx *gin.Context) {

	key := ctx.Param("key")

	reader, contentLength, contentType, err := server.cloudfrontService.FetchFile(key)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to fetch file")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch file",
		})
		return
	}

	log.Info().Str("key", key).Msg("file fetched succesfully")

	extraHeaders := map[string]string{
		"Content-Disposition": "inline",
	}

	ctx.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}

func getContentType(file multipart.File) string {

	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		log.Error().Err(err).Msg("failed to read file")
		return "application/octet-stream"
	}

	return http.DetectContentType(buffer)
}
