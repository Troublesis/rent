package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/service"
)

const (
	maxUploadSize        = 10 << 20
	maxUploadRequestSize = maxUploadSize + 1<<20
)

type UploadHandler struct {
	uploadDir   string
	roomService *service.RoomService
}

func NewUploadHandler(uploadDir string, roomService *service.RoomService) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir, roomService: roomService}
}

func (h *UploadHandler) UploadRoomMedia(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadRequestSize)
	roomID, err := parseUintForm(c, "room_id")
	if err != nil {
		redirectWithError(c, "/admin/rooms", "请选择房源")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		redirectWithError(c, adminRoomEditURL(roomID), "请选择上传文件")
		return
	}
	if file.Size > maxUploadSize {
		redirectWithError(c, adminRoomEditURL(roomID), "文件不能超过 10MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	mediaType, ok := mediaTypeFromExt(ext)
	if !ok {
		redirectWithError(c, adminRoomEditURL(roomID), "仅支持 jpg、png、mp4 文件")
		return
	}

	dir := filepath.Join(h.uploadDir, fmt.Sprintf("%d", roomID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		redirectWithError(c, adminRoomEditURL(roomID), "创建上传目录失败")
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(file, path); err != nil {
		redirectWithError(c, adminRoomEditURL(roomID), "保存文件失败")
		return
	}

	url := fmt.Sprintf("/uploads/%d/%s", roomID, filename)
	if err := h.roomService.AddRoomMedia(roomID, url, mediaType); err != nil {
		_ = os.Remove(path)
		redirectWithError(c, adminRoomEditURL(roomID), "保存媒体记录失败")
		return
	}
	c.Redirect(http.StatusSeeOther, adminRoomEditURL(roomID))
}

func mediaTypeFromExt(ext string) (string, bool) {
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return model.MediaTypeImage, true
	case ".mp4":
		return model.MediaTypeVideo, true
	default:
		return "", false
	}
}
