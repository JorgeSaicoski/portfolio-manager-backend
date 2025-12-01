package test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/handler"
	"github.com/stretchr/testify/assert"
)

// Helper functions to create test images programmatically

// createTestPNG creates a PNG image of the specified dimensions
func createTestPNG(width, height int) *bytes.Buffer {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with test pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	return buf
}

// createTestJPEG creates a JPEG image of the specified dimensions
func createTestJPEG(width, height int) *bytes.Buffer {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{200, 100, 50, 255})
		}
	}
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	return buf
}

// createMultipartFileHeader creates a multipart.FileHeader for testing
func createMultipartFileHeader(filename string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Size:     size,
	}
}

// TestValidateImage tests the ValidateImage function
func TestValidateImage(t *testing.T) {
	t.Run("Success_ValidPNG", func(t *testing.T) {
		imgData := createTestPNG(200, 200)
		header := createMultipartFileHeader("test.png", int64(imgData.Len()))

		// Create a reader from the buffer
		reader := bytes.NewReader(imgData.Bytes())
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		assert.NoError(t, err)
	})

	t.Run("Success_ValidJPEG", func(t *testing.T) {
		imgData := createTestJPEG(800, 600)
		header := createMultipartFileHeader("test.jpg", int64(imgData.Len()))

		reader := bytes.NewReader(imgData.Bytes())
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		assert.NoError(t, err)
	})

	t.Run("Fail_TooLarge", func(t *testing.T) {
		imgData := createTestPNG(200, 200)
		// Header size exceeds MaxFileSize
		header := createMultipartFileHeader("test.png", handler.MaxFileSize+1)

		reader := bytes.NewReader(imgData.Bytes())
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file size exceeds")
	})

	t.Run("Success_ExactlyAtMaxFileSize", func(t *testing.T) {
		imgData := createTestPNG(200, 200)
		header := createMultipartFileHeader("test.png", handler.MaxFileSize)

		reader := bytes.NewReader(imgData.Bytes())
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		// Should succeed as it's at the limit, not exceeding
		assert.NoError(t, err)
	})

	t.Run("Fail_InvalidFormat_GIF", func(t *testing.T) {
		// GIF header
		gifData := []byte("GIF89a")
		header := createMultipartFileHeader("test.gif", int64(len(gifData)))

		reader := bytes.NewReader(gifData)
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid image file")
	})

	t.Run("Fail_CorruptedImageData", func(t *testing.T) {
		// Corrupted data - PNG signature but invalid content
		corruptedData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
		header := createMultipartFileHeader("test.png", int64(len(corruptedData)))

		reader := bytes.NewReader(corruptedData)
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid image file")
	})

	t.Run("Fail_ZeroByteFile", func(t *testing.T) {
		emptyData := []byte{}
		header := createMultipartFileHeader("test.png", 0)

		reader := bytes.NewReader(emptyData)
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		assert.Error(t, err)
	})

	t.Run("Fail_NonImageFile_Text", func(t *testing.T) {
		textData := []byte("This is just text, not an image")
		header := createMultipartFileHeader("test.txt", int64(len(textData)))

		reader := bytes.NewReader(textData)
		file := &mockFile{Reader: reader}

		err := handler.ValidateImage(file, header)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid image file")
	})
}

// TestGenerateUniqueFilename tests the GenerateUniqueFilename function
func TestGenerateUniqueFilename(t *testing.T) {
	t.Run("Success_GeneratesUniqueFilenames", func(t *testing.T) {
		filenames := make(map[string]bool)
		originalFilename := "test.png"

		// Generate 100 filenames and verify they're all unique
		for i := 0; i < 100; i++ {
			filename := handler.GenerateUniqueFilename(originalFilename)
			assert.NotContains(t, filenames, filename, "Filename should be unique")
			filenames[filename] = true
		}

		assert.Equal(t, 100, len(filenames), "Should have 100 unique filenames")
	})

	t.Run("Success_PreservesExtension_PNG", func(t *testing.T) {
		filename := handler.GenerateUniqueFilename("test.png")
		assert.True(t, strings.HasSuffix(filename, ".png"))
	})

	t.Run("Success_PreservesExtension_JPG", func(t *testing.T) {
		filename := handler.GenerateUniqueFilename("test.jpg")
		assert.True(t, strings.HasSuffix(filename, ".jpg"))
	})

	t.Run("Success_PreservesExtension_JPEG", func(t *testing.T) {
		filename := handler.GenerateUniqueFilename("image.jpeg")
		assert.True(t, strings.HasSuffix(filename, ".jpeg"))
	})

	t.Run("Success_HandlesEmptyExtension", func(t *testing.T) {
		filename := handler.GenerateUniqueFilename("noextension")
		assert.NotEmpty(t, filename)
		// Should not have an extension
		assert.False(t, strings.Contains(filename, "."))
	})

	t.Run("Success_ThreadSafety_ConcurrentGeneration", func(t *testing.T) {
		var wg sync.WaitGroup
		filenames := make([]string, 50)
		var mu sync.Mutex

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				filename := handler.GenerateUniqueFilename("concurrent.png")
				mu.Lock()
				filenames[index] = filename
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Check all filenames are unique
		uniqueMap := make(map[string]bool)
		for _, filename := range filenames {
			uniqueMap[filename] = true
		}

		assert.Equal(t, 50, len(uniqueMap), "All concurrently generated filenames should be unique")
	})
}

