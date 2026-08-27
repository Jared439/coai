package utils

import (
	"chat/globals"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chai2010/webp"
)

type Image struct {
	Object  image.Image
	Content string
}

type Images []Image

func NewImage(url string) (*Image, error) {
	if strings.HasPrefix(url, "data:image/") {
		data := SafeSplit(url, ",", 2)
		if data[1] == "" {
			return nil, nil
		}

		decoded, err := Base64Decode(data[1])
		if err != nil {
			return nil, err
		}

		img, _, err := image.Decode(strings.NewReader(string(decoded)))
		if err != nil {
			return nil, err
		}

		return &Image{
			Object:  img,
			Content: url,
		}, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var img image.Image
	suffix := strings.ToLower(path.Ext(url))

	switch suffix {
	case ".png":
		if img, _, err = image.Decode(res.Body); err != nil {
			return nil, err
		}

	case ".jpg", ".jpeg":
		if img, err = jpeg.Decode(res.Body); err != nil {
			return nil, err
		}

	case ".webp":
		if img, err = webp.Decode(res.Body); err != nil {
			return nil, err
		}

	case ".gif":
		ticks, decodeErr := gif.DecodeAll(res.Body)
		if decodeErr != nil {
			return nil, decodeErr
		}

		if len(ticks.Image) == 0 {
			return nil, fmt.Errorf("GIF contains no image frames")
		}

		img = ticks.Image[0]

	default:
		return nil, fmt.Errorf("unsupported image format: %s", suffix)
	}

	return &Image{
		Object:  img,
		Content: url,
	}, nil
}

func NewImageContent(content string) *Image {
	return &Image{
		Content: content,
	}
}

func ConvertToBase64(url string) (string, error) {
	if strings.HasPrefix(url, "data:image/") {
		data := strings.SplitN(url, ",", 2)
		if len(data) != 2 {
			return "", nil
		}

		return data[1], nil
	}

	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return Base64EncodeBytes(data), nil
}

func (i *Image) GetWidth() int {
	if i == nil || i.Object == nil {
		return 0
	}

	return i.Object.Bounds().Dx()
}

func (i *Image) GetHeight() int {
	if i == nil || i.Object == nil {
		return 0
	}

	return i.Object.Bounds().Dy()
}

func (i *Image) GetPixel(
	x int,
	y int,
) (uint32, uint32, uint32, uint32) {
	if i == nil || i.Object == nil {
		return 0, 0, 0, 0
	}

	return i.Object.At(x, y).RGBA()
}

func (i *Image) GetPixelColor(x int, y int) (int, int, int) {
	r, g, b, _ := i.GetPixel(x, y)
	return int(r), int(g), int(b)
}

func (i *Image) CountTokens(model string) int {
	if i == nil || i.Object == nil {
		return 0
	}

	if globals.IsVisionModel(model) {
		// Tile size is 512x512.
		// Images larger than 2048x2048 are limited to 16 tiles.
		x := LimitMax(
			math.Ceil(float64(i.GetWidth())/512),
			4,
		)
		y := LimitMax(
			math.Ceil(float64(i.GetHeight())/512),
			4,
		)

		tiles := int(x) * int(y)
		return 85 + 170*tiles
	}

	return 0
}

func (i *Image) IsBase64() bool {
	return i != nil &&
		strings.HasPrefix(i.Content, "data:image/")
}

func (i *Image) GetType() string {
	if i == nil {
		return ""
	}

	// Examples: image/jpeg, image/png, image/gif.
	if i.IsBase64() {
		t := SafeSplit(i.Content, ";", 2)[0]
		return strings.ReplaceAll(t, "data:", "")
	}

	switch strings.ToLower(path.Ext(i.Content)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"

	case ".png":
		return "image/png"

	case ".gif":
		return "image/gif"

	case ".webp":
		return "image/webp"

	case ".bmp":
		return "image/bmp"

	default:
		return ""
	}
}

func (i *Image) ToBase64() string {
	if i == nil {
		return ""
	}

	if i.IsBase64() {
		return i.Content
	}

	data, err := ConvertToBase64(i.Content)
	if err != nil {
		globals.Warn(
			fmt.Sprintf(
				"cannot convert image to base64: %s",
				err.Error(),
			),
		)
		return ""
	}

	return fmt.Sprintf(
		"data:%s;base64,%s",
		i.GetType(),
		data,
	)
}

func (i *Image) ToRawBase64() string {
	if i == nil {
		return ""
	}

	if i.IsBase64() {
		return SafeSplit(i.Content, ",", 2)[1]
	}

	data, err := ConvertToBase64(i.Content)
	if err != nil {
		globals.Warn(
			fmt.Sprintf(
				"cannot convert image to base64: %s",
				err.Error(),
			),
		)
		return ""
	}

	return data
}

func DownloadImage(url string, filePath string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer func(body io.ReadCloser) {
		closeErr := body.Close()
		if closeErr != nil {
			globals.Debug(
				"[utils] close response body error: %s (path: %s)",
				closeErr.Error(),
				filePath,
			)
		}
	}(res.Body)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			globals.Debug(
				"[utils] close file error: %s (path: %s)",
				closeErr.Error(),
				filePath,
			)
		}
	}(file)

	_, err = io.Copy(file, res.Body)
	return err
}

func StoreImage(url string) string {
	if !globals.AcceptImageStore {
		return url
	}

	hash := Md5Encrypt(url) + path.Ext(url)
	filePath := fmt.Sprintf(
		"storage/attachments/%s",
		hash,
	)

	if err := DownloadImage(url, filePath); err != nil {
		globals.Warn(
			fmt.Sprintf(
				"[utils] save image error: %s",
				err.Error(),
			),
		)
		return url
	}

	return fmt.Sprintf(
		"%s/attachments/%s",
		globals.NotifyUrl,
		hash,
	)
}
