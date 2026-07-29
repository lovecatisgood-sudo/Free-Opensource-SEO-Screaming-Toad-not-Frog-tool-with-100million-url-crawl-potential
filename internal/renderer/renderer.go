package renderer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

const (
	protocolVersion          = 1
	maximumFrameBytes        = 8 << 20
	maximumResourceBytes     = 5 << 20
	maximumRenderedHTMLBytes = 4 << 20
	maximumProtocolFrames    = 4_100
)

type Fetcher interface {
	Fetch(context.Context, string) (fetchpolicy.FetchResult, error)
}

// Supervisor starts the trusted renderer worker and mediates every browser
// resource request through the guarded Go fetcher. The worker never receives
// authority to connect to the network directly.
type Supervisor struct {
	NodeBinary       string
	ScriptPath       string
	BrowserPath      string
	ContainerSandbox bool
	Fetcher          Fetcher
}

type Request struct {
	RequestID       string
	URL             string
	Deadline        time.Duration
	MaximumRequests int
	MaximumBytes    int64
}

type Result struct {
	Status           string
	HTML             string
	FinalURL         string
	ErrorCode        string
	RequestCount     int
	TransferredBytes int64
}

type wireRenderRequest struct {
	Kind            string `json:"kind"`
	ProtocolVersion int    `json:"protocolVersion"`
	RequestID       string `json:"requestId"`
	URL             string `json:"url"`
	DeadlineMS      int64  `json:"deadlineMs"`
	MaximumRequests int    `json:"maximumRequests"`
	MaximumBytes    int64  `json:"maximumBytes"`
}

