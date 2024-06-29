package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/machinebox/graphql"
)

const (
	gqlEndpoint = "http://localhost:8081/query"
	repoURL     = "git@bitbucket.org:thevport/agreement.git"
)

type WebhookPayload struct {
	Push struct {
		Changes []struct {
			New struct {
				Target struct {
					Hash string `json:"hash"`
				} `json:"target"`
				Links struct {
					Commits struct {
						Href string `json:"href"`
					} `json:"commits"`
				} `json:"links"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
}

var (
	processedCommits = make(map[string]struct{})
	mu               sync.Mutex
)

func main() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Webhook received")

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		log.Printf("Invalid payload: %v\n", err)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for _, change := range payload.Push.Changes {
		commitHash := change.New.Target.Hash

		if _, found := processedCommits[commitHash]; found {
			log.Printf("Commit %s has already been processed, skipping...\n", commitHash)
			continue
		}

		processedCommits[commitHash] = struct{}{}
		w.WriteHeader(http.StatusOK)

		go processCommit(change.New.Links.Commits.Href, commitHash)
	}

	log.Println("Webhook processed successfully")
}

func processCommit(commitURL, commitHash string) {
	log.Printf("Processing commit: %s\n", commitURL)

	tempDir, err := downloadRepoAtCommit(commitHash)
	if err != nil {
		log.Printf("Failed to download repo: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	changedFiles, err := getChangedFiles(tempDir, commitHash)
	if err != nil {
		log.Printf("Failed to get changed files: %v\n", err)
		return
	}

	for _, filePath := range changedFiles {
		fullFilePath := filepath.Join(tempDir, filePath)
		if !isFile(fullFilePath) {
			continue
		}

		log.Printf("Uploading file: %s\n", fullFilePath)
		if err := uploadFileToGCS(fullFilePath); err != nil {
			log.Printf("Failed to upload to GCS: %v\n", err)
			return
		}
	}
}

func downloadRepoAtCommit(commitHash string) (string, error) {
	log.Printf("Downloading repo at commit: %s\n", commitHash)

	tempDir, err := os.MkdirTemp("", "repo")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}

	cmd := exec.Command("git", "clone", repoURL, tempDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Clone output: %s\n", string(output))
		return "", fmt.Errorf("failed to clone repository: %v", err)
	}

	cmd = exec.Command("git", "checkout", commitHash)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Checkout output: %s\n", string(output))
		return "", fmt.Errorf("failed to checkout commit: %v", err)
	}

	return tempDir, nil
}

func getChangedFiles(repoDir, commitHash string) ([]string, error) {
	log.Printf("Getting changed files for commit: %s\n", commitHash)

	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-only", "-r", commitHash)
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Get changed files output: %s\n", string(output))
		return nil, fmt.Errorf("failed to get changed files: %v", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	log.Printf("Changed files: %v\n", changedFiles)
	return changedFiles, nil
}

func isFile(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Failed to stat file: %v\n", err)
		return false
	}
	return !fileInfo.IsDir()
}
func uploadFileToGCS(filePath string) error {
	log.Printf("Uploading file to GCS: %s\n", filePath)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file content: %v", err)
	}

	// Encode file content to base64
	encodedContent := base64.StdEncoding.EncodeToString(fileContent)

	// Example version
	version := 2

	// Authenticate and get a token
	client := graphql.NewClient(gqlEndpoint)
	req := graphql.NewRequest(`
        mutation {
            login(input: {
                username: "saeed"
                password: "Saeed123@"
            }) {
                isAuthenticated
                message
            }
        }
    `)
	var authResponse map[string]interface{}
	if err := client.Run(context.Background(), req, &authResponse); err != nil {
		body, _ := json.Marshal(authResponse) // Marshal before logging
		log.Printf("Failed to authenticate: %v, Response: %s\n", err, body)
		return fmt.Errorf("failed to authenticate: %v", err)
	}
	authData := authResponse["login"].(map[string]interface{})
	if !authData["isAuthenticated"].(bool) {
		return fmt.Errorf("authentication failed: %s", authData["message"].(string))
	}
	log.Printf("Authenticated successfully: %s\n", authData["message"].(string))

	// Initiate EULA upload
	req = graphql.NewRequest(`
        mutation ($version: Int!, $content: String!) {
            initiateEulaUpload(version: $version, content: $content) {
                url
                filePath
            }
        }
    `)
	req.Var("version", version)
	req.Var("content", encodedContent) // Set content to the file content
	var initUploadResponse map[string]interface{}
	if err := client.Run(context.Background(), req, &initUploadResponse); err != nil {
		log.Printf("Failed to initiate EULA upload: %v\n", err)
		return fmt.Errorf("failed to initiate EULA upload: %v", err)
	}
	initUploadData := initUploadResponse["initiateEulaUpload"].(map[string]interface{})
	log.Printf("Initiated EULA upload with URL: %s\n", initUploadData["url"].(string))

	// Upload the file using the pre-signed URL
	uploadURL := initUploadData["url"].(string)
	httpReq, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	// Log response for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	bodyString := string(bodyBytes)
	log.Printf("Upload response status: %d, body: %s\n", resp.StatusCode, bodyString)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status code: %d", resp.StatusCode)
	}
	log.Println("File uploaded successfully")

	// Complete the EULA upload
	req = graphql.NewRequest(`
        mutation uploadComplete($version: Int!) {
            completeEulaUpload(version: $version) {
                version
                publicUrl
                status
            }
        }
    `)
	req.Var("version", version)
	var completeUploadResponse map[string]interface{}
	if err := client.Run(context.Background(), req, &completeUploadResponse); err != nil {
		log.Printf("Failed to complete EULA upload: %v\n", err)
		return fmt.Errorf("failed to complete EULA upload: %v", err)
	}
	completeUploadData := completeUploadResponse["completeEulaUpload"].(map[string]interface{})
	log.Printf("EULA available at: %s\n", completeUploadData["publicUrl"].(string))

	return nil
}