// TestOptimizeImage tests the OptimizeImage function
func TestOptimizeImage(t *testing.T) {
	t.Run("Success_ResizesLargeImage", func(t *testing.T) {
		// Create image larger than MaxImageWidth (1920)
		img := image.NewRGBA(image.Rect(0, 0, 2400, 1600))

		optimized := handler.OptimizeImage(img)

		bounds := optimized.Bounds()
		assert.Equal(t, handler.MaxImageWidth, bounds.Dx(), "Width should be resized to MaxImageWidth")
		// Height should be proportionally scaled: 1600 * (1920/2400) = 1280
		assert.Equal(t, 1280, bounds.Dy(), "Height should maintain aspect ratio")
	})

	t.Run("Success_DoesNotResizeSmallImage", func(t *testing.T) {
		// Create image smaller than MaxImageWidth
		originalWidth := 1200
		originalHeight := 800
		img := image.NewRGBA(image.Rect(0, 0, originalWidth, originalHeight))

		optimized := handler.OptimizeImage(img)

		bounds := optimized.Bounds()
		assert.Equal(t, originalWidth, bounds.Dx(), "Width should not change")
		assert.Equal(t, originalHeight, bounds.Dy(), "Height should not change")
	})

	t.Run("Success_VeryWideImage_PreservesAspectRatio", func(t *testing.T) {
		// Very wide panorama: 4000x100
		img := image.NewRGBA(image.Rect(0, 0, 4000, 100))

		optimized := handler.OptimizeImage(img)

		bounds := optimized.Bounds()
		assert.Equal(t, handler.MaxImageWidth, bounds.Dx())
		// Height: 100 * (1920/4000) = 48
		assert.Equal(t, 48, bounds.Dy())
	})

	t.Run("Success_VeryTallImage_PreservesAspectRatio", func(t *testing.T) {
		// Very tall image: 100x4000
		img := image.NewRGBA(image.Rect(0, 0, 100, 4000))

		optimized := handler.OptimizeImage(img)

		bounds := optimized.Bounds()
		// Width is already less than max, should not resize
		assert.Equal(t, 100, bounds.Dx())
		assert.Equal(t, 4000, bounds.Dy())
	})

	t.Run("Success_SquareImage_ResizesMaintainingSquare", func(t *testing.T) {
		// Square image larger than max: 2000x2000
		img := image.NewRGBA(image.Rect(0, 0, 2000, 2000))

		optimized := handler.OptimizeImage(img)

		bounds := optimized.Bounds()
		assert.Equal(t, handler.MaxImageWidth, bounds.Dx())
		assert.Equal(t, handler.MaxImageWidth, bounds.Dy())
	})

	t.Run("Success_ExactlyAtMaxDimensions", func(t *testing.T) {
		// Image exactly at MaxImageWidth
		img := image.NewRGBA(image.Rect(0, 0, handler.MaxImageWidth, 1080))

		optimized := handler.OptimizeImage(img)

		bounds := optimized.Bounds()
		assert.Equal(t, handler.MaxImageWidth, bounds.Dx())
		assert.Equal(t, 1080, bounds.Dy())
	})
}

