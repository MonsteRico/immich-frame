package immich

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	defaultLimit   = 100
)

type ErrorKind string

const (
	ErrorInvalidURL   ErrorKind = "invalid_url"
	ErrorInvalidKey   ErrorKind = "invalid_key"
	ErrorNetwork      ErrorKind = "network"
	ErrorPermission   ErrorKind = "permission"
	ErrorUnavailable  ErrorKind = "unavailable"
	ErrorIncompatible ErrorKind = "incompatible_response"
	ErrorInvalidInput ErrorKind = "invalid_input"
)

type APIError struct {
	Kind       ErrorKind
	Message    string
	StatusCode int
}

func (e *APIError) Error() string {
	return e.Message
}

type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func NewClient(baseURL, apiKey string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, userError(ErrorInvalidURL, "Immich URL is required", 0)
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, userError(ErrorInvalidURL, "Immich URL must be a valid http or https URL", 0)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, userError(ErrorInvalidURL, "Immich URL must use http or https", 0)
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, userError(ErrorInvalidKey, "Immich API key is required", 0)
	}
	c := &Client{
		baseURL: parsed,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

type ConnectionInfo struct {
	Version string
	KeyName string
}

func (c *Client) TestConnection(ctx context.Context) (ConnectionInfo, error) {
	var version struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
		Patch int `json:"patch"`
	}
	if err := c.getJSON(ctx, "/server/version", &version); err != nil {
		return ConnectionInfo{}, err
	}
	if version.Major == 0 && version.Minor == 0 && version.Patch == 0 {
		return ConnectionInfo{}, userError(ErrorIncompatible, "Immich server returned an incompatible version response", 0)
	}
	var key struct {
		Name string `json:"name"`
	}
	if err := c.getJSON(ctx, "/api-keys/me", &key); err != nil {
		return ConnectionInfo{}, err
	}
	return ConnectionInfo{
		Version: fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch),
		KeyName: key.Name,
	}, nil
}

type Album struct {
	ID         string
	Name       string
	AssetCount int
}

func (c *Client) ListAlbums(ctx context.Context) ([]Album, error) {
	var raw []albumResponse
	if err := c.getJSON(ctx, "/albums", &raw); err != nil {
		return nil, err
	}
	albums := make([]Album, 0, len(raw))
	for _, album := range raw {
		if strings.TrimSpace(album.ID) == "" {
			continue
		}
		albums = append(albums, Album{
			ID:         album.ID,
			Name:       album.AlbumName,
			AssetCount: album.AssetCount,
		})
	}
	sort.Slice(albums, func(i, j int) bool {
		if albums[i].Name == albums[j].Name {
			return albums[i].ID < albums[j].ID
		}
		return albums[i].Name < albums[j].Name
	})
	return albums, nil
}

func (c *Client) ListAlbumCandidates(ctx context.Context, albumID string) ([]AssetCandidate, error) {
	albumID = strings.TrimSpace(albumID)
	if albumID == "" {
		return nil, userError(ErrorInvalidInput, "Immich album id is required", 0)
	}
	var raw albumResponse
	if err := c.getJSON(ctx, "/albums/"+url.PathEscape(albumID), &raw); err != nil {
		return nil, err
	}
	return normalizeAssets(raw.Assets, raw.AlbumName), nil
}

func (c *Client) ListRandomCandidates(ctx context.Context, limit int) ([]AssetCandidate, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	body := map[string]any{
		"size":        limit,
		"type":        "IMAGE",
		"visibility":  "timeline",
		"withDeleted": false,
		"withExif":    true,
		"withPeople":  false,
		"withStacked": false,
	}
	var raw []assetResponse
	if err := c.postJSON(ctx, "/search/random", body, &raw); err != nil {
		return nil, err
	}
	return normalizeAssets(raw, "Immich library"), nil
}

type Target struct {
	Width  int
	Height int
	Format string
}

type Rendition struct {
	AssetID     string
	Identity    string
	ContentType string
	Body        io.ReadCloser
}

func (c *Client) FetchRendition(ctx context.Context, assetID string, target Target) (Rendition, error) {
	assetID = strings.TrimSpace(assetID)
	if assetID == "" {
		return Rendition{}, userError(ErrorInvalidInput, "Immich asset id is required", 0)
	}
	format := strings.ToUpper(strings.TrimSpace(target.Format))
	if format == "" {
		format = "WEBP"
	}
	if format != "WEBP" && format != "JPEG" {
		format = "WEBP"
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/assets/"+url.PathEscape(assetID)+"/thumbnail", nil)
	if err != nil {
		return Rendition{}, err
	}
	q := req.URL.Query()
	q.Set("format", format)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/octet-stream")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Rendition{}, userError(ErrorNetwork, "Unable to reach Immich server", 0)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return Rendition{}, c.statusError(resp)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension("." + strings.ToLower(format))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return Rendition{
		AssetID:     assetID,
		Identity:    "thumbnail-" + strings.ToLower(format),
		ContentType: contentType,
		Body:        resp.Body,
	}, nil
}

