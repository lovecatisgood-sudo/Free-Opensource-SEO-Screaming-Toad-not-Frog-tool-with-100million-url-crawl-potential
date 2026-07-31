package renderer

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

type fixtureFetcher struct {
	blocked bool
	seen    string
}

func (f *fixtureFetcher) Fetch(_ context.Context, raw string) (fetchpolicy.FetchResult, error) {
	f.seen = raw
	if f.blocked {
		return fetchpolicy.FetchResult{}, errors.New("blocked")
	}
	return fetchpolicy.FetchResult{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/html"}}, Body: []byte("<html><title>raw</title></html>")}, nil
}
func TestSupervisorMediatesWorkerResourceFetch(t *testing.T) {
	t.Parallel()
	script := filepath.Join(t.TempDir(), "worker.mjs")
	source := `import{stdin,stdout}from'node:process';let b=Buffer.alloc(0);function send(v){let p=Buffer.from(JSON.stringify(v));let h=Buffer.alloc(4);h.writeUInt32BE(p.length);stdout.write(Buffer.concat([h,p]));}stdin.on('data',c=>{b=Buffer.concat([b,c]);while(b.length>=4&&b.length>=4+b.readUInt32BE()){let n=b.readUInt32BE();let m=JSON.parse(b.subarray(4,4+n));b=b.subarray(4+n);if(m.kind==='render_request')send({kind:'fetch_resource',protocolVersion:1,requestId:m.requestId,fetchId:'f1',url:m.url,resourceType:'document'});else if(m.kind==='fetch_resource_result')send({kind:'render_result',protocolVersion:1,requestId:m.requestId,status:m.status,html:m.status==='completed'?'<html><title>rendered</title></html>':'',finalURL:'https://example.com/',requestCount:1,transferredBytes:31,errorCode:m.errorCode});}});`
	if err := os.WriteFile(script, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	fetcher := &fixtureFetcher{}
	supervisor := Supervisor{NodeBinary: "node", ScriptPath: script, Fetcher: fetcher}
	result, err := supervisor.Render(context.Background(), Request{RequestID: "test", URL: "https://example.com/", Deadline: 5 * time.Second, MaximumRequests: 10, MaximumBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || !strings.Contains(result.HTML, "rendered") || fetcher.seen != "https://example.com/" {
		t.Fatalf("result=%+v seen=%q", result, fetcher.seen)
	}
}
func TestSupervisorReturnsBlockedWhenGuardedFetcherRejects(t *testing.T) {
	t.Parallel()
	script := filepath.Join(t.TempDir(), "worker.mjs")
	source := `import{stdin,stdout}from'node:process';let b=Buffer.alloc(0);function s(v){let p=Buffer.from(JSON.stringify(v)),h=Buffer.alloc(4);h.writeUInt32BE(p.length);stdout.write(Buffer.concat([h,p]));}stdin.on('data',c=>{b=Buffer.concat([b,c]);if(b.length<4||b.length<4+b.readUInt32BE())return;let n=b.readUInt32BE(),m=JSON.parse(b.subarray(4,4+n));b=b.subarray(4+n);if(m.kind==='render_request')s({kind:'fetch_resource',protocolVersion:1,requestId:m.requestId,fetchId:'f1',url:'http://127.0.0.1/private',resourceType:'document'});else s({kind:'render_result',protocolVersion:1,requestId:m.requestId,status:m.status,requestCount:1,transferredBytes:0,errorCode:m.errorCode});});`
	if err := os.WriteFile(script, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := (&Supervisor{NodeBinary: "node", ScriptPath: script, Fetcher: &fixtureFetcher{blocked: true}}).Render(context.Background(), Request{RequestID: "blocked", URL: "https://example.com/", Deadline: 5 * time.Second, MaximumRequests: 10, MaximumBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "blocked" || result.ErrorCode != "fetch_rejected" {
		t.Fatalf("result=%+v", result)
	}
}

type browserFixtureFetcher struct {
	seen []string
}

func (f *browserFixtureFetcher) Fetch(_ context.Context, raw string) (fetchpolicy.FetchResult, error) {
	f.seen = append(f.seen, raw)
	header := http.Header{}
	var body string
	switch raw {
	case "https://fixture.test/":
		header.Set("Content-Type", "text/html; charset=utf-8")
		body = `<!doctype html><html><head><title>Raw title</title><script src="/app.js"></script></head><body><main id="content">raw</main><img src="http://127.0.0.1/private"><a id="download" href="/download" download>download</a></body></html>`
	case "https://fixture.test/app.js":
		header.Set("Content-Type", "application/javascript")
		body = `document.addEventListener("DOMContentLoaded",()=>{const prior=localStorage.getItem("renderer-state");document.title=prior?"Persisted state":"Rendered title";localStorage.setItem("renderer-state","set");document.querySelector("#content").textContent="rendered";document.body.insertAdjacentHTML("beforeend",'<a href="/client-link">client</a>');navigator.geolocation.getCurrentPosition(()=>document.body.dataset.geo="granted",()=>document.body.dataset.geo="denied");navigator.serviceWorker.register("/sw.js").then(()=>document.body.dataset.sw="registered",()=>document.body.dataset.sw="blocked");document.querySelector("#download").click();});`
	case "https://fixture.test/download":
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Disposition", `attachment; filename="blocked.bin"`)
		body = "download must be cancelled"
	case "https://fixture.test/sw.js":
		header.Set("Content-Type", "application/javascript")
		body = `self.addEventListener("fetch",()=>{});`
	case "http://127.0.0.1/private":
		return fetchpolicy.FetchResult{}, errors.New("private target rejected")
	default:
		return fetchpolicy.FetchResult{}, errors.New("unexpected fixture URL")
	}
	return fetchpolicy.FetchResult{StatusCode: http.StatusOK, Header: header, Body: []byte(body)}, nil
}

func TestPlaywrightWorkerRendersOnlyMediatedResources(t *testing.T) {
	if os.Getenv("SEO_AUDITOR_RENDERER_INTEGRATION") != "1" {
		t.Skip("set SEO_AUDITOR_RENDERER_INTEGRATION=1 to run the browser integration")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "web", "renderer", "dist", "worker.js"))
	if err != nil {
		t.Fatal(err)
	}
	fetcher := &browserFixtureFetcher{}
	result, err := (&Supervisor{
		NodeBinary: "node", ScriptPath: script, ContainerSandbox: true,
		BrowserPath: os.Getenv("PLAYWRIGHT_BROWSERS_PATH"), Fetcher: fetcher,
	}).Render(context.Background(), Request{
		RequestID: "browser-fixture", URL: "https://fixture.test/",
		Deadline: 15 * time.Second, MaximumRequests: 20, MaximumBytes: 1 << 20,
		CaptureScreenshot: true, RunAccessibility: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || !strings.Contains(result.HTML, "Rendered title") || !strings.Contains(result.HTML, "/client-link") || !strings.Contains(result.HTML, `data-geo="denied"`) || !strings.Contains(result.HTML, `data-sw="blocked"`) {
		t.Fatalf("unexpected render result: %+v", result)
	}
	if len(fetcher.seen) < 3 || fetcher.seen[0] != "https://fixture.test/" || !containsString(fetcher.seen, "https://fixture.test/app.js") || !containsString(fetcher.seen, "http://127.0.0.1/private") {
		t.Fatalf("browser resources bypassed or diverged from mediation: %#v", fetcher.seen)
	}
	if len(result.Screenshot) == 0 || len(result.Accessibility) == 0 || len(result.ResourceFailures) == 0 || result.EngineVersion == "" {
		t.Fatalf("rendered diagnostics missing: screenshot=%d accessibility=%d resources=%d engine=%q", len(result.Screenshot), len(result.Accessibility), len(result.ResourceFailures), result.EngineVersion)
	}
	second, err := (&Supervisor{NodeBinary: "node", ScriptPath: script, ContainerSandbox: true, BrowserPath: os.Getenv("PLAYWRIGHT_BROWSERS_PATH"), Fetcher: &browserFixtureFetcher{}}).Render(context.Background(), Request{RequestID: "browser-fixture-second", URL: "https://fixture.test/", Deadline: 15 * time.Second, MaximumRequests: 20, MaximumBytes: 1 << 20})
	if err != nil || second.Status != "completed" || strings.Contains(second.HTML, "Persisted state") {
		t.Fatalf("browser state persisted across workers: result=%+v err=%v", second, err)
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

type hungFixtureFetcher struct{}

func (hungFixtureFetcher) Fetch(_ context.Context, _ string) (fetchpolicy.FetchResult, error) {
	body := `<html><head><script>while(true){}</script></head><body>never settles</body></html>`
	return fetchpolicy.FetchResult{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/html"}}, Body: []byte(body)}, nil
}

func TestPlaywrightHungPageIsBounded(t *testing.T) {
	if os.Getenv("SEO_AUDITOR_RENDERER_INTEGRATION") != "1" {
		t.Skip("set SEO_AUDITOR_RENDERER_INTEGRATION=1 to run the browser integration")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "web", "renderer", "dist", "worker.js"))
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	result, renderErr := (&Supervisor{NodeBinary: "node", ScriptPath: script, ContainerSandbox: true, BrowserPath: os.Getenv("PLAYWRIGHT_BROWSERS_PATH"), Fetcher: hungFixtureFetcher{}}).Render(context.Background(), Request{RequestID: "hung-fixture", URL: "https://hung.test/", Deadline: 500 * time.Millisecond, MaximumRequests: 5, MaximumBytes: 1 << 20})
	if renderErr == nil && result.Status == "completed" {
		t.Fatalf("hung page unexpectedly completed: %+v", result)
	}
	if elapsed := time.Since(started); elapsed > 7*time.Second {
		t.Fatalf("hung renderer exceeded supervisor bound: %s", elapsed)
	}
}

func TestSupervisorCrashIsBounded(t *testing.T) {
	t.Parallel()
	script := filepath.Join(t.TempDir(), "crash.mjs")
	if err := os.WriteFile(script, []byte(`process.exit(42);`), 0o600); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	_, err := (&Supervisor{NodeBinary: "node", ScriptPath: script, Fetcher: &fixtureFetcher{}}).Render(context.Background(), Request{RequestID: "crash", URL: "https://example.com/", Deadline: time.Second, MaximumRequests: 2, MaximumBytes: 1 << 20})
	if err == nil {
		t.Fatal("expected crashed worker to fail")
	}
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Fatalf("crashed worker was not reaped promptly: %s", elapsed)
	}
}
