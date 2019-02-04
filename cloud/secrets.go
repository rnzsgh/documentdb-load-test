package cloud

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	log "github.com/golang/glog"
)

type AppSecrets struct {
	DatabasePassword string
}

var Secrets *AppSecrets

func init() {

	log.Info("Calling init on secrets")

	Secrets = &AppSecrets{}

	// Get from a local env var or pull from secrets manager
	if len(os.Getenv("DOCUMENT_DB_LOCAL")) > 0 {
		Secrets.DatabasePassword = os.Getenv("DOCUMENT_DB_PASSWORD")
		return
	}

	log.Infof("Secrets name %s", os.Getenv("DOCUMENT_DB_PASSWORD_SECRET_NAME"))
	if databasePassword, err := loadSecret(os.Getenv("DOCUMENT_DB_PASSWORD_SECRET_NAME")); err == nil {
		Secrets.DatabasePassword = databasePassword
	} else {
		log.Errorf("Cannot load secret: %s - problem: %v", os.Getenv("DOCUMENT_DB_PASSWORD_SECRET_NAME"), err)
	}
	log.Infof("Secrets password: %s", Secrets.DatabasePassword)
}

func loadSecret(secretName string) (string, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	svc := secretsmanager.New(session.New(), &aws.Config{HTTPClient: client})
	if result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}); err != nil {
		return "", err
	} else {
		log.Infof("Result returned: %v", result)
		return *result.SecretString, nil
	}
}