// TestGenerateThumbnail tests the GenerateThumbnail function
func TestGenerateThumbnail(t *testing.T) {
	t.Run("Success_LargeImage_FitsWithinSize", func(t *testing.T) {
		// Large image: 1200x800
		img := image.NewRGBA(image.Rect(0, 0, 1200, 800))

		thumbnail := handler.GenerateThumbnail(img, handler.ThumbnailSize)

		bounds := thumbnail.Bounds()
		// Should fit within 400x400, maintaining aspect ratio
		assert.LessOrEqual(t, bounds.Dx(), handler.ThumbnailSize)
		assert.LessOrEqual(t, bounds.Dy(), handler.ThumbnailSize)
		// For 1200x800, aspect ratio 3:2, should be 400x267 (allow ±1 pixel for rounding)
		assert.Equal(t, 400, bounds.Dx())
		assert.InDelta(t, 267, bounds.Dy(), 1.0)
	})

	t.Run("Success_SmallImage_DoesNotEnlarge", func(t *testing.T) {
		// Small image: 200x200
		img := image.NewRGBA(image.Rect(0, 0, 200, 200))

		thumbnail := handler.GenerateThumbnail(img, handler.ThumbnailSize)

		bounds := thumbnail.Bounds()
		// imaging.Fit should not enlarge, keeps at 200x200
		assert.Equal(t, 200, bounds.Dx())
		assert.Equal(t, 200, bounds.Dy())
	})

	t.Run("Success_WideRectangle_MaintainsAspect", func(t *testing.T) {
		// Wide image: 1600x400
		img := image.NewRGBA(image.Rect(0, 0, 1600, 400))

		thumbnail := handler.GenerateThumbnail(img, handler.ThumbnailSize)

		bounds := thumbnail.Bounds()
		// Aspect ratio 4:1, should fit to 400x100
		assert.Equal(t, 400, bounds.Dx())
		assert.Equal(t, 100, bounds.Dy())
	})

	t.Run("Success_TallRectangle_MaintainsAspect", func(t *testing.T) {
		// Tall image: 400x1600
		img := image.NewRGBA(image.Rect(0, 0, 400, 1600))

		thumbnail := handler.GenerateThumbnail(img, handler.ThumbnailSize)

		bounds := thumbnail.Bounds()
		// Aspect ratio 1:4, should fit to 100x400
		assert.Equal(t, 100, bounds.Dx())
		assert.Equal(t, 400, bounds.Dy())
	})

	t.Run("Success_SquareImage_FitsToSquare", func(t *testing.T) {
		// Square image: 800x800
		img := image.NewRGBA(image.Rect(0, 0, 800, 800))

		thumbnail := handler.GenerateThumbnail(img, handler.ThumbnailSize)

		bounds := thumbnail.Bounds()
		// Should be exactly 400x400
		assert.Equal(t, 400, bounds.Dx())
		assert.Equal(t, 400, bounds.Dy())
	})
}

