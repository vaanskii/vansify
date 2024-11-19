package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/nfnt/resize"
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

    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprintf("Unable to get file: %v", err))
        return
    }
    defer file.Close()

    // Decode the image
    img, _, err := image.Decode(file)
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprintf("Unable to decode image: %v", err))
        return
    }

    // Resize the image to a maximum width of 1024 pixels, preserving the aspect ratio
    resizedImage := resize.Resize(800, 0, img, resize.Lanczos3)

    // Create a temporary file to store the resized image
    tempFile, err := os.CreateTemp("", "resized-*.jpg")
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to create temporary file: %v", err))
        return
    }
    defer tempFile.Close()

    // Encode the resized image as a JPEG
    err = jpeg.Encode(tempFile, resizedImage, nil)
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to encode resized image: %v", err))
        return
    }

    // Reopen the temporary file for uploading to Google Drive
    tempFile, err = os.Open(tempFile.Name())
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to open temporary file: %v", err))
        return
    }
    defer tempFile.Close()

    driveFile, err := DriveService.Files.Create(&drive.File{
        Name:    header.Filename,
        Parents: []string{parentFolderID},
    }).Media(tempFile).Do()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to upload file to Drive: %v", err))
        return
    }
    log.Printf("File uploaded successfully: %s (ID: %s)", driveFile.Name, driveFile.Id)

    // Set permissions for the file to your email
    log.Printf("Setting permissions for file ID: %s", driveFile.Id)
    _, err = DriveService.Permissions.Create(driveFile.Id, &drive.Permission{
        Type:        "user",
        Role:        "writer",
        EmailAddress: "giorgivanadze03@gmail.com",
    }).Do()
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to set file permissions: %v", err))
        return
    }
    log.Printf("File permissions set for ID: %s", driveFile.Id)

    backendUrl := os.Getenv("BACKEND_URL")
    fileURL := fmt.Sprintf("%s/v1/file/%s", backendUrl, driveFile.Id)
    driveFileURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", driveFile.Id)
    log.Printf("Generated backend file URL: %s", fileURL)
    log.Printf("Generated Google Drive file URL: %s", driveFileURL)

    c.JSON(http.StatusOK, gin.H{
        "fileID":   driveFile.Id,
        "fileName": driveFile.Name,
        "fileURL":  fileURL,
        "driveFileURL":  driveFileURL,
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

var cache = make(map[string][]byte)
var cacheLock sync.RWMutex

func ServeFile(c *gin.Context) {
    fileID := c.Param("fileID")

    // Check cache first
    cacheLock.RLock()
    if data, found := cache[fileID]; found {
        cacheLock.RUnlock()
        c.Header("Content-Type", "image/jpeg") 
        c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileID))
        c.Writer.Write(data)
        return
    }
    cacheLock.RUnlock()

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

    buf := new(bytes.Buffer)
    buf.ReadFrom(response.Body)
    data := buf.Bytes()

    // Update cache
    cacheLock.Lock()
    cache[fileID] = data
    cacheLock.Unlock()

    c.Header("Content-Type", file.MimeType)
    c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.Name))
    c.Writer.Write(data)
}


// Function to delete the chat folder and its contents from Google Drive
func DeleteChatAndImages(chatID, userEmail string) error {
    folderName := "chat/" + chatID
    log.Printf("Retrieving folder ID for: %s", folderName)

    // List all folders with the specified name
    fileList, err := DriveService.Files.List().Q(fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder'", folderName)).Fields("files(id)").Do()
    if err != nil {
        return fmt.Errorf("unable to list folders: %v", err)
    }

    if len(fileList.Files) == 0 {
        return fmt.Errorf("folder not found for chatID: %s", chatID)
    }
    folderID := fileList.Files[0].Id

    setPermissions := func(fileID string) error {
        permissions := []*drive.Permission{
            {
                Type:        "user",
                Role:        "writer",
                EmailAddress: userEmail,
            },
        }
        for _, permission := range permissions {
            if _, err := DriveService.Permissions.Create(fileID, permission).Do(); err != nil {
                return fmt.Errorf("unable to set permission for file ID: %s, error: %v", fileID, err)
            }
        }
        return nil
    }

    // Set permissions for all files in the folder before deleting
    for _, file := range fileList.Files {
        if err := setPermissions(file.Id); err != nil {
            log.Printf("Unable to set permissions for file ID: %s, error: %v", file.Id, err)
        }
    }

    // List all files in the folder
    filesList, err := DriveService.Files.List().Q(fmt.Sprintf("'%s' in parents", folderID)).Fields("files(id)").Do()
    if err != nil {
        return fmt.Errorf("unable to list files in folder: %v", err)
    }

    for _, file := range filesList.Files {
        if err := DriveService.Files.Delete(file.Id).Do(); err != nil {
            log.Printf("Unable to delete file ID: %s, error: %v", file.Id, err)
        }
    }

    if err := setPermissions(folderID); err != nil {
        log.Printf("Unable to set permissions for folder ID: %s, error: %v", folderID, err)
    }

    // Delete the folder itself
    if err := DriveService.Files.Delete(folderID).Do(); err != nil {
        return fmt.Errorf("unable to delete folder: %v", err)
    }

    return nil
}

func FolderExists(folderName string) (bool, error) {
    q := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder'", folderName)
    fileList, err := DriveService.Files.List().Q(q).Fields("files(id)").Do()
    if err != nil {
        return false, fmt.Errorf("unable to query existing folders: %v", err)
    }
    if len(fileList.Files) > 0 {
        log.Printf("Folder exists: %s (ID: %s)", folderName, fileList.Files[0].Id)
        return true, nil
    }
    log.Printf("Folder does not exist: %s", folderName)
    return false, nil
}
