package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"eng_club/services"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const (
	maxUploadSize   = 10 << 20 // 10 MB
	maxDimension    = 1200     // Maximum width or height
	targetWidth     = 386
	targetHeight    = 300
	uploadDir       = "./uploads"
	collageDir      = "./uploads/collages"
)

// UploadImage handles POST /api/upload-image
func (d *Deps) UploadImage(ctx *fasthttp.RequestCtx) {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to create upload directory")
		return
	}

	// Parse multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "failed to parse form data")
		return
	}

	// Get the file from form
	files := form.File["image"]
	if len(files) == 0 {
		WriteErr(ctx, fasthttp.StatusBadRequest, "no image file provided")
		return
	}

	fileHeader := files[0]

	// Check file size
	if fileHeader.Size > maxUploadSize {
		WriteErr(ctx, fasthttp.StatusBadRequest, "file too large (max 10MB)")
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		WriteErr(ctx, fasthttp.StatusBadRequest, "only jpg, jpeg, and png files are allowed")
		return
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to open uploaded file")
		return
	}
	defer file.Close()

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to read file")
		return
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid image file")
		return
	}

	// Resize image
	resizedImg := resizeImage(img, targetWidth, targetHeight)

	// Generate unique filename
	filename := uuid.New().String() + ext
	filepath := filepath.Join(uploadDir, filename)

	// Create output file
	outFile, err := os.Create(filepath)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to create output file")
		return
	}
	defer outFile.Close()

	// Encode and save resized image
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(&buf, resizedImg)
	default:
		WriteErr(ctx, fasthttp.StatusBadRequest, "unsupported image format")
		return
	}

	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to encode image")
		return
	}

	// Check if size is reasonable (~200KB target)
	// If it's too large, reduce quality for JPEG
	imageData := buf.Bytes()
	if len(imageData) > 250*1024 && (format == "jpeg" || format == "jpg") {
		buf.Reset()
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 70})
		if err != nil {
			WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to re-encode image")
			return
		}
		imageData = buf.Bytes()
	}

	// Write to file
	_, err = outFile.Write(imageData)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to write image file")
		return
	}

	// Return success response with file info
	response := map[string]interface{}{
		"success":  true,
		"filename": filename,
		"size":     len(imageData),
		"width":    resizedImg.Bounds().Dx(),
		"height":   resizedImg.Bounds().Dy(),
	}

	WriteJSON(ctx, fasthttp.StatusOK, response)
}

// ServeImage handles GET /uploads/{filename}
func (d *Deps) ServeImage(ctx *fasthttp.RequestCtx) {
	filename := ctx.UserValue("filename")
	if filename == nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "missing filename")
		return
	}

	// Sanitize filename to prevent directory traversal
	filenameStr := fmt.Sprint(filename)
	if strings.Contains(filenameStr, "..") || strings.Contains(filenameStr, "/") {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid filename")
		return
	}

	filepath := filepath.Join(uploadDir, filenameStr)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		WriteErr(ctx, fasthttp.StatusNotFound, "image not found")
		return
	}

	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to read image")
		return
	}

	// Set content type based on extension
	ext := strings.ToLower(filepath[len(filepath)-4:])
	switch ext {
	case ".jpg", "jpeg":
		ctx.Response.Header.SetContentType("image/jpeg")
	case ".png":
		ctx.Response.Header.SetContentType("image/png")
	default:
		ctx.Response.Header.SetContentType("application/octet-stream")
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(data)
}

// ServeCollage handles GET /uploads/collages/{filename}
func (d *Deps) ServeCollage(ctx *fasthttp.RequestCtx) {
	filename := ctx.UserValue("filename")
	if filename == nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "missing filename")
		return
	}

	// Sanitize filename to prevent directory traversal
	filenameStr := fmt.Sprint(filename)
	if strings.Contains(filenameStr, "..") || strings.Contains(filenameStr, "/") {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid filename")
		return
	}

	filepath := filepath.Join(collageDir, filenameStr)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		WriteErr(ctx, fasthttp.StatusNotFound, "collage not found")
		return
	}

	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, "failed to read collage")
		return
	}

	// Set content type for JPEG (all collages are saved as JPEG)
	ctx.Response.Header.SetContentType("image/jpeg")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(data)
}

// resizeImage resizes an image to fit within the target dimensions while maintaining aspect ratio
// If either dimension exceeds maxDimension (1200px), it scales down proportionally
func resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// First, check if image exceeds max dimension and scale down if needed
	maxSrcDimension := srcWidth
	if srcHeight > maxSrcDimension {
		maxSrcDimension = srcHeight
	}

	var scale float64
	if maxSrcDimension > maxDimension {
		// Scale down to max dimension
		scale = float64(maxDimension) / float64(maxSrcDimension)
	} else {
		// Calculate scaling factor to fit within target dimensions
		scaleX := float64(targetWidth) / float64(srcWidth)
		scaleY := float64(targetHeight) / float64(srcHeight)
		scale = scaleX
		if scaleY < scaleX {
			scale = scaleY
		}

		// If image is already smaller, don't upscale
		if scale > 1.0 {
			return img
		}
	}

	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// Create new image with target dimensions
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor resize
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}

	return dst
}

// CreateCollage creates a collage from multiple image filenames
// Returns the collage filename or error. Uses checksum to avoid recreating existing collages.
func (d *Deps) CreateCollage(imageFilenames []string) (string, error) {
	return services.CreateCollage(imageFilenames)
}
