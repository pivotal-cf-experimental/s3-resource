package check_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/s3-resource"

	"github.com/pivotal-cf-experimental/s3-resource/fakes"

	. "github.com/pivotal-cf-experimental/s3-resource/check"
)

var _ = Describe("Out Command", func() {
	Describe("running the command", func() {
		var (
			tmpPath string
			request CheckRequest

			s3client *fakes.FakeS3Client
			command  *CheckCommand
		)

		BeforeEach(func() {
			var err error
			tmpPath, err = ioutil.TempDir("", "check_command")
			Ω(err).ShouldNot(HaveOccurred())

			request = CheckRequest{
				Source: s3resource.Source{
					Bucket: "bucket-name",
				},
			}

			s3client = &fakes.FakeS3Client{}
			command = NewCheckCommand(s3client)

			s3client.BucketFilesReturns([]string{
				"folder/1/abc",
				"folder/2/abc",
				"folder/3/abc",
				"folder/4/abc",
			}, nil)
		})

		AfterEach(func() {
			err := os.RemoveAll(tmpPath)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when there is a previous version", func() {
			It("includes all versions between the previous one and the current one", func() {
				request.Version.Path = ""
				request.Source.Folder = "folder"
				request.Source.Filename = "abc"

				response, err := command.Run(request)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(1))
				Ω(response).Should(ConsistOf(
					s3resource.Version{
						Path: "folder/4/abc",
					},
				))
			})

			Context("when the regexp does not match anything", func() {
				It("does not explode", func() {
					request.Source.Folder = "wrong-folder"
					request.Source.Filename = "abc"
					response, err := command.Run(request)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(0))
				})
			})
		})

		Context("when there is no previous version", func() {
			It("includes the latest version only", func() {
				request.Version.Path = "folder/2/abc"
				request.Source.Folder = "folder"
				request.Source.Filename = "abc"

				response, err := command.Run(request)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(2))
				Ω(response).Should(ConsistOf(
					s3resource.Version{
						Path: "folder/3/abc",
					},
					s3resource.Version{
						Path: "folder/4/abc",
					},
				))
			})
		})
	})
})
