package source

import (
	"crypto/sha1"
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Candidate struct {
	ID         string    `json:"id"`
	SourcePath string    `json:"-"`
	Title      string    `json:"title"`
	SourceName string    `json:"sourceName"`
	TakenAt    time.Time `json:"takenAt,omitempty"`
	MediaType  string    `json:"mediaType"`
}

type Provider interface {
	Candidates() ([]Candidate, error)
}

type LocalFolderProvider struct {
	Root string
}

func (p LocalFolderProvider) Candidates() ([]Candidate, error) {
	var out []Candidate
	root, err := filepath.Abs(p.Root)
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !isPhoto(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out = append(out, Candidate{
			ID:         stableID(rel),
			SourcePath: path,
			Title:      strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			SourceName: "Local folder",
			TakenAt:    info.ModTime(),
			MediaType:  mediaType(path),
		})
		return nil
	})
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, err
}

func stableID(value string) string {
	sum := sha1.Sum([]byte(filepath.ToSlash(strings.ToLower(value))))
	return hex.EncodeToString(sum[:])[:16]
}

func isPhoto(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg":
		return true
	default:
		return false
	}
}

func mediaType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
