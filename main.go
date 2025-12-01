package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"

	protocolRegistry "antihook/registry"
)

const (
	ProtocolDescription     = "Kiro Protocol Handler"
	AntiProtocolDescription = "Anti Protocol Handler"
	TargetDirName           = "Antihub"
	ExeName                 = "antihook.exe"
	OAuthCallbackPort       = 42532
)

var DefaultServerURL = "http://localhost:8045"
var DefaultBackendURL = "http://localhost:8008"

func init() {
	if url := os.Getenv("KIRO_SERVER_URL"); url != "" {
		DefaultServerURL = url
	}
	if url := os.Getenv("BACKEND_URL"); url != "" {
		DefaultBackendURL = url
	}
}

func main() {
	recoverFlag := flag.Bool("recover", false, "Restore original Kiro protocol handler")
	flag.Parse()

	if *recoverFlag {
		if err := recoverOriginal(); err != nil {
			showMessageBox("Error", "Recovery failed: "+err.Error(), 0x10)
			os.Exit(1)
		}
		showMessageBox("Success", "Protocol handler restored!", 0x40)
		return
	}

	args := flag.Args()
	if len(args) > 0 {
		lowerArg := strings.ToLower(args[0])
		if strings.HasPrefix(lowerArg, "kiro://") {
			handleProtocolCall(args[0])
			return
		}
		if strings.HasPrefix(lowerArg, "anti://") {
			handleAntiProtocolCall(args[0])
			return
		}
	}

	if err := install(); err != nil {
		showMessageBox("Error", "Installation failed: "+err.Error(), 0x10)
		os.Exit(1)
	}

	showMessageBox("Success", "Hooked successfully!", 0x40)
}

