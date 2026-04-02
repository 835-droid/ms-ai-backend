// ----- START OF FILE: backend/MS-AI/internal/api/handler/content/manga/manga_chapter_handler.go -----
package manga

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	"github.com/835-droid/ms-ai-backend/internal/core/content/manga"
	"github.com/835-droid/ms-ai-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaChapterHandler handles chapter-related requests for manga
type MangaChapterHandler struct {
	service manga.MangaChapterService
}

// NewMangaChapterHandler creates a new manga chapter handler
func NewMangaChapterHandler(s manga.MangaChapterService) *MangaChapterHandler {
	return &MangaChapterHandler{service: s}
}

type mangaCreateChapterRequest struct {
	Title  string   `json:"title" binding:"required"`
	Images []string `json:"images" binding:"required,min=1"`
	Number float64  `json:"number"`
}

var uploadFilenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func normalizeChapterImages(images []string) []string {
	normalized := make([]string, 0, len(images))
	for _, image := range images {
		value := strings.TrimSpace(image)
		value = strings.Trim(value, "\"'")
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeChapterForResponse(chapter *manga.MangaChapter) *manga.MangaChapter {
	if chapter == nil {
		return nil
	}
	chapter.Pages = normalizeChapterImages(chapter.Pages)
	return chapter
}

func normalizeChaptersForResponse(chapters []*manga.MangaChapter) []*manga.MangaChapter {
	for i := range chapters {
		chapters[i] = normalizeChapterForResponse(chapters[i])
	}
	return chapters
}

func sanitizeUploadFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "image"
	}
	base = uploadFilenameSanitizer.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._")
	if base == "" {
		return "image"
	}
	return base
}

func chapterUploadPublicPath(mangaID primitive.ObjectID, filename string) string {
	return fmt.Sprintf("/uploads/chapters/%s/%s", mangaID.Hex(), filename)
}

func chapterUploadDiskPath(mangaID primitive.ObjectID, filename string) string {
	return filepath.Join("uploads", "chapters", mangaID.Hex(), filename)
}

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

	response.SuccessResp(c, http.StatusCreated, gin.H{
		"images": urls,
	})
}
func (h *MangaChapterHandler) ProxyImage(c *gin.Context) {

	rawURL := strings.TrimSpace(c.Query("url"))
	if rawURL == "" {
		response.ValidationError(c, "missing image url")
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		response.ValidationError(c, "invalid image url")
		return
	}

	// Create request with context
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		response.InternalError(c, "failed to prepare image request")
		return
	}

	// Set headers to mimic a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", parsedURL.Scheme+"://"+parsedURL.Host+"/")

	// Use a client with timeout
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
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   fmt.Sprintf("failed to fetch image: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"success": false,
			"error":   fmt.Sprintf("image source returned status %d", resp.StatusCode),
		})
		return
	}

	// Determine content type
	contentType := resp.Header.Get("Content-Type")
	// If missing, try to detect from extension
	if contentType == "" {
		ext := filepath.Ext(parsedURL.Path)
		contentType = mime.TypeByExtension(ext)
	}
	// Fallback to octet-stream
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Cache for 5 minutes
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Content-Type", contentType)
	// Serve the image
	c.DataFromReader(http.StatusOK, resp.ContentLength, contentType, resp.Body, nil)
}

func (h *MangaChapterHandler) CreateChapter(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	var req mangaCreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid payload")
		return
	}
	req.Images = normalizeChapterImages(req.Images)
	if len(req.Images) == 0 {
		response.ValidationError(c, "at least one valid image url is required")
		return
	}

	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	chapter := &manga.MangaChapter{
		Number:  int(req.Number),
		Title:   req.Title,
		MangaID: mangaID,
		Pages:   req.Images,
	}

	if _, err := h.service.CreateMangaChapter(c.Request.Context(), chapter, callerID, roles); err != nil {
		response.InternalError(c, "failed to create chapter")
		return
	}
	chapter = normalizeChapterForResponse(chapter)

	response.SuccessResp(c, http.StatusCreated, chapter)
}

func (h *MangaChapterHandler) ListChapters(c *gin.Context) {
	mangaIDStr := c.Param("mangaID")
	mangaID, err := primitive.ObjectIDFromHex(mangaIDStr)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	skip := int64((page - 1) * limit)
	lmt := int64(limit)

	var chapters []*manga.MangaChapter
	var total int64
	chapters, total, err = h.service.ListMangaChapters(c.Request.Context(), mangaID, skip, lmt)
	chapters = normalizeChaptersForResponse(chapters)

	response.SuccessResp(c, http.StatusOK, gin.H{
		"total":    total,
		"page":     page,
		"limit":    limit,
		"chapters": chapters,
	})
}

func (h *MangaChapterHandler) GetChapter(c *gin.Context) {
	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapterID"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	chapter, err := h.service.GetMangaChapter(c.Request.Context(), chapterID)
	if err != nil {
		response.InternalError(c, "failed to get chapter")
		return
	}
	if chapter == nil {
		response.NotFound(c, "chapter not found")
		return
	}
	chapter = normalizeChapterForResponse(chapter)

	response.SuccessResp(c, http.StatusOK, chapter)
}

