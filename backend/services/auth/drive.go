package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var DriveService *drive.Service

func InitGoogleDrive() {
    log.Println("Initializing Google Drive service...")
    ctx := context.Background()

    tokFile := "token.json"
    tok, err := tokenFromFile(tokFile)
    if err != nil {
        log.Fatalf("Failed to load token from file: %v", err)
    }

    if tok.Expiry.Before(time.Now()) {
        tok, err = refreshAccessToken(tok)
        if err != nil {
            log.Fatalf("Unable to refresh access token: %v", err)
        }
    }

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

    var errDrive error
    DriveService, errDrive = drive.NewService(ctx, option.WithHTTPClient(client))
    if errDrive != nil {
        log.Fatalf("Unable to initialize Google Drive client: %v", errDrive)
    }

    fileList, err := DriveService.Files.List().Do()
    if err != nil {
        log.Fatalf("Unable to retrieve files: %v", err)
    }

    log.Printf("Found %d files in Google Drive.", len(fileList.Files))
    log.Println("Google Drive client initialized successfully.")
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
    envErr := godotenv.Load() 
    if envErr != nil { 
        log.Fatalf("Error loading .env file: %v", envErr) 
    }

    var folderName string
    var parentFolderID string

    if chatID := c.Param("chatid"); chatID != "" {
        folderName = "chat/" + chatID
        log.Printf("Chat ID provided: %s", chatID)
    } else if username := c.Param("username"); username != "" {
        folderName = "profile/" + username
        log.Printf("Username provided: %s", username)
    } else {
        c.String(http.StatusBadRequest, "Invalid request: missing folder identifier")
        return
    }

    log.Printf("Ensuring folder exists: %s", folderName)
    parentFolderID, err := EnsureFolderExists(folderName)
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to ensure folder exists: %v", err))
        return
    }

    // Get the file from the request
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprintf("Unable to get file: %v", err))
        return
    }
    defer file.Close()

    driveFile, err := DriveService.Files.Create(&drive.File{
        Name:    header.Filename,
        Parents: []string{parentFolderID},
    }).Media(file).Do()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to upload file to Drive: %v", err))
        return
    }
    log.Printf("File uploaded successfully: %s (ID: %s)", driveFile.Name, driveFile.Id)

    log.Printf("Setting permissions for file ID: %s", driveFile.Id)
    _, err = DriveService.Permissions.Create(driveFile.Id, &drive.Permission{
        Type: "anyone",
        Role: "reader",
    }).Do()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to set file permissions: %v", err))
        return
    }
    log.Printf("File permissions set to public for ID: %s", driveFile.Id)

    backendUrl := os.Getenv("BACKEND_URL")
    fileURL := fmt.Sprintf("%s/v1/file/%s", backendUrl, driveFile.Id)
    log.Printf("Generated backend file URL: %s", fileURL)

    c.JSON(http.StatusOK, gin.H{
        "fileID":   driveFile.Id,
        "fileName": driveFile.Name,
        "fileURL":  fileURL,
    })
}

func EnsureFolderExists(folderName string) (string, error) {
    q := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder'", folderName)
    fileList, err := DriveService.Files.List().Q(q).Do()
    if err != nil {
        return "", fmt.Errorf("unable to query existing folders: %v", err)
    }
    if len(fileList.Files) > 0 {
        log.Printf("Folder already exists: %s (ID: %s)", folderName, fileList.Files[0].Id)
        return fileList.Files[0].Id, nil
    }

    log.Printf("Creating new folder: %s", folderName)
    folder := &drive.File{
        Name:     folderName,
        MimeType: "application/vnd.google-apps.folder",
    }
    folder, err = DriveService.Files.Create(folder).Do()
    if err != nil {
        return "", fmt.Errorf("unable to create folder: %v", err)
    }
    log.Printf("New folder created: %s (ID: %s)", folderName, folder.Id)

    return folder.Id, nil
}

func ServeFile(c *gin.Context) {
    fileID := c.Param("fileID")
    file, err := DriveService.Files.Get(fileID).Fields("mimeType, name").Do()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to retrieve file metadata: %v", err))
        return
    }
    
    response, err := DriveService.Files.Get(fileID).Download()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to download file: %v", err))
        return
    }
    defer response.Body.Close()

    c.Header("Content-Type", file.MimeType)
    c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.Name))
    
    // Stream the file content to the client
    _, err = io.Copy(c.Writer, response.Body)
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to write file to response: %v", err))
        return
    }
}