package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	maxUploadSize  = 12 << 20
	maxRequestSize = maxUploadSize + (1 << 20)
)

type uploadedImage struct {
	data     []byte
	filename string
}

func (a *app) upload(w http.ResponseWriter, r *http.Request) {
	if !a.limiter.allow(r.RemoteAddr) {
		writeUploadError(w, "Tul sok feltoltesi keres. Varj egy percet, aztan probald ujra.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	image, err := readUploadedImage(r)
	if err != nil {
		writeUploadError(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	imageURL, host, err := a.uploadImage(ctx, image.data, image.filename)
	if err != nil {
		log.Printf("upload failed: %v", err)
		writeUploadError(w, "A feltoltes most nem sikerult. Probald ujra kesobb.")
		return
	}

	bbcode := fmt.Sprintf(`[img src=%s]`, imageURL)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, resultHTML, html.EscapeString(host), html.EscapeString(bbcode), html.EscapeString(imageURL))
}

func readUploadedImage(r *http.Request) (uploadedImage, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return uploadedImage{}, errors.New("A kep tul nagy vagy nem olvashato. Maximum 12 MB engedelyezett.")
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			return uploadedImage{}, errors.New("Valassz ki egy kepet feltolteshez.")
		}
		if err != nil {
			return uploadedImage{}, errors.New("A kep tul nagy vagy nem olvashato. Maximum 12 MB engedelyezett.")
		}
		if part.FormName() != "image" {
			_ = part.Close()
			continue
		}
		defer part.Close()

		var buffer bytes.Buffer
		if _, err := io.Copy(&buffer, io.LimitReader(part, maxUploadSize+1)); err != nil {
			return uploadedImage{}, errors.New("A kepet nem sikerult beolvasni.")
		}
		data := buffer.Bytes()
		if len(data) == 0 || len(data) > maxUploadSize {
			return uploadedImage{}, errors.New("A kep tul nagy vagy nem olvashato. Maximum 12 MB engedelyezett.")
		}

		contentType := http.DetectContentType(data)
		if !allowedImageType(contentType) {
			return uploadedImage{}, errors.New("Csak JPG, PNG, GIF vagy WebP kepet lehet feltolteni.")
		}

		return uploadedImage{
			data:     data,
			filename: part.FileName(),
		}, nil
	}
}

func cleanFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	if name == "" {
		return "upload.jpg"
	}
	return name
}

func allowedImageType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func writeUploadError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div id="upload-result" class="result error" data-ready="true"><strong>Hiba</strong><p>%s</p></div>`, html.EscapeString(message))
}
