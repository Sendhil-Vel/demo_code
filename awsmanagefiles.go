package myawsfiles

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
)

/*
	init : this is the function for init
*/
var logStatus bool
var awsAccessKey string
var awsSecretKey string
var awsRegion string
var OrgFolderpath string
var uploadBucket string
var awsBaseEndpoint string

var sess *session.Session
var err error
var _svc *ses.SES

type sizer interface {
	Size() int64
}

/*
loginfo - This will be used to login information.
	If set true while package intialization, then everything will be logged
*/
func loginfo(info string) {
	if logStatus == true {
		fmt.Println(info)
	}
}
func NewuploadFiletoS3(w http.ResponseWriter, r *http.Request) {
	folderpath := OrgFolderpath
	if folderpath == "" {
		http.Error(w, "missing path query parameter", http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(32 << 30)

	dataFile, dataheader, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	if dataFile == nil {
		fmt.Fprintln(w, errors.New("No Data Found"))
		return
	}

	defer dataFile.Close()
	processFile(dataFile)

	fileHeader := make([]byte, dataFile.(sizer).Size())

	// Copy the headers into the FileHeader buffer
	_, err = dataFile.Read(fileHeader)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	// set position back to start.
	_, err = dataFile.Seek(0, 0)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	filename := dataheader.Filename
	if strings.Contains(filename, ".PDF") {
		filename = strings.Replace(filename, ".PDF", ".pdf", -1)
	}
	if strings.Contains(filename, "'") {
		filename = strings.Replace(filename, "'", "", -1)
	}

	folderpath = filepath.Clean(folderpath) + "/" + filename
	if strings.Contains(folderpath, "'") {
		folderpath = strings.Replace(folderpath, "'", "", -1)
	}
	buffer := make([]byte, 512)
	_, err = dataFile.Read(buffer)
	if err != nil {
		return
	}
	contenttype := http.DetectContentType(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	if contenttype == "application/pdf" {
		_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
			Bucket:        aws.String(uploadBucket),
			Key:           aws.String(folderpath),
			ACL:           aws.String("public-read"),
			Body:          bytes.NewReader(fileHeader),
			ContentLength: aws.Int64(dataFile.(sizer).Size()),
			ContentType:   aws.String("application/pdf"),
			Metadata: map[string]*string{
				"Key": aws.String(http.DetectContentType(fileHeader)),
			},
		})
	} else {
		_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
			Bucket:        aws.String(uploadBucket),
			Key:           aws.String(folderpath),
			ACL:           aws.String("public-read"),
			Body:          bytes.NewReader(fileHeader),
			ContentLength: aws.Int64(dataFile.(sizer).Size()),
			ContentType:   aws.String(contenttype),
			Metadata: map[string]*string{
				"Key": aws.String(http.DetectContentType(fileHeader)),
			},
		})
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.Contains(folderpath, " ") {
		folderpath = strings.Replace(folderpath, " ", "+", -1)
	}

	msg := fmt.Sprintln("https://" + awsBaseEndpoint + "/" + uploadBucket + "/" + folderpath)

	fmt.Fprintln(w, msg)

}

/*
processFile for processing the file
*/
func processFile(f io.Reader) error {
	return nil
}

func initAws() {
	var token = ""

	if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
		loginfo("Missing Required credentials")
		return
	}

	sess, err = session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, token),
	})
	if err != nil {
		loginfo("bad credentials")
	} else {
		loginfo("AWS connection established")
	}

}

func init() {
	fmt.Println("From Package")
}

/*
SetupPackage - This will take all the parameters and setup channel
	logstatus : If you set this to true then information will be displayed as and when process is going on
	awsAccessKey : This is the AWS access Key
	awsSecretKey : This is the AWS secret Key
	awsRegion : This is the AWS region value
*/
func SetupPackage(vLogStatus bool, vAwsAccessKey string, vAwsSecretKey string, vAwsRegion string, vFolderPath string, vUploadBucket string, vAwsBaseEndpoint string) {
	fmt.Println(vLogStatus, " : ", vAwsAccessKey, " : ", vAwsSecretKey, " : ", vAwsRegion, " : ", vFolderPath, " : ", vUploadBucket, " : ", vAwsBaseEndpoint)
	logStatus = vLogStatus
	if vAwsAccessKey == "" {
		loginfo("Missing Required credentials [vAwsAccessKey]")
		return
	} else {
		loginfo("Parameter exists [awsAccessKey]")
		awsAccessKey = vAwsAccessKey
	}
	if vAwsSecretKey == "" {
		loginfo("Missing Required credentials [vAwsSecretKey]")
		return
	} else {
		loginfo("Parameter exists [awsSecretKey]")
		awsSecretKey = vAwsSecretKey
	}
	if vAwsRegion == "" {
		loginfo("Missing Required credentials [vAwsRegion]")
		return
	} else {
		loginfo("Parameter exists [vAwsRegion]")
		awsRegion = vAwsRegion
	}
	if vFolderPath == "" {
		loginfo("Missing Required credentials [vFolderPath]")
		return
	} else {
		loginfo("Parameter exists [vFolderPath]")
		OrgFolderpath = vFolderPath
	}
	if vUploadBucket == "" {
		loginfo("Missing Required credentials [vUploadBucket]")
		return
	} else {
		loginfo("Parameter exists [vUploadBucket]")
		uploadBucket = vUploadBucket
	}
	if vAwsBaseEndpoint == "" {
		loginfo("Missing Required credentials [vAwsBaseEndpoint]")
		return
	} else {
		loginfo("Parameter exists [vAwsBaseEndpoint]")
		awsBaseEndpoint = vAwsBaseEndpoint
	}
	initAws()
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))

	// create a ses session
	_svc = ses.New(sess)
}
