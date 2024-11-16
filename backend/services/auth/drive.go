package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var driveService *drive.Service

func InitGoogleDrive() {
    ctx := context.Background()

    tokFile := "token.json"
    tok, err := tokenFromFile(tokFile)
    if err != nil {
        if os.IsNotExist(err) {
            log.Println("Token file not found. Please authenticate using the authentication flow.")
            return
        } else {
            log.Fatalf("Failed to load token from file: %v", err)
        }
    }

    if tok.Expiry.Before(time.Now()) {
        tok, err = refreshAccessToken(tok)
        if err != nil {
            log.Printf("Unable to refresh access token: %v", err)
            log.Println("Please authenticate using the authentication flow.")
            return
        }
    }

    // Create OAuth2 config to generate the client
    config := &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        Scopes: []string{
            "https://www.googleapis.com/auth/drive.file",
            "https://www.googleapis.com/auth/drive.readonly",
            "https://www.googleapis.com/auth/drive",
        },
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://accounts.google.com/o/oauth2/auth",
            TokenURL: "https://accounts.google.com/o/oauth2/token",
        },
    }

    client := config.Client(ctx, tok)

    // Initialize the Drive service
    driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
        log.Fatalf("Unable to retrieve Drive client: %v", err)
    }
    log.Println("Google Drive client initialized.")

    fileList, err := driveService.Files.List().Do()
    if err != nil {
        log.Fatalf("Unable to retrieve files: %v", err)
    }
    log.Printf("Found %d files in Google Drive.", len(fileList.Files))
}


func refreshAccessToken(tok *oauth2.Token) (*oauth2.Token, error) {
    config := &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        Scopes:       []string{drive.DriveScope},
        Endpoint:     google.Endpoint,
    }

    // Use the refresh token to get a new access token
    newToken, err := config.TokenSource(context.Background(), tok).Token()
    if err != nil {
        return nil, fmt.Errorf("unable to refresh token: %v", err)
    }

    // Save the new token to the file
    saveToken("token.json", newToken)

    return newToken, nil
}

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

func saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.Create(path)
    if err != nil {
        log.Fatalf("Unable to save oauth token: %v", err)
    }
    defer f.Close()

    json.NewEncoder(f).Encode(token)
}

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