func install() error {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("cannot get LOCALAPPDATA environment variable")
	}

	targetDir := filepath.Join(localAppData, TargetDirName)
	targetPath := filepath.Join(targetDir, ExeName)

	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}
	currentPath, _ = filepath.Abs(currentPath)

	if !strings.EqualFold(currentPath, targetPath) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}

		if _, err := os.Stat(targetPath); err == nil {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("failed to remove old file: %w", err)
			}
		}

		if err := copyFile(currentPath, targetPath); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	kiroHandler := &protocolRegistry.ProtocolHandler{
		Protocol:    protocolRegistry.ProtocolName,
		ExePath:     targetPath,
		Description: ProtocolDescription,
	}

	if err := kiroHandler.Register(); err != nil {
		return fmt.Errorf("failed to register kiro protocol: %w", err)
	}

	antiHandler := &protocolRegistry.ProtocolHandler{
		Protocol:    protocolRegistry.AntiProtocolName,
		ExePath:     targetPath,
		Description: AntiProtocolDescription,
	}

	if err := antiHandler.Register(); err != nil {
		return fmt.Errorf("failed to register anti protocol: %w", err)
	}

	if err := addToPath(targetDir); err != nil {
		fmt.Printf("Warning: failed to add to PATH: %v\n", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

func handleProtocolCall(rawURL string) {
	go showMessageBox("Info", "Logging in...", 0x40)

	if err := postCallback(rawURL); err != nil {
		showMessageBox("Error", "Login failed: "+err.Error(), 0x10)
		return
	}

	showMessageBox("Success", "Login successful!", 0x40)
}

func postCallback(callbackURL string) error {
	requestBody := map[string]string{
		"callback_url": callbackURL,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to serialize request body: %w", err)
	}

	serverURL := os.Getenv("KIRO_SERVER_URL")
	if serverURL == "" {
		serverURL = DefaultServerURL
	}

	apiURL := serverURL + "/api/kiro/oauth/callback"

	resp, err := http.Post(
		apiURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %d, %s", resp.StatusCode, string(body))
	}

	return nil
}

type AntiProtocolParams struct {
	Bearer   string
	IsShared int
}

type OAuthAuthorizeResponse struct {
	Success bool `json:"success"`
	Data    struct {
		AuthURL   string `json:"auth_url"`
		State     string `json:"state"`
		ExpiresIn int    `json:"expires_in"`
	} `json:"data"`
}

func parseAntiProtocolURL(rawURL string) (*AntiProtocolParams, error) {
	withoutProtocol := strings.TrimPrefix(rawURL, "anti://")
	withoutProtocol = strings.TrimPrefix(withoutProtocol, "Anti://")
	withoutProtocol = strings.TrimPrefix(withoutProtocol, "ANTI://")

	parts := strings.SplitN(withoutProtocol, "?", 2)
	
	var bearer string
	isShared := 0

	if len(parts) > 1 {
		queryParams, err := url.ParseQuery(parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse query parameters: %w", err)
		}

		identity := queryParams.Get("identity")
		if identity == "" {
			return nil, fmt.Errorf("missing identity parameter")
		}

		if strings.HasPrefix(identity, "Bearer ") {
			bearer = identity
		} else if strings.HasPrefix(identity, "bearer ") {
			bearer = "Bearer " + strings.TrimPrefix(identity, "bearer ")
		} else {
			bearer = "Bearer " + identity
		}

		if val := queryParams.Get("is_shared"); val != "" {
			if val == "1" || strings.ToLower(val) == "true" {
				isShared = 1
			}
		}
	} else {
		return nil, fmt.Errorf("missing query parameters")
	}

	return &AntiProtocolParams{
		Bearer:   bearer,
		IsShared: isShared,
	}, nil
}

func handleAntiProtocolCall(rawURL string) {
	params, err := parseAntiProtocolURL(rawURL)
	if err != nil {
		showMessageBox("Error", "Failed to parse URL: "+err.Error(), 0x10)
		return
	}

	serverURL := os.Getenv("BACKEND_URL")
	if serverURL == "" {
		serverURL = DefaultBackendURL
	}

	authResp, err := requestOAuthAuthorize(serverURL, params.Bearer, params.IsShared)
	if err != nil {
		showMessageBox("Error", "OAuth authorization failed: "+err.Error(), 0x10)
		return
	}

	callbackChan := make(chan string, 1)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(authResp.Data.ExpiresIn)*time.Second)
	defer cancel()

	server := startOAuthCallbackServer(ctx, callbackChan, errChan, &wg)

	if err := openBrowser(authResp.Data.AuthURL); err != nil {
		showMessageBox("Error", "Failed to open browser: "+err.Error(), 0x10)
		server.Shutdown(context.Background())
		return
	}

	select {
	case callbackURL := <-callbackChan:
		if err := postOAuthCallbackManual(serverURL, params.Bearer, callbackURL); err != nil {
			showMessageBox("Error", "Failed to complete OAuth: "+err.Error(), 0x10)
		} else {
			showMessageBox("Success", "Login successful!", 0x40)
		}
	case err := <-errChan:
		showMessageBox("Error", "Callback server error: "+err.Error(), 0x10)
	case <-ctx.Done():
		showMessageBox("Error", "OAuth timeout - please try again", 0x10)
	}

	server.Shutdown(context.Background())
	wg.Wait()
}

func requestOAuthAuthorize(serverURL, bearer string, isShared int) (*OAuthAuthorizeResponse, error) {
	apiURL := serverURL + "/api/plugin-api/oauth/authorize"

	requestBody := map[string]int{
		"is_shared": isShared,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error: %d, %s", resp.StatusCode, string(body))
	}

	var authResp OAuthAuthorizeResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !authResp.Success {
		return nil, fmt.Errorf("authorization failed")
	}

	return &authResp, nil
}

func startOAuthCallbackServer(ctx context.Context, callbackChan chan<- string, errChan chan<- error, wg *sync.WaitGroup) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/oauth-callback", func(w http.ResponseWriter, r *http.Request) {
		callbackURL := fmt.Sprintf("http://localhost:%d%s", OAuthCallbackPort, r.URL.String())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Login Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .container {
            text-align: center;
            padding: 40px 60px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
        }
        p {
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Login Successful!</h1>
        <p>You can close this window now.</p>
    </div>
</body>
</html>`))

		select {
		case callbackChan <- callbackURL:
		default:
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", OAuthCallbackPort),
		Handler: mux,
	}

	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)

	return server
}

func openBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func postOAuthCallbackManual(serverURL, bearer, callbackURL string) error {
	apiURL := serverURL + "/api/plugin-api/oauth/callback"

	requestBody := map[string]string{
		"callback_url": callbackURL,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to serialize request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %d, %s", resp.StatusCode, string(body))
	}

	return nil
}

func showMessageBox(title, message string, flags uint) {
	var mod = syscall.NewLazyDLL("user32.dll")
	var proc = mod.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	proc.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)
}

func addToPath(dir string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open Environment key: %w", err)
	}
	defer key.Close()

	currentPath, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to read PATH: %w", err)
	}

	if strings.Contains(strings.ToLower(currentPath), strings.ToLower(dir)) {
		return nil
	}

	var newPath string
	if currentPath == "" {
		newPath = dir
	} else {
		newPath = currentPath + ";" + dir
	}

	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("failed to set PATH: %w", err)
	}

	return nil
}

func recoverOriginal() error {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("cannot get LOCALAPPDATA environment variable")
	}

	originalPath := filepath.Join(localAppData, "Programs", "Kiro", "Kiro.exe")
	originalCommand := fmt.Sprintf(`"%s" "--open-url" "--" "%%1"`, originalPath)

	keyPath := `Software\Classes\kiro\shell\open\command`
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open command key: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue("", originalCommand); err != nil {
		return fmt.Errorf("failed to set command: %w", err)
	}

	return nil
}