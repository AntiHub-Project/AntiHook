package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
	
	protocolRegistry "antihook/registry"
)

const (
	ProtocolDescription = "Kiro Protocol Handler"
	TargetDirName       = "Antihub"
	ExeName             = "antihook.exe"
)

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
	if len(args) > 0 && strings.HasPrefix(strings.ToLower(args[0]), "kiro://") {
		handleProtocolCall(args[0])
		return
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

	handler := &protocolRegistry.ProtocolHandler{
		Protocol:    protocolRegistry.ProtocolName,
		ExePath:     targetPath,
		Description: ProtocolDescription,
	}

	if err := handler.Register(); err != nil {
		return fmt.Errorf("failed to register protocol: %w", err)
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
		serverURL = "http://localhost:8045"
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