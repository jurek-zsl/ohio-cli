package api

import "time"

// PublicSession represents a public OhioFiles user session.
type PublicSession struct {
	ID                string     `json:"id"`
	Key               string     `json:"key"`
	Nickname          string     `json:"nickname"`
	ProfileSlug       string     `json:"profileSlug"`
	NicknameChangedAt *time.Time `json:"nicknameChangedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	LastActive        time.Time  `json:"lastActive"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
	FileCount         int        `json:"fileCount,omitempty"`
	UploadCount       int        `json:"uploadCount,omitempty"`
	Files             []FileItem `json:"files,omitempty"`
}

// SessionResponse is the wrapper returned by public session endpoints.
type SessionResponse struct {
	Success  bool            `json:"success"`
	Session  *PublicSession  `json:"session,omitempty"`
	Sessions []PublicSession `json:"sessions,omitempty"`
	Error    string          `json:"error,omitempty"`
	Message  string          `json:"message,omitempty"`
}

// FileItem represents a file in OhioFiles.
type FileItem struct {
	ID                       string     `json:"id"`
	Slug                     string     `json:"slug"`
	Filename                 string     `json:"filename"`
	Size                     int64      `json:"size"`
	SizeBytes                int64      `json:"sizeBytes,omitempty"`
	MimeType                 string     `json:"mimeType"`
	UploadedAt               time.Time  `json:"uploadedAt"`
	CreatedAt                time.Time  `json:"createdAt,omitempty"`
	IsPublic                 bool       `json:"isPublic"`
	IsOneTime                bool       `json:"isOneTime"`
	OneTimeConsumed          bool       `json:"oneTimeConsumed,omitempty"`
	PasswordRequired         bool       `json:"passwordRequired"`
	PasswordUnlocked         bool       `json:"passwordUnlocked"`
	ChecksumSHA256           *string    `json:"checksumSha256,omitempty"`
	Downloads                int        `json:"downloads"`
	FolderID                 *string    `json:"folderId,omitempty"`
	FolderName               *string    `json:"folderName,omitempty"`
	FolderPath               *string    `json:"folderPath,omitempty"`
	IsOwner                  bool       `json:"isOwner,omitempty"`
	IsSaved                  bool       `json:"isSaved,omitempty"`
	ReadOnly                 bool       `json:"readOnly,omitempty"`
	PublicSessionKey         *string    `json:"publicSessionKey,omitempty"`
	PublicSessionNickname    *string    `json:"publicSessionNickname,omitempty"`
	PublicSessionProfileSlug *string    `json:"publicSessionProfileSlug,omitempty"`
	PublicSessionAvatarURL   *string    `json:"publicSessionAvatarUrl,omitempty"`
	URL                      string     `json:"url"`
}

// GetEffectiveSize returns either Size or SizeBytes depending on which is populated.
func (f *FileItem) GetEffectiveSize() int64 {
	if f.Size > 0 {
		return f.Size
	}
	return f.SizeBytes
}

// FileListResponse is returned by GET /files.
type FileListResponse struct {
	Success    bool       `json:"success"`
	Files      []FileItem `json:"files"`
	Total      int        `json:"total,omitempty"`
	Page       int        `json:"page,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	TotalPages int        `json:"totalPages,omitempty"`
	HasMore    bool       `json:"hasMore,omitempty"`
	Error      string     `json:"error,omitempty"`
	Message    string     `json:"message,omitempty"`
}

// FolderItem represents a folder in OhioFiles.
type FolderItem struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	ParentID         *string   `json:"parentId,omitempty"`
	IsPublic         bool      `json:"isPublic"`
	DefaultAccess    string    `json:"defaultAccess,omitempty"`
	ShortCode        *string   `json:"shortCode,omitempty"`
	PasswordRequired bool      `json:"passwordRequired"`
	IsShared         bool      `json:"isShared,omitempty"`
	ShareAccess      *string   `json:"shareAccess,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// FolderListResponse is returned by GET /folders.
type FolderListResponse struct {
	Success       bool         `json:"success"`
	CurrentFolder *FolderItem  `json:"currentFolder,omitempty"`
	Folders       []FolderItem `json:"folders"`
	Error         string       `json:"error,omitempty"`
	Message       string       `json:"message,omitempty"`
}

// FolderResponse is returned when creating or fetching a single folder.
type FolderResponse struct {
	Success bool        `json:"success"`
	Folder  *FolderItem `json:"folder,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// UploadResponse is returned upon completing a file upload.
type UploadResponse struct {
	Success bool      `json:"success"`
	File    *FileItem `json:"file,omitempty"`
	Error   string    `json:"error,omitempty"`
	Message string    `json:"message,omitempty"`
}

// ChunkInitResponse is returned by POST /upload/chunks/init.
type ChunkInitResponse struct {
	Success     bool   `json:"success"`
	UploadID    string `json:"uploadId"`
	ChunkSize   int    `json:"chunkSize"`
	TotalChunks int    `json:"totalChunks"`
	Slug        string `json:"slug"`
	IsPublic    bool   `json:"isPublic"`
	IsOneTime   bool   `json:"isOneTime"`
	Error       string `json:"error,omitempty"`
	Message     string `json:"message,omitempty"`
}

// VerifyPasswordResponse is returned by POST /files/:slug/access.
type VerifyPasswordResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// GenericSuccessResponse is returned for simple success/error payloads.
type GenericSuccessResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message,omitempty"`
	RetryAfter int    `json:"retryAfter,omitempty"`
}

// UploadOptions provides parameters for file uploads.
type UploadOptions struct {
	IsPublic       bool
	IsOneTime      bool
	Password       string
	CustomSlug     string
	CustomFilename string
	FolderID       string
	SessionKey     string
	DeviceId       string
	ForceChunked   bool
}
