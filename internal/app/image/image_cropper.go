package image

import (
	nImage "image"
	"image/gif"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
)

type HeaderFile struct {
	Mime       string
	FileHeader *multipart.FileHeader
}

// DetectMimeOfHeaderFiles retrieve file mimetype from the first 512 byte of the file.
func DetectMimeOfHeaderFiles(ihfs []*multipart.FileHeader) ([]*HeaderFile, error) {
	hfs := make([]*HeaderFile, 0)

	for _, h := range ihfs {
		f, err := h.Open()

		if err != nil {
			return hfs, err
		}

		defer f.Close()

		headBytes := make([]byte, 512)

		if _, err := f.Read(headBytes); err != nil {
			return hfs, err
		}

		mime := http.DetectContentType(headBytes)

		hfs = append(hfs, &HeaderFile{
			Mime:       mime,
			FileHeader: h,
		})
	}

	return hfs, nil
}

func calcImageCenterCoord(image nImage.Image) (cx int, cy int) {
	rect := image.Bounds()

	cx = (rect.Max.X + rect.Min.X) / 2
	cy = (rect.Max.Y + rect.Min.Y) / 2

	return cx, cy
}

type DecodedImage struct {
	Mime  string
	Image nImage.Image
}

func decodeImageByMime(r io.Reader, mime string) (*DecodedImage, error) {
	var (
		img nImage.Image
		err error
	)

	switch mime {
	case "image/png":
		img, err = png.Decode(r)
	case "image/jpeg":
		img, err = jpeg.Decode(r)
	case "image/gif":
		img, err = gif.Decode(r)
	default:
		img, _, err = nImage.Decode(r)
	}

	if err != nil {
		return (*DecodedImage)(nil), err
	}

	return &DecodedImage{
		Image: img,
		Mime:  mime,
	}, err
}

type SubImager interface {
	SubImage(r nImage.Rectangle) nImage.Image
}

const (
	ThumbnailHeight = 150
	ThumbnailWidth  = 150
)

type SizedImage struct {
	Name      string
	Mime      string
	OrigImage nImage.Image
	Thumbnail nImage.Image
}

// imageUnderSize determine whether
func skipThumbnailCropping(img nImage.Image) bool {
	rect := img.Bounds()

	l := rect.Max.X - rect.Min.X
	h := rect.Max.Y - rect.Min.Y

	return l <= 150 && h <= 150
}

// CropThumbnaie crops the image to size of 150x150 to save client bandwidth.
func CropThumbnail(ihfs []*multipart.FileHeader) ([]*SizedImage, error) {
	hfs, err := DetectMimeOfHeaderFiles(ihfs)

	if err != nil {
		return nil, err
	}

	sis := make([]*SizedImage, 0)

	for _, hf := range hfs {

		fr, err := hf.FileHeader.Open()

		if err != nil {
			return nil, err
		}

		dImg, err := decodeImageByMime(fr, hf.Mime)

		if err != nil {
			return nil, err
		}

		// If image size is below 150x150, we skip cropping.
		if skipThumbnailCropping(dImg.Image) {
			sis = append(
				sis,
				&SizedImage{
					Name:      hf.FileHeader.Filename,
					Mime:      hf.Mime,
					OrigImage: dImg.Image,
					Thumbnail: dImg.Image,
				},
			)

			continue
		}

		cx, cy := calcImageCenterCoord(dImg.Image)

		simg := dImg.Image.(SubImager)

		// Crop the original image to thumbnail (150x150).
		cimg := simg.SubImage(nImage.Rect(cx+-75, cy-75, cx+75, cy+75))

		si := &SizedImage{
			Name:      hf.FileHeader.Filename,
			Mime:      hf.Mime,
			OrigImage: dImg.Image,
			Thumbnail: cimg,
		}

		sis = append(sis, si)
	}

	return sis, nil
}
