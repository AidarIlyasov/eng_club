package telegram

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

const (
	collageDir      = "./uploads/collages"
	uploadDir       = "./uploads"
	collageWidth    = 600  // Total width of collage
	collageHeight   = 400  // Total height of collage
	thumbnailWidth  = 200  // Width of each thumbnail in collage
	thumbnailHeight = 133  // Height of each thumbnail in collage
)

// CreateCollage creates a collage from multiple image filenames
// Returns the collage filename or error. Uses checksum to avoid recreating existing collages.
func CreateCollage(imageFilenames []string) (string, error) {
	fmt.Printf("DEBUG: Creating collage with filenames: %v\n", imageFilenames)
	
	// Ensure collage directory exists
	if err := os.MkdirAll(collageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create collage directory: %v", err)
	}

	// Generate checksum from filenames in original order (order matters for events)
	hash := md5.New()
	for _, filename := range imageFilenames {
		hash.Write([]byte(filename))
	}
	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	collageFilename := fmt.Sprintf("collage_%s.jpg", checksum)
	collagePath := filepath.Join(collageDir, collageFilename)
	
	fmt.Printf("DEBUG: Collage filename will be: %s\n", collageFilename)

	// Check if collage already exists
	if _, err := os.Stat(collagePath); err == nil {
		fmt.Printf("DEBUG: Using existing collage: %s\n", collageFilename)
		return collageFilename, nil
	}

	// Load images
	images := make([]image.Image, 0, len(imageFilenames))
	for i, filename := range imageFilenames {
		if filename == "" {
			fmt.Printf("DEBUG: Skipping empty filename at index %d\n", i)
			continue // Skip empty filenames
		}
		
		imagePath := filepath.Join(uploadDir, filename)
		fmt.Printf("DEBUG: Loading image from: %s\n", imagePath)
		img, err := loadImage(imagePath)
		if err != nil {
			fmt.Printf("DEBUG: Failed to load image %s: %v - using placeholder\n", filename, err)
			// If image doesn't exist, create a placeholder
			img = createPlaceholder()
		} else {
			fmt.Printf("DEBUG: Successfully loaded image: %s\n", filename)
		}
		images = append(images, img)
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no valid images to create collage")
	}

	// Create collage
	collage := createImageGrid(images, collageWidth, collageHeight)

	// Save collage
	outFile, err := os.Create(collagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create collage file: %v", err)
	}
	defer outFile.Close()

	// Encode as JPEG with good quality
	err = jpeg.Encode(outFile, collage, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf("failed to encode collage: %v", err)
	}

	fmt.Printf("DEBUG: Successfully created new collage: %s\n", collageFilename)
	return collageFilename, nil
}

// loadImage loads an image from file path
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// createPlaceholder creates a simple placeholder image
func createPlaceholder() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, thumbnailWidth, thumbnailHeight))
	// Fill with light gray color
	gray := image.NewUniform(color.RGBA{200, 200, 200, 255})
	draw.Draw(img, img.Bounds(), gray, image.Point{}, draw.Src)
	return img
}

// createImageGrid creates a grid collage from images
func createImageGrid(images []image.Image, totalWidth, totalHeight int) image.Image {
	numImages := len(images)
	if numImages == 0 {
		return createPlaceholder()
	}

	// Determine grid layout based on number of images
	var cols, rows int
	switch {
	case numImages == 1:
		cols, rows = 1, 1
	case numImages == 2:
		cols, rows = 2, 1
	case numImages <= 4:
		cols, rows = 2, 2
	case numImages <= 6:
		cols, rows = 3, 2
	default:
		cols, rows = 3, 2
		images = images[:6] // Limit to 6 images
	}

	// Calculate thumbnail size
	thumbWidth := totalWidth / cols
	thumbHeight := totalHeight / rows

	// Create collage canvas
	collage := image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))
	
	// Fill background with white
	white := image.NewUniform(color.RGBA{255, 255, 255, 255})
	draw.Draw(collage, collage.Bounds(), white, image.Point{}, draw.Src)

	// Place images in grid
	for i, img := range images {
		if i >= cols*rows {
			break
		}

		// Calculate position
		col := i % cols
		row := i / cols
		x := col * thumbWidth
		y := row * thumbHeight

		// Resize image to thumbnail size
		resized := imaging.Fill(img, thumbWidth, thumbHeight, imaging.Center, imaging.Lanczos)

		// Draw resized image onto collage
		dst := image.Rect(x, y, x+thumbWidth, y+thumbHeight)
		draw.Draw(collage, dst, resized, image.Point{}, draw.Src)
	}

	return collage
}