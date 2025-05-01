package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
    const maxMemory = 10 << 20
    r.ParseMultipartForm(maxMemory)

    fileData, fileHeader, err := r.FormFile("thumbnail")
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
        return
    }
    defer fileData.Close()
    mediaType := fileHeader.Header.Get("Content-Type")

    thumbnailData, err := io.ReadAll(fileData)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to parse image", err)
        return
    }

    video, err := cfg.db.GetVideo(videoID)
    if video.UserID != userID {
        respondWithError(w, http.StatusUnauthorized, "User missmatch", err)
        return
    }

    mediaThumbnail := thumbnail{
        data: thumbnailData,
        mediaType: mediaType,
    }

    videoThumbnails[video.ID] = mediaThumbnail

    thumbnailURL := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v", cfg.port, video.ID)

    video.ThumbnailURL = &thumbnailURL
    if video.ThumbnailURL == nil {
        respondWithError(w, http.StatusInternalServerError, "Thumbnail url missing", nil)
        return
    }

    err = cfg.db.UpdateVideo(video)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to update db", err)
        return
    }

	respondWithJSON(w, http.StatusOK, video)
}
