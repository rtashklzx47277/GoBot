package tools

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

var (
	imgurToken = os.Getenv("IMGUR_TOKEN")
	albumId    = os.Getenv("IMGUR_ALBUM")
)

func ImageCheck(oldImagePath, newImageUrl string) (int, string, error) {
	old, err := imageLoad(oldImagePath, "file")
	if err != nil {
		return 0, "", err
	}

	new, err := imageLoad(newImageUrl, "url")
	if err != nil {
		return 0, "", err
	}

	check := checkPixel(old, new)

	if check == 0 {
		url, err := imageChange(old, new)
		if err != nil {
			return 0, "", err
		}

		return 0, url, nil
	} else if check == 2 {
		return 2, "", nil
	}

	return 1, "", nil
}

func ImageUpload(imagePath string) (string, error) {
	pic, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file!\n%w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create form file!\n%w", err)
	}

	_, err = part.Write(pic)
	if err != nil {
		return "", fmt.Errorf("failed to write image data to form file!\n%w", err)
	}

	err = writer.WriteField("type", "file")
	if err != nil {
		return "", fmt.Errorf("failed to write field 'type'!\n%w", err)
	}

	err = writer.WriteField("album", albumId)
	if err != nil {
		return "", fmt.Errorf("failed to write field 'album'!\n%w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	reader, err := Post("https://api.imgur.com/3/image", body).
		AddHeader("Authorization", fmt.Sprintf("Bearer %s", imgurToken)).
		AddHeader("Content-Type", writer.FormDataContentType()).Do()
	if err != nil {
		return "", err
	}

	data, err := ToJson(reader)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://imgur.com/%s.png", data.Get("data").Get("id").String()), nil
}

func ImageDownload(imageUrl string, filePath ...string) error {
	reader, err := Get(imageUrl).Do()
	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("/bot/media/%s.jpg", strings.Join(filePath, "/")))
	if err != nil {
		return fmt.Errorf("failed to create file!\n%w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to copy data from reader to file!\n%w", err)
	}

	return nil
}

func ImageRemove(imagePath string) {
	err := os.Remove(imagePath)
	if err != nil {
		fmt.Println("failed to remove file!\n%w", err)
	}
}

func imageChange(old, new image.Image) (string, error) {
	width := max(old.Bounds().Max.X, new.Bounds().Max.X)
	height := max(old.Bounds().Max.Y, new.Bounds().Max.Y) * 2

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Over)

	oldX := (width - old.Bounds().Max.X) / 2
	oldY := (height/2 - old.Bounds().Max.Y) / 2
	draw.Draw(canvas, image.Rect(oldX, oldY, oldX+old.Bounds().Max.X, oldY+old.Bounds().Max.Y), old, image.Point{}, draw.Over)

	newX := (width - new.Bounds().Max.X) / 2
	newY := (height/2-new.Bounds().Max.Y)/2 + height/2
	draw.Draw(canvas, image.Rect(newX, newY, newX+new.Bounds().Max.X, newY+new.Bounds().Max.Y), new, image.Point{}, draw.Over)

	arrow, err := imageLoad("/bot/media/arrow.png", "file")
	if err != nil {
		return "", err
	}

	draw.Draw(canvas, image.Rect(width/2-100, height/2-100, width, height), arrow, image.Point{}, draw.Over)

	outputPath := "/bot/media/change.jpg"
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file!\n%w", err)
	}
	defer outputFile.Close()

	err = jpeg.Encode(outputFile, canvas, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encode JPEG image!\n%w", err)
	}

	link, err := ImageUpload(outputPath)
	if err != nil {
		return "", err
	}

	return link, nil
}

func imageLoad(imagePath, uploadFrom string) (image.Image, error) {
	var reader io.ReadCloser

	switch uploadFrom {
	case "file":
		file, err := os.Open(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file!\n%w", err)
		}
		defer file.Close()

		reader = file
	case "url":
		bytes, err := Get(imagePath).AddHeader("User-Agent", UserAgent).Do()
		if err != nil {
			return nil, err
		}

		reader = bytes
	}

	pic, err := ToImage(reader)
	if err != nil {
		return nil, err
	}

	return pic, err
}

func checkPixel(old, new image.Image) int {
	if old.Bounds() == image.Rect(0, 0, 480, 360) && new.Bounds() == image.Rect(0, 0, 1280, 720) {
		return 2
	} else if old.Bounds() != new.Bounds() {
		return 0
	}

	bounds := old.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if old.At(x, y) != new.At(x, y) {
				return 0
			}
		}
	}

	return 1
}
