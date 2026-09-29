package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTimeout  = 60 * time.Second
	UploadTimeout   = 30 * time.Minute
	DownloadTimeout = 60 * time.Minute
	UserAgentHeader = "OhioFiles-CLI/1.0"
	DefaultChunkSize = 8 * 1024 * 1024 // 8 MB
	LargeFileThreshold = 50 * 1024 * 1024 // 50 MB
)

// ProgressWriter tracks written bytes and reports to callback.
type ProgressWriter struct {
	Total    int64
	Written  int64
	Callback func(written, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Written += int64(n)
	if pw.Callback != nil {
		pw.Callback(pw.Written, pw.Total)
	}
	return n, nil
}

// ProgressReader tracks read bytes and reports to callback.
type ProgressReader struct {
	Reader   io.Reader
	Total    int64
	Current  int64
	Callback func(written, total int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	if pr.Callback != nil {
		pr.Callback(pr.Current, pr.Total)
	}
	return n, err
}

// Client interacts with the OhioFiles REST API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	SessionKey string
	DeviceId   string
	UserAgent  string
}

// NewClient creates a new API client instance.
func NewClient(baseURL, sessionKey, deviceId string) *Client {
	if baseURL == "" {
		baseURL = "https://api.ohiofiles.cloud"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		SessionKey: strings.TrimSpace(sessionKey),
		DeviceId:   strings.TrimSpace(deviceId),
		UserAgent:  UserAgentHeader,
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	fullURL := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.UserAgent)
	if c.SessionKey != "" {
		req.Header.Set("X-Session-Key", c.SessionKey)
	}
	if c.DeviceId != "" {
		req.Header.Set("X-Device-Id", c.DeviceId)
	}
	return req, nil
}

func (c *Client) doRequest(req *http.Request, target interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && (errResp.Error != "" || errResp.Message != "") {
			msg := errResp.Message
			if msg == "" {
				msg = errResp.Error
			}
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	if target != nil {
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			return fmt.Errorf("failed to decode JSON response: %w", err)
		}
	}
	return nil
}

// CreatePublicSession creates or retrieves a public session (memorable or secure).
func (c *Client) CreatePublicSession(keyType string) (*PublicSession, error) {
	if keyType == "" {
		keyType = "memorable"
	}
	payload := map[string]string{
		"keyType": keyType,
	}
	data, _ := json.Marshal(payload)

	req, err := c.newRequest(context.Background(), http.MethodPost, "/public-sessions", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var res SessionResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	if !res.Success || res.Session == nil {
		return nil, errors.New("failed to create public session: empty response")
	}
	return res.Session, nil
}

// GetPublicSession retrieves public session information by key.
func (c *Client) GetPublicSession(key string) (*PublicSession, error) {
	if key == "" {
		key = c.SessionKey
	}
	if key == "" {
		return nil, errors.New("session key is required")
	}

	path := fmt.Sprintf("/public-sessions/%s", url.PathEscape(key))
	req, err := c.newRequest(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var res SessionResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	if !res.Success || res.Session == nil {
		return nil, errors.New("session not found")
	}
	return res.Session, nil
}

// UpdateNickname updates the active session's nickname.
func (c *Client) UpdateNickname(sessionKey, nickname string) (*PublicSession, error) {
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}
	if sessionKey == "" {
		return nil, errors.New("session key is required")
	}

	payload := map[string]string{"nickname": nickname}
	data, _ := json.Marshal(payload)

	path := fmt.Sprintf("/public-sessions/%s/nickname", url.PathEscape(sessionKey))
	req, err := c.newRequest(context.Background(), http.MethodPatch, path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var res SessionResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	if !res.Success || res.Session == nil {
		return nil, errors.New("failed to update nickname")
	}
	return res.Session, nil
}

// RevokePublicSession destroys a session on the server and revokes all associated assets.
func (c *Client) RevokePublicSession(sessionKey string) error {
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}
	if sessionKey == "" {
		return errors.New("session key is required")
	}

	path := fmt.Sprintf("/public-sessions/%s/revoke", url.PathEscape(sessionKey))
	req, err := c.newRequest(context.Background(), http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

// ListFiles retrieves files for a session, folder, or public feed.
func (c *Client) ListFiles(sessionKey, folderId string, publicOnly bool, page, limit int) (*FileListResponse, error) {
	params := url.Values{}
	if publicOnly {
		params.Set("public", "true")
	}
	if sessionKey != "" {
		params.Set("sessionKey", sessionKey)
	} else if c.SessionKey != "" && !publicOnly {
		params.Set("sessionKey", c.SessionKey)
	}
	if folderId != "" {
		params.Set("folderId", folderId)
	}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	path := "/files"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	req, err := c.newRequest(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var res FileListResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetFileInfo retrieves metadata for a specific file by its slug.
func (c *Client) GetFileInfo(slug string) (*FileItem, error) {
	params := url.Values{}
	params.Set("slug", slug)
	if c.SessionKey != "" {
		params.Set("sessionKey", c.SessionKey)
	}

	path := "/files?" + params.Encode()
	req, err := c.newRequest(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var res FileListResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	if len(res.Files) == 0 {
		return nil, fmt.Errorf("file with slug '%s' not found", slug)
	}
	return &res.Files[0], nil
}

// VerifyFilePassword verifies a password for a protected file and returns an access token.
func (c *Client) VerifyFilePassword(slug, password string) (string, error) {
	payload := map[string]string{"password": password}
	data, _ := json.Marshal(payload)

	path := fmt.Sprintf("/files/%s/access", url.PathEscape(slug))
	req, err := c.newRequest(context.Background(), http.MethodPost, path, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	var res VerifyPasswordResponse
	if err := c.doRequest(req, &res); err != nil {
		return "", err
	}
	if !res.Success || res.Token == "" {
		return "", errors.New("invalid file password")
	}
	return res.Token, nil
}

// DeleteFile deletes a file by its ID or slug.
func (c *Client) DeleteFile(idOrSlug, sessionKey string) error {
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}

	slug := idOrSlug
	// If idOrSlug might be an ID or needs slug resolution, try direct slug delete first
	params := url.Values{}
	if sessionKey != "" {
		params.Set("sessionKey", sessionKey)
	}

	path := fmt.Sprintf("/files/%s?%s", url.PathEscape(slug), params.Encode())
	req, err := c.newRequest(context.Background(), http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	var res GenericSuccessResponse
	err = c.doRequest(req, &res)
	if err == nil && res.Success {
		return nil
	}

	// If failed, attempt lookup by slug or ID
	if item, infoErr := c.GetFileInfo(idOrSlug); infoErr == nil && item.Slug != "" && item.Slug != slug {
		path = fmt.Sprintf("/files/%s?%s", url.PathEscape(item.Slug), params.Encode())
		req2, err2 := c.newRequest(context.Background(), http.MethodDelete, path, nil)
		if err2 == nil {
			var res2 GenericSuccessResponse
			if c.doRequest(req2, &res2) == nil && res2.Success {
				return nil
			}
		}
	}

	if err != nil {
		return err
	}
	return errors.New("failed to delete file")
}

// DownloadFile streams a file to outputPath with progress reporting and range resume.
func (c *Client) DownloadFile(idOrSlug, outputPath, token, sessionKey string, onProgress func(written, total int64)) error {
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}

	// Resolve file info first to acquire proper file ID, filename, and total size
	fileInfo, err := c.GetFileInfo(idOrSlug)
	var fileID string
	var expectedFilename string
	var totalSize int64
	if err == nil && fileInfo != nil {
		fileID = fileInfo.ID
		expectedFilename = fileInfo.Filename
		totalSize = fileInfo.GetEffectiveSize()
	} else {
		fileID = idOrSlug
	}

	// Determine final output path
	if outputPath == "" {
		if expectedFilename != "" {
			outputPath = expectedFilename
		} else {
			outputPath = idOrSlug
		}
	} else {
		fileStat, statErr := os.Stat(outputPath)
		if statErr == nil && fileStat.IsDir() {
			if expectedFilename != "" {
				outputPath = filepath.Join(outputPath, expectedFilename)
			} else {
				outputPath = filepath.Join(outputPath, idOrSlug)
			}
		}
	}

	// Check existing file for resume
	var existingSize int64 = 0
	if fi, err := os.Stat(outputPath); err == nil {
		existingSize = fi.Size()
		if totalSize > 0 && existingSize >= totalSize {
			if onProgress != nil {
				onProgress(existingSize, totalSize)
			}
			return nil // Already fully downloaded
		}
	}

	params := url.Values{}
	if token != "" {
		params.Set("fileAccessToken", token)
		params.Set("token", token)
	}
	if sessionKey != "" {
		params.Set("sessionKey", sessionKey)
	}

	downloadPath := fmt.Sprintf("/files/%s", url.PathEscape(fileID))
	if len(params) > 0 {
		downloadPath += "?" + params.Encode()
	}

	client := &http.Client{Timeout: DownloadTimeout}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+downloadPath, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		// Existing file was invalid or server didn't accept range, restart from beginning
		existingSize = 0
		req.Header.Del("Range")
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(body))
	}

	isPartial := resp.StatusCode == http.StatusPartialContent
	var openFlags int
	if isPartial && existingSize > 0 {
		openFlags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	} else {
		openFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		existingSize = 0
	}

	out, err := os.OpenFile(outputPath, openFlags, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer out.Close()

	contentLen := resp.ContentLength
	if contentLen > 0 {
		totalSize = existingSize + contentLen
	}

	pw := &ProgressWriter{
		Total:    totalSize,
		Written:  existingSize,
		Callback: onProgress,
	}

	tee := io.TeeReader(resp.Body, pw)
	_, err = io.Copy(out, tee)
	if err != nil {
		return fmt.Errorf("failed writing downloaded file: %w", err)
	}

	return nil
}

// UploadDirect uploads a file via single multipart/form-data POST /upload.
func (c *Client) UploadDirect(filePath string, opts UploadOptions, onProgress func(written, total int64)) (*FileItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}
	fileSize := stat.Size()

	sessionKey := opts.SessionKey
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}
	deviceId := opts.DeviceId
	if deviceId == "" {
		deviceId = c.DeviceId
	}

	// Prepare multipart form in memory buffer or streaming pipe
	pipeReader, pipeWriter := io.Pipe()
	mpWriter := multipart.NewWriter(pipeWriter)

	// Channel to capture errors during writing
	errChan := make(chan error, 1)

	go func() {
		defer pipeWriter.Close()
		defer mpWriter.Close()

		// IMPORTANT: append BOTH "sessionId" and "sessionKey" with session key value
		if sessionKey != "" {
			_ = mpWriter.WriteField("sessionId", sessionKey)
			_ = mpWriter.WriteField("sessionKey", sessionKey)
		}
		if deviceId != "" {
			_ = mpWriter.WriteField("deviceId", deviceId)
		}
		if opts.IsPublic {
			_ = mpWriter.WriteField("isPublic", "true")
		} else {
			_ = mpWriter.WriteField("isPublic", "false")
		}
		if opts.IsOneTime {
			_ = mpWriter.WriteField("isOneTime", "true")
		}
		if opts.Password != "" {
			_ = mpWriter.WriteField("password", opts.Password)
		}
		if opts.CustomSlug != "" {
			_ = mpWriter.WriteField("customSlug", opts.CustomSlug)
		}
		if opts.CustomFilename != "" {
			_ = mpWriter.WriteField("customFilename", opts.CustomFilename)
		}
		if opts.FolderID != "" {
			_ = mpWriter.WriteField("folderId", opts.FolderID)
		}

		filename := filepath.Base(filePath)
		if opts.CustomFilename != "" {
			filename = opts.CustomFilename
		}

		partWriter, err := mpWriter.CreateFormFile("file", filename)
		if err != nil {
			errChan <- err
			return
		}

		pw := &ProgressWriter{
			Total:    fileSize,
			Written:  0,
			Callback: onProgress,
		}
		tee := io.TeeReader(file, pw)
		if _, err := io.Copy(partWriter, tee); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	client := &http.Client{Timeout: UploadTimeout}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/upload", pipeReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mpWriter.FormDataContentType())
	req.Header.Set("User-Agent", c.UserAgent)
	if sessionKey != "" {
		req.Header.Set("X-Session-Key", sessionKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if writeErr := <-errChan; writeErr != nil {
		return nil, fmt.Errorf("error while piping upload stream: %w", writeErr)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading upload response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Message != "" {
			return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var res UploadResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %w", err)
	}
	if !res.Success || res.File == nil {
		return nil, errors.New("upload response reported failure")
	}

	return res.File, nil
}

// UploadChunked uploads a large file in pieces using the chunks endpoints.
func (c *Client) UploadChunked(filePath string, opts UploadOptions, onProgress func(chunk, totalChunks int)) (*FileItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}
	fileSize := stat.Size()
	chunkSize := DefaultChunkSize
	totalChunks := int((fileSize + int64(chunkSize) - 1) / int64(chunkSize))

	sessionKey := opts.SessionKey
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}
	deviceId := opts.DeviceId
	if deviceId == "" {
		deviceId = c.DeviceId
	}

	filename := filepath.Base(filePath)
	if opts.CustomFilename != "" {
		filename = opts.CustomFilename
	}

	// 1. Initialize Chunk Session
	initPayload := map[string]interface{}{
		"filename":       filename,
		"sizeBytes":      fileSize,
		"mimeType":       "application/octet-stream",
		"chunkSize":      chunkSize,
		"isPublic":       opts.IsPublic,
		"isOneTime":      opts.IsOneTime,
		"sessionId":      sessionKey,
		"deviceId":       deviceId,
	}
	if opts.Password != "" {
		initPayload["password"] = opts.Password
	}
	if opts.CustomSlug != "" {
		initPayload["customSlug"] = opts.CustomSlug
	}
	if opts.FolderID != "" {
		initPayload["folderId"] = opts.FolderID
	}

	initData, _ := json.Marshal(initPayload)
	reqInit, err := c.newRequest(context.Background(), http.MethodPost, "/upload/chunks/init", bytes.NewReader(initData))
	if err != nil {
		return nil, err
	}
	reqInit.Header.Set("Content-Type", "application/json")

	var initRes ChunkInitResponse
	if err := c.doRequest(reqInit, &initRes); err != nil {
		return nil, fmt.Errorf("chunk upload init failed: %w", err)
	}
	if !initRes.Success || initRes.UploadID == "" {
		return nil, errors.New("chunk init returned empty uploadId")
	}

	uploadID := initRes.UploadID
	chunkBuf := make([]byte, chunkSize)

	// 2. Append chunks
	for chunkIndex := 0; chunkIndex < totalChunks; chunkIndex++ {
		bytesRead, rErr := io.ReadFull(file, chunkBuf)
		if rErr != nil && rErr != io.EOF && rErr != io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("failed reading chunk %d: %w", chunkIndex, rErr)
		}
		if bytesRead == 0 {
			break
		}

		// Multipart append
		bodyBuf := &bytes.Buffer{}
		mp := multipart.NewWriter(bodyBuf)
		_ = mp.WriteField("uploadId", uploadID)
		_ = mp.WriteField("chunkIndex", strconv.Itoa(chunkIndex))

		part, err := mp.CreateFormFile("chunk", fmt.Sprintf("chunk_%d", chunkIndex))
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(chunkBuf[:bytesRead]); err != nil {
			return nil, err
		}
		_ = mp.Close()

		reqAppend, err := c.newRequest(context.Background(), http.MethodPost, "/upload/chunks/append", bodyBuf)
		if err != nil {
			return nil, err
		}
		reqAppend.Header.Set("Content-Type", mp.FormDataContentType())

		var appendRes GenericSuccessResponse
		if err := c.doRequest(reqAppend, &appendRes); err != nil {
			return nil, fmt.Errorf("failed uploading chunk %d: %w", chunkIndex+1, err)
		}

		if onProgress != nil {
			onProgress(chunkIndex+1, totalChunks)
		}
	}

	// 3. Finalize upload
	finalPayload := map[string]string{"uploadId": uploadID}
	finalData, _ := json.Marshal(finalPayload)
	reqFinal, err := c.newRequest(context.Background(), http.MethodPost, "/upload/chunks/finalize", bytes.NewReader(finalData))
	if err != nil {
		return nil, err
	}
	reqFinal.Header.Set("Content-Type", "application/json")

	var finalRes UploadResponse
	if err := c.doRequest(reqFinal, &finalRes); err != nil {
		return nil, fmt.Errorf("chunk finalize failed: %w", err)
	}
	if !finalRes.Success || finalRes.File == nil {
		return nil, errors.New("chunk finalize failed")
	}

	return finalRes.File, nil
}

// Upload chooses direct or chunked upload based on file size or options.
func (c *Client) Upload(filePath string, opts UploadOptions, onProgressDirect func(written, total int64), onProgressChunked func(chunk, totalChunks int)) (*FileItem, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if opts.ForceChunked || fi.Size() > LargeFileThreshold {
		return c.UploadChunked(filePath, opts, onProgressChunked)
	}
	return c.UploadDirect(filePath, opts, onProgressDirect)
}

// CreateFolder creates a folder.
func (c *Client) CreateFolder(name, slug, sessionKey string, isPublic bool) (*FolderItem, error) {
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}

	payload := map[string]interface{}{
		"name":     name,
		"isPublic": isPublic,
	}
	if slug != "" {
		payload["slug"] = slug
	}
	if sessionKey != "" {
		payload["sessionKey"] = sessionKey
	}

	data, _ := json.Marshal(payload)
	req, err := c.newRequest(context.Background(), http.MethodPost, "/folders", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var res FolderResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	if !res.Success || res.Folder == nil {
		return nil, errors.New("failed to create folder")
	}
	return res.Folder, nil
}

// ListFolders lists folders for the active session.
func (c *Client) ListFolders(sessionKey string) ([]FolderItem, error) {
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}

	params := url.Values{}
	if sessionKey != "" {
		params.Set("sessionKey", sessionKey)
	}

	path := "/folders"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	req, err := c.newRequest(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var res FolderListResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	return res.Folders, nil
}

// DeleteFolder deletes a folder by ID.
func (c *Client) DeleteFolder(id, sessionKey string) error {
	if sessionKey == "" {
		sessionKey = c.SessionKey
	}

	params := url.Values{}
	if sessionKey != "" {
		params.Set("sessionKey", sessionKey)
	}

	path := fmt.Sprintf("/folders/%s?%s", url.PathEscape(id), params.Encode())
	req, err := c.newRequest(context.Background(), http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	var res GenericSuccessResponse
	if err := c.doRequest(req, &res); err != nil {
		return err
	}
	if !res.Success {
		return errors.New("failed to delete folder")
	}
	return nil
}

// DownloadFolderZip streams the ZIP archive of a folder to outputPath.
func (c *Client) DownloadFolderZip(folderIdOrSlug, outputPath string, onProgress func(written int64)) error {
	params := url.Values{}
	if c.SessionKey != "" {
		params.Set("sessionKey", c.SessionKey)
	}

	// Try /folders/:id/download-zip first, fallback to /public-folders/:token/download-zip
	downloadURL := fmt.Sprintf("%s/folders/%s/download-zip?%s", c.BaseURL, url.PathEscape(folderIdOrSlug), params.Encode())

	client := &http.Client{Timeout: DownloadTimeout}
	resp, err := client.Get(downloadURL)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		if resp != nil {
			resp.Body.Close()
		}
		publicURL := fmt.Sprintf("%s/public-folders/%s/download-zip?%s", c.BaseURL, url.PathEscape(folderIdOrSlug), params.Encode())
		resp, err = client.Get(publicURL)
		if err != nil {
			return fmt.Errorf("folder zip request failed: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("folder zip download failed (%d): %s", resp.StatusCode, string(body))
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf("folder-%s.zip", folderIdOrSlug)
	}

	out, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed opening output zip file: %w", err)
	}
	defer out.Close()

	pw := &ProgressWriter{
		Callback: func(written, total int64) {
			if onProgress != nil {
				onProgress(written)
			}
		},
	}

	tee := io.TeeReader(resp.Body, pw)
	if _, err := io.Copy(out, tee); err != nil {
		return fmt.Errorf("error writing zip file: %w", err)
	}

	return nil
}
