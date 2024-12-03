package aws

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

var sess *session.Session

func InitAWSSession() {
    if err := godotenv.Load(); err != nil {
        log.Printf("Error loading .env file: %v", err)
    }

    var err error
    sess, err = session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
        Credentials: credentials.NewStaticCredentials(
            os.Getenv("AWS_ACCESS_KEY_ID"),
            os.Getenv("AWS_SECRET_ACCESS_KEY"),
            "",
        ),
    })
    if err != nil {
        log.Printf("Unable to initialize AWS session: %v", err)
    } else {
        log.Println("AWS session initialized successfully.")
    }
}

func UploadFile(c *gin.Context) {
    c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
    c.Header("Access-Control-Expose-Headers", "Content-Length, ETag")

    if err := godotenv.Load(); err != nil {
        log.Printf("Error loading .env file: %v", err)
    }

    var folderName string
    if chatID := c.Param("chatid"); chatID != "" {
        folderName = "chat/" + chatID
    } else if username := c.Param("username"); username != "" {
        folderName = "profile/" + username
    } else {
        c.String(http.StatusBadRequest, "Invalid request: missing folder identifier")
        return
    }

    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprintf("Unable to get file: %v", err))
        return
    }
    defer file.Close()

    img, format, err := image.Decode(file)
    if err != nil {
        c.String(http.StatusBadRequest, fmt.Sprintf("Unable to decode image: %v", err))
        return
    }
    resizedImage := resize.Resize(800, 0, img, resize.Lanczos3)

    buf := new(bytes.Buffer)
    switch format {
    case "jpeg", "jpg":
        err = jpeg.Encode(buf, resizedImage, &jpeg.Options{Quality: 75})
    case "png":
        err = png.Encode(buf, resizedImage)
    case "gif":
        err = gif.Encode(buf, resizedImage, nil)
    case "bmp":
        err = bmp.Encode(buf, resizedImage)
    case "tiff":
        err = tiff.Encode(buf, resizedImage, nil)
    default:
        c.String(http.StatusBadRequest, "Unsupported image format")
        return
    }

    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to encode resized image: %v", err))
        return
    }

    svc := s3.New(sess)
    bucketName := os.Getenv("AWS_BUCKET_NAME")

    uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(header.Filename))
    key := fmt.Sprintf("%s/%s", folderName, uniqueFilename)

    contentType := map[string]string{
        "jpeg": "image/jpeg",
        "jpg":  "image/jpeg",
        "png":  "image/png",
        "gif":  "image/gif",
        "bmp":  "image/bmp",
        "tiff": "image/tiff",
    }[format]

    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket:             aws.String(bucketName),
        Key:                aws.String(key),
        Body:               bytes.NewReader(buf.Bytes()),
        ContentType:        aws.String(contentType),
        ContentDisposition: aws.String("inline"),
        ACL:                aws.String("public-read"),
    })
    if err != nil {
        c.String(http.StatusInternalServerError, fmt.Sprintf("Unable to upload file to S3: %v", err))
        return
    }

    // Generate file URL
    fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, os.Getenv("AWS_REGION"), key)

    c.JSON(http.StatusOK, gin.H{
        "fileName": uniqueFilename,
        "fileURL":  fileURL,
    })
}
