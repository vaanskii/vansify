package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var driveService *drive.Service

// Initialize Google Drive client
func InitGoogleDrive() {
    ctx := context.Background()

    // Load the token from the file
    tokFile := "token.json"
    tok, err := tokenFromFile(tokFile)
    if err != nil {
        if os.IsNotExist(err) {
            log.Fatalf("Token file not found. Please authenticate using the authentication flow.")
        } else {
            log.Fatalf("Failed to load token from file: %v", err)
        }
        return
    }

    // Create OAuth2 config to generate the client
    config := &oauth2.Config{} // Dummy config as we only need the token
    client := config.Client(ctx, tok)

    // Initialize the Drive service
    driveService, err = drive.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
        log.Fatalf("Unable to retrieve Drive client: %v", err)
    }
    log.Println("Google Drive client initialized.")

    // Use driveService to list files as a placeholder action
    fileList, err := driveService.Files.List().Do()
    if err != nil {
        log.Fatalf("Unable to retrieve files: %v", err)
    }
    log.Printf("Found %d files in Google Drive.", len(fileList.Files))
}

// Reads the OAuth2 token from the specified file
func tokenFromFile(file string) (*oauth2.Token, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    tok := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(tok)
    return tok, err
}

// Saves the OAuth2 token to the specified file
func saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.Create(path)
    if err != nil {
        log.Fatalf("Unable to save oauth token: %v", err)
    }
    defer f.Close()

    json.NewEncoder(f).Encode(token)
}

// Handles file uploads to Google Drive
func UploadFile(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprintf("Unable to get file: %v", err))
        return
    }
    defer file.Close()

    driveFile, err := driveService.Files.Create(&drive.File{
        Name: header.Filename,
    }).Media(file).Do()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to upload file to Drive: %v", err))
        return
    }
    c.JSON(http.StatusOK, gin.H{"fileID": driveFile.Id, "fileName": driveFile.Name})
}