type AssetCandidate struct {
	ID                string
	RenditionIdentity string
	Title             string
	SourceName        string
	TakenAt           time.Time
	Width             int
	Height            int
	Orientation       string
	MediaType         string
}

type albumResponse struct {
	ID         string          `json:"id"`
	AlbumName  string          `json:"albumName"`
	AssetCount int             `json:"assetCount"`
	Assets     []assetResponse `json:"assets"`
}

type assetResponse struct {
	ID               string       `json:"id"`
	Type             string       `json:"type"`
	OriginalFileName string       `json:"originalFileName"`
	OriginalMimeType string       `json:"originalMimeType"`
	FileCreatedAt    string       `json:"fileCreatedAt"`
	LocalDateTime    string       `json:"localDateTime"`
	IsArchived       bool         `json:"isArchived"`
	IsTrashed        bool         `json:"isTrashed"`
	Visibility       string       `json:"visibility"`
	ExifInfo         exifResponse `json:"exifInfo"`
}

type exifResponse struct {
	DateTimeOriginal string `json:"dateTimeOriginal"`
	ExifImageWidth   int    `json:"exifImageWidth"`
	ExifImageHeight  int    `json:"exifImageHeight"`
	Orientation      string `json:"orientation"`
}

func normalizeAssets(raw []assetResponse, sourceName string) []AssetCandidate {
	out := make([]AssetCandidate, 0, len(raw))
	for _, asset := range raw {
		candidate, ok := normalizeAsset(asset, sourceName)
		if ok {
			out = append(out, candidate)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TakenAt.Equal(out[j].TakenAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].TakenAt.Before(out[j].TakenAt)
	})
	return out
}

func normalizeAsset(asset assetResponse, sourceName string) (AssetCandidate, bool) {
	if strings.TrimSpace(asset.ID) == "" || asset.Type != "IMAGE" || asset.IsArchived || asset.IsTrashed {
		return AssetCandidate{}, false
	}
	if asset.Visibility != "" && asset.Visibility != "timeline" {
		return AssetCandidate{}, false
	}
	takenAt := firstTime(asset.ExifInfo.DateTimeOriginal, asset.LocalDateTime, asset.FileCreatedAt)
	title := strings.TrimSpace(asset.OriginalFileName)
	if title == "" {
		title = asset.ID
	}
	width := asset.ExifInfo.ExifImageWidth
	height := asset.ExifInfo.ExifImageHeight
	return AssetCandidate{
		ID:                asset.ID,
		RenditionIdentity: "thumbnail-webp",
		Title:             title,
		SourceName:        sourceName,
		TakenAt:           takenAt,
		Width:             width,
		Height:            height,
		Orientation:       normalizeOrientation(asset.ExifInfo.Orientation, width, height),
		MediaType:         nonEmpty(asset.OriginalMimeType, "image/*"),
	}, true
}

func firstTime(values ...string) time.Time {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
			return t
		}
	}
	return time.Time{}
}

func normalizeOrientation(value string, width, height int) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "portrait" || value == "landscape" || value == "square" {
		return value
	}
	if width > 0 && height > 0 {
		if width == height {
			return "square"
		}
		if width > height {
			return "landscape"
		}
		return "portrait"
	}
	return ""
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (c *Client) getJSON(ctx context.Context, endpoint string, out any) error {
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	return c.doJSON(req, out)
}

func (c *Client) postJSON(ctx context.Context, endpoint string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return c.doJSON(req, out)
}

func (c *Client) doJSON(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return userError(ErrorNetwork, "Unable to reach Immich server", 0)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.statusError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return userError(ErrorIncompatible, "Immich server returned an incompatible response", 0)
	}
	return nil
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	u := *c.baseURL
	basePath := strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(basePath, "/api") {
		basePath = path.Join(basePath, "api")
	}
	u.Path = path.Join(basePath, strings.TrimPrefix(endpoint, "/"))
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	return req, nil
}

func (c *Client) statusError(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return userError(ErrorInvalidKey, "Immich API key was rejected", resp.StatusCode)
	case http.StatusForbidden:
		return userError(ErrorPermission, "Immich API key does not have the required permission", resp.StatusCode)
	case http.StatusNotFound:
		return userError(ErrorUnavailable, "Immich endpoint or asset was not found", resp.StatusCode)
	default:
		if resp.StatusCode >= 500 {
			return userError(ErrorUnavailable, "Immich server is unavailable", resp.StatusCode)
		}
		return userError(ErrorIncompatible, "Immich server returned an unexpected status "+strconv.Itoa(resp.StatusCode), resp.StatusCode)
	}
}

func userError(kind ErrorKind, message string, status int) *APIError {
	return &APIError{Kind: kind, Message: message, StatusCode: status}
}

func IsKind(err error, kind ErrorKind) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Kind == kind
}
