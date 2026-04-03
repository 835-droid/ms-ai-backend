package manga

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/835-droid/ms-ai-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *MangaChapterHandler) UploadChapterImages(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.ValidationError(c, "invalid multipart form")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		response.ValidationError(c, "no image files uploaded")
		return
	}

	dir := filepath.Join("uploads", "chapters", mangaID.Hex())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		response.InternalError(c, "failed to prepare upload directory")
		return
	}

	urls := make([]string, 0, len(files))
	for _, file := range files {
		name := sanitizeUploadFilename(file.Filename)
		finalName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), name)
		target := chapterUploadDiskPath(mangaID, finalName)
		if err := c.SaveUploadedFile(file, target); err != nil {
			response.InternalError(c, "failed to save uploaded image")
			return
		}
		urls = append(urls, chapterUploadPublicPath(mangaID, finalName))
	}

	response.SuccessResp(c, http.StatusCreated, gin.H{"images": urls})
}

func (h *MangaChapterHandler) ProxyImage(c *gin.Context) {
	rawURL := strings.TrimSpace(c.Query("url"))
	if rawURL == "" {
		response.ValidationError(c, "missing image url")
		return
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		response.ValidationError(c, "invalid image url")
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		response.InternalError(c, "failed to prepare image request")
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", parsedURL.Scheme+"://"+parsedURL.Host+"/")

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": fmt.Sprintf("failed to fetch image: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"success": false, "error": fmt.Sprintf("image source returned status %d", resp.StatusCode)})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		ext := filepath.Ext(parsedURL.Path)
		contentType = mime.TypeByExtension(ext)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Content-Type", contentType)
	c.DataFromReader(http.StatusOK, resp.ContentLength, contentType, resp.Body, nil)
}
