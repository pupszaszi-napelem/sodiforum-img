package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const (
	imgbbUploadURL  = "https://api.imgbb.com/1/upload"
	catboxUploadURL = "https://catbox.moe/user/api.php"
	responseBodyMax = 1 << 20
)

func (a *app) uploadImage(ctx context.Context, data []byte, filename string) (string, string, error) {
	uploaded, catboxErr := a.uploadToCatbox(ctx, data, filename)
	if catboxErr == nil {
		return uploaded, "catbox.moe", nil
	}

	var imgbbErr error
	if a.imgbbKey != "" {
		if uploaded, err := a.uploadToImgBB(ctx, data); err == nil {
			return uploaded, "imgbb", nil
		} else {
			imgbbErr = err
		}
	} else {
		imgbbErr = errors.New("IMGBB_API_KEY nincs beallitva")
	}

	return "", "", fmt.Errorf("catbox hiba: %v; imgbb hiba: %v", catboxErr, imgbbErr)
}

func (a *app) uploadToImgBB(ctx context.Context, data []byte) (string, error) {
	form := url.Values{}
	form.Set("key", a.imgbbKey)
	form.Set("image", base64.StdEncoding.EncodeToString(data))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, imgbbUploadURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, responseBodyMax))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var parsed struct {
		Success bool `json:"success"`
		Data    struct {
			URL string `json:"url"`
		} `json:"data"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if !parsed.Success || parsed.Data.URL == "" {
		return "", fmt.Errorf("ervenytelen imgbb valasz")
	}
	if !validUploadedURL(parsed.Data.URL) {
		return "", fmt.Errorf("ervenytelen imgbb URL")
	}
	return parsed.Data.URL, nil
}

func (a *app) uploadToCatbox(ctx context.Context, data []byte, filename string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("reqtype", "fileupload"); err != nil {
		return "", err
	}
	part, err := writer.CreateFormFile("fileToUpload", cleanFilename(filename))
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, catboxUploadURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "sodiforum-img/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, responseBodyMax))
	text := strings.TrimSpace(string(respBody))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if !validUploadedURL(text) {
		return "", errors.New(text)
	}
	return text, nil
}

func validUploadedURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	return !strings.ContainsAny(raw, "\"'[] \t\r\n")
}
