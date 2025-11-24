package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

const (
	MaxFileSize   = 10 * 1024 * 1024 // 10 MB
	MaxImageWidth = 1920
	ThumbnailSize = 400
	JPEGQuality   = 85
	UploadDir     = "./uploads/images"
	OriginalDir   = "./uploads/images/original"
	ThumbnailDir  = "./uploads/images/thumbnail"
)

var AllowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// ValidateImage validates the uploaded image file
func ValidateImage(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)
	}

	// Check mime type
	mimeType := header.Header.Get("Content-Type")
	if !AllowedMimeTypes[mimeType] {
		return fmt.Errorf("invalid file type: %s. Allowed types: jpeg, png, webp", mimeType)
	}

	// Try to decode the image to ensure it's valid
	_, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %v", err)
	}

	// Reset file pointer to beginning
	file.Seek(0, 0)

	return nil
}

// GenerateUniqueFilename generates a unique filename using timestamp and hash
func GenerateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().UnixNano()

	// Create a hash of the original filename + timestamp
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s%d", originalFilename, timestamp)))
	hashStr := hex.EncodeToString(hash.Sum(nil))[:12]

	return fmt.Sprintf("%d_%s%s", timestamp, hashStr, ext)
}

// OptimizeImage resizes and compresses the image
func OptimizeImage(src image.Image) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()

	// Only resize if image is larger than max width
	if width > MaxImageWidth {
		return imaging.Resize(src, MaxImageWidth, 0, imaging.Lanczos)
	}

	return src
}

// GenerateThumbnail creates a thumbnail version of the image
func GenerateThumbnail(src image.Image, size int) image.Image {
	return imaging.Fit(src, size, size, imaging.Lanczos)
}

// SaveImage saves both the optimized original and thumbnail versions
func SaveImage(file multipart.File, filename string) (originalPath, thumbnailPath string, err error) {
	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Ensure directories exist
	if err := os.MkdirAll(OriginalDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create original directory: %v", err)
	}
	if err := os.MkdirAll(ThumbnailDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create thumbnail directory: %v", err)
	}

	// Optimize the original image
	optimized := OptimizeImage(img)

	// Save optimized original
	originalPath = filepath.Join(OriginalDir, filename)
	if err := saveImageFile(optimized, originalPath, format); err != nil {
		return "", "", fmt.Errorf("failed to save original image: %v", err)
	}

	// Generate and save thumbnail
	thumbnail := GenerateThumbnail(optimized, ThumbnailSize)
	thumbnailPath = filepath.Join(ThumbnailDir, filename)
	if err := saveImageFile(thumbnail, thumbnailPath, format); err != nil {
		// Clean up original if thumbnail fails
		os.Remove(originalPath)
		return "", "", fmt.Errorf("failed to save thumbnail: %v", err)
	}

	// Return URL paths (not filesystem paths)
	originalURL := strings.Replace(originalPath, "./", "/", 1)
	thumbnailURL := strings.Replace(thumbnailPath, "./", "/", 1)

	return originalURL, thumbnailURL, nil
}

// saveImageFile saves an image to disk with appropriate compression
func saveImageFile(img image.Image, path string, format string) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch format {
	case "jpeg", "jpg":
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: JPEGQuality})
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		return encoder.Encode(outFile, img)
	default:
		// Default to JPEG for other formats
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: JPEGQuality})
	}
}

// DeleteImageFiles removes both the original and thumbnail files from disk
func DeleteImageFiles(originalURL, thumbnailURL string) error {
	// Convert URLs back to filesystem paths
	originalPath := strings.Replace(originalURL, "/uploads", "./uploads", 1)
	thumbnailPath := strings.Replace(thumbnailURL, "/uploads", "./uploads", 1)

	// Delete original
	if err := os.Remove(originalPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete original image: %v", err)
	}

	// Delete thumbnail
	if err := os.Remove(thumbnailPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete thumbnail image: %v", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