// TestSaveImage tests the SaveImage function
func TestSaveImage(t *testing.T) {
	t.Run("Success_SavePNG", func(t *testing.T) {
		// Test SaveImage with existing directories
		pngData := createTestPNG(200, 200)
		reader := bytes.NewReader(pngData.Bytes())
		file := &mockFile{Reader: reader}

		filename := "test-save.png"
		originalPath, thumbnailPath, err := handler.SaveImage(file, filename)

		assert.NoError(t, err)
		assert.NotEmpty(t, originalPath)
		assert.NotEmpty(t, thumbnailPath)
		assert.Contains(t, originalPath, filename)
		assert.Contains(t, thumbnailPath, filename)

		// Cleanup - remove test files
		os.Remove(strings.Replace(originalPath, "/uploads", "./uploads", 1))
		os.Remove(strings.Replace(thumbnailPath, "/uploads", "./uploads", 1))
	})

	t.Run("Success_SaveJPEG", func(t *testing.T) {
		jpegData := createTestJPEG(300, 300)
		reader := bytes.NewReader(jpegData.Bytes())
		file := &mockFile{Reader: reader}

		filename := "test-save.jpg"
		originalPath, thumbnailPath, err := handler.SaveImage(file, filename)

		assert.NoError(t, err)
		assert.NotEmpty(t, originalPath)
		assert.NotEmpty(t, thumbnailPath)

		// Cleanup
		os.Remove(strings.Replace(originalPath, "/uploads", "./uploads", 1))
		os.Remove(strings.Replace(thumbnailPath, "/uploads", "./uploads", 1))
	})

	t.Run("Success_CreatesDirectoriesIfNotExist", func(t *testing.T) {
		// The function should create directories if they don't exist
		// This is already tested implicitly, but we verify
		pngData := createTestPNG(100, 100)
		reader := bytes.NewReader(pngData.Bytes())
		file := &mockFile{Reader: reader}

		filename := "test-directories.png"
		originalPath, thumbnailPath, err := handler.SaveImage(file, filename)

		assert.NoError(t, err)

		// Verify directories exist
		assert.DirExists(t, handler.OriginalDir)
		assert.DirExists(t, handler.ThumbnailDir)

		// Cleanup
		os.Remove(strings.Replace(originalPath, "/uploads", "./uploads", 1))
		os.Remove(strings.Replace(thumbnailPath, "/uploads", "./uploads", 1))
	})

	t.Run("Fail_InvalidImageData", func(t *testing.T) {
		// Invalid image data
		invalidData := []byte("not an image")
		reader := bytes.NewReader(invalidData)
		file := &mockFile{Reader: reader}

		filename := "invalid.png"
		originalPath, thumbnailPath, err := handler.SaveImage(file, filename)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
		assert.Empty(t, originalPath)
		assert.Empty(t, thumbnailPath)
	})

	t.Run("Success_VerifyFileExistsAfterSave", func(t *testing.T) {
		pngData := createTestPNG(150, 150)
		reader := bytes.NewReader(pngData.Bytes())
		file := &mockFile{Reader: reader}

		filename := "verify-exists.png"
		originalPath, thumbnailPath, err := handler.SaveImage(file, filename)

		assert.NoError(t, err)

		// Convert URLs to filesystem paths
		originalFsPath := strings.Replace(originalPath, "/uploads", "./uploads", 1)
		thumbnailFsPath := strings.Replace(thumbnailPath, "/uploads", "./uploads", 1)

		// Verify files exist
		assert.FileExists(t, originalFsPath)
		assert.FileExists(t, thumbnailFsPath)

		// Cleanup
		os.Remove(originalFsPath)
		os.Remove(thumbnailFsPath)
	})

	t.Run("Success_OptimizesLargeImage", func(t *testing.T) {
		// Create a large image > MaxImageWidth
		largePng := createTestPNG(2400, 1600)
		reader := bytes.NewReader(largePng.Bytes())
		file := &mockFile{Reader: reader}

		filename := "large-optimized.png"
		originalPath, thumbnailPath, err := handler.SaveImage(file, filename)

		assert.NoError(t, err)
		assert.NotEmpty(t, originalPath)
		assert.NotEmpty(t, thumbnailPath)

		// Cleanup
		os.Remove(strings.Replace(originalPath, "/uploads", "./uploads", 1))
		os.Remove(strings.Replace(thumbnailPath, "/uploads", "./uploads", 1))
	})

	t.Run("Success_GeneratesThumbnail", func(t *testing.T) {
		pngData := createTestPNG(800, 600)
		reader := bytes.NewReader(pngData.Bytes())
		file := &mockFile{Reader: reader}

		filename := "thumbnail-test.png"
		originalPath, thumbnailPath, err := handler.SaveImage(file, filename)

		assert.NoError(t, err)

		// Verify both paths returned
		assert.NotEmpty(t, originalPath)
		assert.NotEmpty(t, thumbnailPath)
		assert.NotEqual(t, originalPath, thumbnailPath)

		// Cleanup
		os.Remove(strings.Replace(originalPath, "/uploads", "./uploads", 1))
		os.Remove(strings.Replace(thumbnailPath, "/uploads", "./uploads", 1))
	})
}

