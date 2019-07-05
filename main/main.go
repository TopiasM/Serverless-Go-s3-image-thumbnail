package main

import (
	"context"
	"os"
	"log"
	"image"
	"image/jpeg"
	"bytes"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/disintegration/imaging"
)

func Handler(ctx context.Context, s3Event events.S3Event)  {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if(err != nil) {
		log.Printf("%s", err)
	}

	svc := s3.New(sess)

	imgKey := s3Event.Records[0].S3.Object.Key

	log.Printf("%s to thumbnail", imgKey)

	result, err := svc.GetObject(&s3.GetObjectInput {
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key: aws.String(imgKey),
	})
	if(err != nil) {
		log.Printf("%s", err)
	}

	body := result.Body
	ImageData, _, err := image.Decode(body)

	options := jpeg.Options{Quality: 90}
	img := imaging.Resize(ImageData, os.Getenv("THUMB_HEIGHT"), os.Getenv("THUMB_WIDTH"), imaging.CatmullRom)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &options)
	if(err != nil) {
		log.Printf("%s", err)
	}

	nameSlice := strings.Split(imgKey, "/")
	thumbKey := strings.Join([]string{nameSlice[0], "-thumb/", nameSlice[1]}, "")
	b := buf.Bytes()

	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key: aws.String(thumbKey),
		Body: bytes.NewReader(b),
	})
	if(err != nil) {
		log.Printf("d %s", err)
	}
}

func main() {
	lambda.Start(Handler)
}
