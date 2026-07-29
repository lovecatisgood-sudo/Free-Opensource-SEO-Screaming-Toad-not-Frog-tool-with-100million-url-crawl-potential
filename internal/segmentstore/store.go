package segmentstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var identityPattern = regexp.MustCompile(`^[a-z0-9_]{1,160}$`)

type Store struct{ root string }

type Commit struct {
	ObjectKey string `json:"object_key"`
	Checksum  string `json:"checksum"`
	SizeBytes int64  `json:"size_bytes"`
	Inserted  bool   `json:"inserted"`
}

func Open(root string) (*Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(abs, "segments"), 0o700); err != nil {
		return nil, err
	}
	return &Store{root: abs}, nil
}

// Commit writes one immutable result segment to a managed key. Callers cannot
// choose a path, and bytes beyond the explicit ceiling are rejected.
func (s *Store) Commit(ctx context.Context, crawlID string, sequence, maximumBytes int64, source io.Reader) (Commit, error) {
	if !identityPattern.MatchString(crawlID) || sequence < 0 || maximumBytes < 1 || source == nil {
		return Commit{}, errors.New("valid crawl, sequence, byte ceiling and source are required")
	}
	directory := filepath.Join(s.root, "segments", crawlID)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return Commit{}, err
	}
	name := fmt.Sprintf("%012d.segment", sequence)
	destination := filepath.Join(directory, name)
	temporary, err := os.CreateTemp(directory, ".segment-*")
	if err != nil {
		return Commit{}, err
	}
	temporaryName := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryName)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return Commit{}, err
	}
	hash := sha256.New()
	reader := &contextReader{ctx: ctx, reader: io.LimitReader(source, maximumBytes+1)}
	written, err := io.Copy(io.MultiWriter(temporary, hash), reader)
	if err != nil {
		return Commit{}, err
	}
	if written > maximumBytes {
		return Commit{}, errors.New("result segment exceeds byte ceiling")
	}
	if err := temporary.Sync(); err != nil {
		return Commit{}, err
	}
	if err := temporary.Close(); err != nil {
		return Commit{}, err
	}
	checksum := hex.EncodeToString(hash.Sum(nil))
	objectKey := filepath.ToSlash(filepath.Join("segments", crawlID, name))
	if existing, err := checksumFile(destination); err == nil {
		if existing != checksum {
			return Commit{}, errors.New("immutable result segment already exists with different content")
		}
		return Commit{ObjectKey: objectKey, Checksum: checksum, SizeBytes: written, Inserted: false}, nil
	} else if !os.IsNotExist(err) {
		return Commit{}, err
	}
	if err := os.Link(temporaryName, destination); err != nil {
		if existing, readErr := checksumFile(destination); readErr == nil {
			if existing != checksum {
				return Commit{}, errors.New("immutable result segment already exists with different content")
			}
			return Commit{ObjectKey: objectKey, Checksum: checksum, SizeBytes: written, Inserted: false}, nil
		}
		return Commit{}, err
	}
	if err := os.Remove(temporaryName); err != nil {
		return Commit{}, err
	}
	if directoryHandle, err := os.Open(directory); err == nil {
		_ = directoryHandle.Sync()
		_ = directoryHandle.Close()
	}
	committed = true
	return Commit{ObjectKey: objectKey, Checksum: checksum, SizeBytes: written, Inserted: true}, nil
}

func (s *Store) Verify(objectKey, expected string) error {
	clean := filepath.Clean(filepath.FromSlash(objectKey))
	if clean != filepath.FromSlash(objectKey) || filepath.IsAbs(clean) || len(clean) < len("segments/x") {
		return errors.New("object key is invalid")
	}
	path := filepath.Join(s.root, clean)
	relative, err := filepath.Rel(s.root, path)
	if err != nil || relative == ".." || len(relative) >= 3 && relative[:3] == ".."+string(filepath.Separator) {
		return errors.New("object key escapes segment store")
	}
	actual, err := checksumFile(path)
	if err != nil {
		return err
	}
	if actual != expected {
		return errors.New("result segment checksum mismatch")
	}
	return nil
}

func checksumFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(buffer)
	}
}