type wireMessage struct {
	Kind             string            `json:"kind"`
	ProtocolVersion  int               `json:"protocolVersion"`
	RequestID        string            `json:"requestId"`
	FetchID          string            `json:"fetchId"`
	URL              string            `json:"url"`
	ResourceType     string            `json:"resourceType"`
	Status           string            `json:"status"`
	StatusCode       int               `json:"statusCode,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	BodyBase64       string            `json:"bodyBase64,omitempty"`
	HTML             string            `json:"html,omitempty"`
	FinalURL         string            `json:"finalURL,omitempty"`
	RequestCount     int               `json:"requestCount,omitempty"`
	TransferredBytes int64             `json:"transferredBytes,omitempty"`
	ErrorCode        string            `json:"errorCode,omitempty"`
}

func (s *Supervisor) Render(ctx context.Context, request Request) (Result, error) {
	if err := s.validate(request); err != nil {
		return Result{}, err
	}
	script, err := filepath.Abs(s.ScriptPath)
	if err != nil {
		return Result{}, err
	}
	info, err := os.Stat(script)
	if err != nil || !info.Mode().IsRegular() {
		return Result{}, errors.New("renderer worker script is unavailable")
	}
	node := s.NodeBinary
	if node == "" {
		node = "node"
	}
	node, err = exec.LookPath(node)
	if err != nil {
		return Result{}, errors.New("renderer Node.js runtime is unavailable")
	}

	runCtx, cancel := context.WithTimeout(ctx, request.Deadline+5*time.Second)
	defer cancel()
	command := exec.CommandContext(runCtx, node, script)
	command.Dir = filepath.Dir(script)
	command.Env = []string{"PATH=" + os.Getenv("PATH"), "NODE_ENV=production"}
	if s.BrowserPath != "" {
		command.Env = append(command.Env, "PLAYWRIGHT_BROWSERS_PATH="+s.BrowserPath)
	}
	if s.ContainerSandbox {
		command.Env = append(command.Env, "SEO_AUDITOR_CONTAINER_SANDBOX=1")
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		return Result{}, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	stderr := &limitedBuffer{maximum: 32 << 10}
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		return Result{}, err
	}
	waited := false
	defer func() {
		if !waited && command.Process != nil {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	}()

	writer := bufio.NewWriter(stdin)
	reader := bufio.NewReader(stdout)
	initial := wireRenderRequest{
		Kind: "render_request", ProtocolVersion: protocolVersion,
		RequestID: request.RequestID, URL: request.URL,
		DeadlineMS:      request.Deadline.Milliseconds(),
		MaximumRequests: request.MaximumRequests, MaximumBytes: request.MaximumBytes,
	}
	if err := writeFrame(writer, initial); err != nil {
		return Result{}, err
	}

	requests := 0
	var transferred int64
	for frames := 0; frames < maximumProtocolFrames; frames++ {
		var message wireMessage
		if err := readFrame(reader, &message); err != nil {
			return Result{}, fmt.Errorf("renderer protocol: %w", err)
		}
		if message.ProtocolVersion != protocolVersion || message.RequestID != request.RequestID {
			return Result{}, errors.New("renderer protocol identity mismatch")
		}
		switch message.Kind {
		case "fetch_resource":
			requests++
			response := wireMessage{
				Kind: "fetch_resource_result", ProtocolVersion: protocolVersion,
				RequestID: request.RequestID, FetchID: message.FetchID,
				Status: "blocked", ErrorCode: "target_blocked",
			}
			if requests <= request.MaximumRequests && message.FetchID != "" && len(message.URL) <= 8192 {
				fetched, fetchErr := s.Fetcher.Fetch(runCtx, message.URL)
				size := int64(len(fetched.Body))
				switch {
				case fetchErr != nil:
					response.ErrorCode = "fetch_rejected"
				case size > maximumResourceBytes:
					response.ErrorCode = "resource_byte_limit"
				case transferred+size > request.MaximumBytes:
					response.ErrorCode = "render_byte_limit"
				default:
					transferred += size
					response.Status = "completed"
					response.StatusCode = fetched.StatusCode
					response.Headers = selectedHeaders(fetched.Header)
					response.BodyBase64 = base64.StdEncoding.EncodeToString(fetched.Body)
					response.ErrorCode = ""
				}
			}
			if err := writeFrame(writer, response); err != nil {
				return Result{}, err
			}
		case "render_result":
			if message.Status != "completed" && message.Status != "blocked" && message.Status != "failed" {
				return Result{}, errors.New("renderer returned an invalid status")
			}
			if len(message.HTML) > maximumRenderedHTMLBytes {
				return Result{}, errors.New("rendered HTML exceeds limit")
			}
			if message.RequestCount < 0 || message.RequestCount > request.MaximumRequests || message.TransferredBytes < 0 || message.TransferredBytes > transferred {
				return Result{}, errors.New("renderer returned invalid accounting")
			}
			_ = stdin.Close()
			waitErr := command.Wait()
			waited = true
			if waitErr != nil && message.Status == "completed" {
				return Result{}, fmt.Errorf("renderer exited unsuccessfully: %s", stderr.String())
			}
			return Result{
				Status: message.Status, HTML: message.HTML, FinalURL: message.FinalURL,
				ErrorCode: message.ErrorCode, RequestCount: message.RequestCount,
				TransferredBytes: message.TransferredBytes,
			}, nil
		default:
			return Result{}, errors.New("renderer sent an unsupported message")
		}
	}
	return Result{}, errors.New("renderer protocol frame limit reached")
}

func (s *Supervisor) validate(request Request) error {
	if s.Fetcher == nil {
		return errors.New("renderer fetcher is required")
	}
	if request.RequestID == "" || len(request.RequestID) > 200 || request.URL == "" || len(request.URL) > 8192 ||
		request.Deadline < 100*time.Millisecond || request.Deadline > 2*time.Minute ||
		request.MaximumRequests < 1 || request.MaximumRequests > 2_000 ||
		request.MaximumBytes < 1_024 || request.MaximumBytes > 100<<20 {
		return errors.New("render request is outside supported budgets")
	}
	return nil
}

func selectedHeaders(headers http.Header) map[string]string {
	result := map[string]string{}
	for _, name := range []string{"Content-Type", "Cache-Control", "Content-Language", "X-Robots-Tag"} {
		if value := headers.Get(name); value != "" {
			result[strings.ToLower(name)] = value
		}
	}
	return result
}

func writeFrame(writer *bufio.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(payload) > maximumFrameBytes {
		return errors.New("renderer frame exceeds limit")
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if _, err := writer.Write(header[:]); err != nil {
		return err
	}
	if _, err := writer.Write(payload); err != nil {
		return err
	}
	return writer.Flush()
}

func readFrame(reader *bufio.Reader, value any) error {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size == 0 || size > maximumFrameBytes {
		return errors.New("renderer frame has an invalid size")
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("renderer frame contains trailing JSON")
	}
	return nil
}

type limitedBuffer struct {
	mu      sync.Mutex
	value   []byte
	maximum int
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	remaining := b.maximum - len(b.value)
	if remaining > 0 {
		b.value = append(b.value, value[:min(remaining, len(value))]...)
	}
	return len(value), nil
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.value)
}