// TestDeleteImageFiles tests the DeleteImageFiles function
func TestDeleteImageFiles(t *testing.T) {
	t.Run("Success_DeleteExistingFiles", func(t *testing.T) {
		// Create test files
		os.MkdirAll(handler.OriginalDir, 0755)
		os.MkdirAll(handler.ThumbnailDir, 0755)

		originalPath := filepath.Join(handler.OriginalDir, "delete-test.png")
		thumbnailPath := filepath.Join(handler.ThumbnailDir, "delete-test.png")

		// Create actual files
		os.WriteFile(originalPath, []byte("test"), 0644)
		os.WriteFile(thumbnailPath, []byte("test"), 0644)

		// Convert to URL format
		originalURL := strings.Replace(originalPath, "./", "/", 1)
		thumbnailURL := strings.Replace(thumbnailPath, "./", "/", 1)

		err := handler.DeleteImageFiles(originalURL, thumbnailURL)
		assert.NoError(t, err)

		// Verify files are deleted
		_, err1 := os.Stat(originalPath)
		_, err2 := os.Stat(thumbnailPath)
		assert.True(t, os.IsNotExist(err1))
		assert.True(t, os.IsNotExist(err2))
	})

	t.Run("Success_HandlesMissingFiles", func(t *testing.T) {
		// Non-existent files should not error
		originalURL := "/uploads/images/original/nonexistent.png"
		thumbnailURL := "/uploads/images/thumbnail/nonexistent.png"

		err := handler.DeleteImageFiles(originalURL, thumbnailURL)
		assert.NoError(t, err, "Should not error when files don't exist")
	})

	t.Run("Success_PartialDeletion_OnlyOriginalExists", func(t *testing.T) {
		os.MkdirAll(handler.OriginalDir, 0755)

		originalPath := filepath.Join(handler.OriginalDir, "partial.png")
		os.WriteFile(originalPath, []byte("test"), 0644)

		originalURL := strings.Replace(originalPath, "./", "/", 1)
		thumbnailURL := "/uploads/images/thumbnail/partial.png" // Doesn't exist

		err := handler.DeleteImageFiles(originalURL, thumbnailURL)
		assert.NoError(t, err)

		// Original should be deleted
		_, err1 := os.Stat(originalPath)
		assert.True(t, os.IsNotExist(err1))
	})

	t.Run("Success_IdempotentDeletion", func(t *testing.T) {
		// Create and delete files
		os.MkdirAll(handler.OriginalDir, 0755)
		os.MkdirAll(handler.ThumbnailDir, 0755)

		originalPath := filepath.Join(handler.OriginalDir, "idempotent.png")
		thumbnailPath := filepath.Join(handler.ThumbnailDir, "idempotent.png")

		os.WriteFile(originalPath, []byte("test"), 0644)
		os.WriteFile(thumbnailPath, []byte("test"), 0644)

		originalURL := strings.Replace(originalPath, "./", "/", 1)
		thumbnailURL := strings.Replace(thumbnailPath, "./", "/", 1)

		// First deletion
		err := handler.DeleteImageFiles(originalURL, thumbnailURL)
		assert.NoError(t, err)

		// Second deletion - should not error
		err = handler.DeleteImageFiles(originalURL, thumbnailURL)
		assert.NoError(t, err)
	})
}

// TestCopyFile tests the CopyFile function
func TestCopyFile(t *testing.T) {
	t.Run("Success_CopyFile", func(t *testing.T) {
		tempDir := t.TempDir()

		srcPath := filepath.Join(tempDir, "source.txt")
		dstPath := filepath.Join(tempDir, "destination.txt")

		// Create source file
		testContent := []byte("Test file content for copying")
		err := os.WriteFile(srcPath, testContent, 0644)
		assert.NoError(t, err)

		// Copy file
		err = handler.CopyFile(srcPath, dstPath)
		assert.NoError(t, err)

		// Verify destination exists and has same content
		dstContent, err := os.ReadFile(dstPath)
		assert.NoError(t, err)
		assert.Equal(t, testContent, dstContent)
	})

	t.Run("Fail_SourceFileDoesNotExist", func(t *testing.T) {
		tempDir := t.TempDir()

		srcPath := filepath.Join(tempDir, "nonexistent.txt")
		dstPath := filepath.Join(tempDir, "destination.txt")

		err := handler.CopyFile(srcPath, dstPath)
		assert.Error(t, err)
	})

	t.Run("Fail_DestinationDirectoryDoesNotExist", func(t *testing.T) {
		tempDir := t.TempDir()

		srcPath := filepath.Join(tempDir, "source.txt")
		dstPath := filepath.Join(tempDir, "nonexistent-dir", "destination.txt")

		// Create source
		os.WriteFile(srcPath, []byte("test"), 0644)

		err := handler.CopyFile(srcPath, dstPath)
		assert.Error(t, err)
	})

	t.Run("Success_VerifyContentIntegrity", func(t *testing.T) {
		tempDir := t.TempDir()

		srcPath := filepath.Join(tempDir, "binary-source.bin")
		dstPath := filepath.Join(tempDir, "binary-dest.bin")

		// Create binary content
		binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		os.WriteFile(srcPath, binaryContent, 0644)

		err := handler.CopyFile(srcPath, dstPath)
		assert.NoError(t, err)

		// Verify exact byte match
		dstContent, _ := os.ReadFile(dstPath)
		assert.Equal(t, binaryContent, dstContent)
	})
}

// mockFile implements multipart.File interface for testing
type mockFile struct {
	*bytes.Reader
	closed bool
}

func (m *mockFile) Close() error {
	m.closed = true
	return nil
}

func (m *mockFile) Read(p []byte) (n int, err error) {
	return m.Reader.Read(p)
}

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	return m.Reader.ReadAt(p, off)
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	return m.Reader.Seek(offset, whence)
}
