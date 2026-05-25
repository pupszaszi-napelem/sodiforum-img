package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const maxUploadSize = 12 << 20

func (a *app) upload(w http.ResponseWriter, r *http.Request) {
	if !a.limiter.allow(r.RemoteAddr) {
		writeUploadError(w, "Tul sok feltoltesi keres. Varj egy percet, aztan probald ujra.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeUploadError(w, "A kep tul nagy vagy nem olvashato. Maximum 12 MB engedelyezett.")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeUploadError(w, "Valassz ki egy kepet feltolteshez.")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil || len(data) == 0 || len(data) > maxUploadSize {
		writeUploadError(w, "A kepet nem sikerult beolvasni.")
		return
	}
	contentType := http.DetectContentType(data)
	if !allowedImageType(contentType) {
		writeUploadError(w, "Csak JPG, PNG, GIF vagy WebP kepet lehet feltolteni.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	imageURL, host, err := a.uploadImage(ctx, data, header.Filename)
	if err != nil {
		log.Printf("upload failed: %v", err)
		writeUploadError(w, "A feltoltes most nem sikerult. Probald ujra kesobb.")
		return
	}

	bbcode := fmt.Sprintf(`[img src=%s]`, imageURL)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, resultHTML, html.EscapeString(host), html.EscapeString(bbcode), html.EscapeString(imageURL))
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
