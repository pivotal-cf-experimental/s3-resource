package in

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/cloudfoundry/gunk/urljoiner"
	"github.com/pivotal-cf-experimental/s3-resource"
	"github.com/pivotal-cf-experimental/s3-resource/versions"
)

type RequestURLProvider struct {
	s3Client s3resource.S3Client
}

func (up *RequestURLProvider) GetURL(request InRequest, remotePath string) string {
	if request.Source.CloudfrontURL != "" {
		return up.cloudfrontURL(request, remotePath)
	}

	return up.s3URL(request, remotePath)
}

func (up *RequestURLProvider) s3URL(request InRequest, remotePath string) string {
	return up.s3Client.URL(request.Source.Bucket, remotePath, request.Source.Private, request.Version.VersionID)
}

func (up *RequestURLProvider) cloudfrontURL(request InRequest, remotePath string) string {
	url := urljoiner.Join(request.Source.CloudfrontURL, remotePath)

	if request.Version.VersionID != "" {
		url = url + "?versionId=" + request.Version.VersionID
	}

	return url
}

type InCommand struct {
	s3client    s3resource.S3Client
	urlProvider RequestURLProvider
}

func NewInCommand(s3client s3resource.S3Client) *InCommand {
	return &InCommand{
		s3client: s3client,
		urlProvider: RequestURLProvider{
			s3Client: s3client,
		},
	}
}

func (command *InCommand) Run(destinationDir string, request InRequest) (InResponse, error) {
	if ok, message := request.Source.IsValid(); !ok {
		return InResponse{}, errors.New(message)
	}

	err := command.createDirectory(destinationDir)
	if err != nil {
		return InResponse{}, err
	}

	if request.Source.Regexp != "" {
		return command.inByRegex(destinationDir, request)
	} else {
		return command.inByVersionedFile(destinationDir, request)
	}
}

func (command *InCommand) inByRegex(destinationDir string, request InRequest) (InResponse, error) {
	remotePath, err := command.pathToDownload(request)
	if err != nil {
		return InResponse{}, err
	}

	extraction, ok := versions.Extract(remotePath, request.Source.Regexp)
	if ok {
		err := command.writeVersionFile(extraction.VersionNumber, destinationDir)
		if err != nil {
			return InResponse{}, err
		}
	}

	err = command.downloadFile(
		request.Source.Bucket,
		remotePath,
		destinationDir,
		path.Base(remotePath),
	)
	if err != nil {
		return InResponse{}, err
	}

	url := command.urlProvider.GetURL(request, remotePath)
	err = command.writeURLFile(
		destinationDir,
		url,
	)
	if err != nil {
		return InResponse{}, err
	}

	return InResponse{
		Version: s3resource.Version{
			Path: remotePath,
		},
		Metadata: command.metadata(remotePath, request.Source.Private, url),
	}, nil
}

func (command *InCommand) inByVersionedFile(destinationDir string, request InRequest) (InResponse, error) {

	err := command.writeVersionFile(request.Version.VersionID, destinationDir)
	if err != nil {
		return InResponse{}, err
	}

	remotePath := request.Source.VersionedFile
	versionedPath := remotePath + "?versionId=" + request.Version.VersionID
	err = command.downloadFile(
		request.Source.Bucket,
		versionedPath,
		destinationDir,
		path.Base(remotePath),
	)

	if err != nil {
		return InResponse{}, err
	}

	url := command.urlProvider.GetURL(request, remotePath)
	err = command.writeURLFile(
		destinationDir,
		url,
	)

	if err != nil {
		return InResponse{}, err
	}

	return InResponse{
		Version: s3resource.Version{
			VersionID: request.Version.VersionID,
		},
		Metadata: command.metadata(remotePath, request.Source.Private, url),
	}, nil

}

func (command *InCommand) pathToDownload(request InRequest) (string, error) {
	if request.Version.Path == "" {
		extractions := versions.GetBucketFileVersions(command.s3client, request.Source)

		if len(extractions) == 0 {
			return "", errors.New("no extractions could be found - is your regexp correct?")
		}

		lastExtraction := extractions[len(extractions)-1]
		return lastExtraction.Path, nil
	}

	return request.Version.Path, nil
}

func (command *InCommand) createDirectory(destDir string) error {
	return os.MkdirAll(destDir, 0755)
}

func (command *InCommand) writeURLFile(destDir string, url string) error {
	return ioutil.WriteFile(filepath.Join(destDir, "url"), []byte(url), 0644)
}

func (command *InCommand) writeVersionFile(versionNumber string, destDir string) error {
	return ioutil.WriteFile(filepath.Join(destDir, "version"), []byte(versionNumber), 0644)
}

func (command *InCommand) downloadFile(bucketName string, remotePath string, destinationDir string, destinationFile string) error {
	localPath := filepath.Join(destinationDir, destinationFile)

	return command.s3client.DownloadFile(
		bucketName,
		remotePath,
		localPath,
	)
}

func (command *InCommand) metadata(remotePath string, private bool, url string) []s3resource.MetadataPair {
	remoteFilename := filepath.Base(remotePath)

	metadata := []s3resource.MetadataPair{
		s3resource.MetadataPair{
			Name:  "filename",
			Value: remoteFilename,
		},
	}

	if !private {
		metadata = append(metadata, s3resource.MetadataPair{
			Name:  "url",
			Value: url,
		})
	}

	return metadata
}