func (h *MangaChapterHandler) DeleteChapter(c *gin.Context) {
	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapterID"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	if err := h.service.DeleteMangaChapter(c.Request.Context(), chapterID, callerID, roles); err != nil {
		response.InternalError(c, "failed to delete chapter")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MangaChapterHandler) UpdateChapter(c *gin.Context) {
	var req mangaCreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid payload")
		return
	}
	req.Images = normalizeChapterImages(req.Images)
	if len(req.Images) == 0 {
		response.ValidationError(c, "at least one valid image url is required")
		return
	}

	chapterID, err := primitive.ObjectIDFromHex(c.Param("chapterID"))
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	chapter, err := h.service.GetMangaChapter(c.Request.Context(), chapterID)
	if err != nil {
		response.InternalError(c, "failed to get chapter")
		return
	}
	if chapter == nil {
		response.NotFound(c, "chapter not found")
		return
	}

	chapter.Title = req.Title
	if req.Number != 0 {
		chapter.Number = int(req.Number)
	}
	chapter.Pages = req.Images

	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	if err := h.service.UpdateMangaChapter(c.Request.Context(), chapter, callerID, roles); err != nil {
		response.InternalError(c, "failed to update chapter")
		return
	}
	chapter = normalizeChapterForResponse(chapter)

	response.SuccessResp(c, http.StatusOK, chapter)
}

// ========== ENGAGEMENT METHODS ==========

// IncrementChapterViews increments view count for a chapter
func (h *MangaChapterHandler) IncrementChapterViews(c *gin.Context) {
	chapterID := c.Param("chapterID")
	mangaID := c.Param("mangaID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	mID, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	if err := h.service.IncrementChapterViews(c.Request.Context(), chID, mID); err != nil {
		response.InternalError(c, "failed to increment views")
		return
	}

	response.SuccessResp(c, http.StatusOK, gin.H{"message": "views incremented"})
}

// AddChapterRating adds a rating to a chapter
func (h *MangaChapterHandler) AddChapterRating(c *gin.Context) {
	chapterID := c.Param("chapterID")
	mangaID := c.Param("mangaID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	var req struct {
		Score float64 `json:"score" binding:"required,min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body")
		return
	}

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	mID, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	rating := &manga.ChapterRating{
		ChapterID: chID,
		MangaID:   mID,
		UserID:    userID,
		Score:     req.Score,
	}

	avgRating, err := h.service.AddChapterRating(c.Request.Context(), rating)
	if err != nil {
		response.InternalError(c, "failed to add rating")
		return
	}

	response.SuccessResp(c, http.StatusCreated, gin.H{"average_rating": avgRating})
}

// AddChapterComment adds a comment to a chapter
func (h *MangaChapterHandler) AddChapterComment(c *gin.Context) {
	chapterID := c.Param("chapterID")
	mangaID := c.Param("mangaID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}
	if mangaID == "" {
		response.ValidationError(c, "manga id required")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid request body")
		return
	}

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	mID, err := primitive.ObjectIDFromHex(mangaID)
	if err != nil {
		response.ValidationError(c, "invalid manga id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	username, _ := c.Get("username")
	usernameStr, _ := username.(string)

	comment := &manga.ChapterComment{
		ChapterID: chID,
		MangaID:   mID,
		UserID:    userID,
		Username:  usernameStr,
		Content:   req.Content,
	}

	if err := h.service.AddChapterComment(c.Request.Context(), comment); err != nil {
		response.InternalError(c, "failed to add comment")
		return
	}

	response.SuccessResp(c, http.StatusCreated, comment)
}

// ListChapterComments retrieves comments for a chapter
func (h *MangaChapterHandler) ListChapterComments(c *gin.Context) {
	chapterID := c.Param("chapterID")
	if chapterID == "" {
		response.ValidationError(c, "chapter id required")
		return
	}

	chID, err := primitive.ObjectIDFromHex(chapterID)
	if err != nil {
		response.ValidationError(c, "invalid chapter id")
		return
	}

	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	skip := int64((page - 1) * limit)
	lmt := int64(limit)

	comments, total, err := h.service.ListChapterComments(c.Request.Context(), chID, skip, lmt)
	if err != nil {
		response.InternalError(c, "failed to list comments")
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	response.SuccessResp(c, http.StatusOK, gin.H{
		"data":         comments,
		"total":        total,
		"total_pages":  totalPages,
		"current_page": page,
		"limit":        limit,
	})
}

// DeleteChapterComment deletes a chapter comment
func (h *MangaChapterHandler) DeleteChapterComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		response.ValidationError(c, "comment id required")
		return
	}

	cID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		response.ValidationError(c, "invalid comment id")
		return
	}

	userID, _, err := getCallerInfo(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.DeleteChapterComment(c.Request.Context(), cID, userID); err != nil {
		response.InternalError(c, "failed to delete comment")
		return
	}

	c.Status(http.StatusNoContent)
}

// ----- END OF FILE: backend/MS-AI/internal/api/handler/content/manga/manga_chapter_handler.go -----
