package versions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/s3-resource/versions"
)

type MatchFunc func(paths []string, pattern string) ([]string, error)

var ItMatchesPaths = func(matchFunc MatchFunc) {
	Describe("checking if paths in the bucket should be searched", func() {
		Context("when given an empty list of paths", func() {
			It("returns an empty list of matches", func() {
				result, err := versions.Match([]string{}, "regex")

				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(BeEmpty())
			})
		})

		Context("when given a single path", func() {
			It("returns it in a singleton list if it matches the regex", func() {
				paths := []string{"a-folder/1/abc"}
				regex := "abc"

				result, err := versions.Match(paths, regex)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(ConsistOf("a-folder/1/abc"))
			})

			It("returns an empty list if it does not match the regexp", func() {
				paths := []string{"abc"}
				regex := "ad"

				result, err := versions.Match(paths, regex)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(BeEmpty())
			})

			It("accepts full regexes", func() {
				paths := []string{"a-folder/1/abc", "a-folder/1/adc"}
				regex := "a.*c"

				result, err := versions.Match(paths, regex)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(ConsistOf("a-folder/1/abc", "a-folder/1/adc"))
			})

			It("errors when the regex is bad", func() {
				paths := []string{"abc"}
				regex := "a(c"

				_, err := versions.Match(paths, regex)
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when given a multiple paths", func() {
			It("returns the matches", func() {
				paths := []string{"a-folder/1/abc", "a-folder/2/bcd"}
				regex := ".*bc.*"

				result, err := versions.Match(paths, regex)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(ConsistOf("a-folder/1/abc", "a-folder/2/bcd"))
			})

			It("returns an empty list if none match the regexp", func() {
				paths := []string{"abc", "def"}
				regex := "ge.*h"

				result, err := versions.Match(paths, regex)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(result).Should(BeEmpty())
			})
		})
	})
}

var _ = Describe("Match", func() {
	Describe("Match", func() {
		ItMatchesPaths(versions.Match)
	})

	Describe("MatchUnanchored", func() {
		ItMatchesPaths(versions.MatchUnanchored)
	})
})

var _ = Describe("PrefixHint", func() {
	It("turns a regexp into a limiter for s3", func() {
		Ω(versions.PrefixHint("hello/world")).Should(Equal("hello/world"))
		Ω(versions.PrefixHint("hello/*.tgz")).Should(Equal("hello"))
		Ω(versions.PrefixHint("")).Should(Equal(""))
		Ω(versions.PrefixHint("*")).Should(Equal(""))
		Ω(versions.PrefixHint("hello/*/what.txt")).Should(Equal("hello"))
	})
})

var _ = Describe("Extract", func() {
	Context("when the path does not contain extractable information", func() {
		It("doesn't extract it", func() {
			result, ok := versions.Extract("a-folder/version12/file", "a-folder/version-(.*)/file")
			Ω(ok).Should(BeFalse())
			Ω(result).Should(BeZero())
		})
	})

	Context("when the path contains extractable information", func() {
		It("extracts it", func() {
			result, ok := versions.Extract("a-folder/105/a-file", "a-folder/(.*)/a-file")
			Ω(ok).Should(BeTrue())

			Ω(result.Path).Should(Equal("a-folder/105/a-file"))
			Ω(result.Version.String()).Should(Equal("105.0.0"))
			Ω(result.VersionNumber).Should(Equal("105"))
		})

		It("extracts semantics version numbers", func() {
			result, ok := versions.Extract("a-folder/1.0.5/file", "a-folder/(.*)/file")
			Ω(ok).Should(BeTrue())

			Ω(result.Path).Should(Equal("a-folder/1.0.5/file"))
			Ω(result.Version.String()).Should(Equal("1.0.5"))
			Ω(result.VersionNumber).Should(Equal("1.0.5"))
		})

		It("takes the first match if there are many", func() {
			result, ok := versions.Extract("abc/1.0.5/file/2.3.4", "abc/(.*)/file/(.*)")
			Ω(ok).Should(BeTrue())

			Ω(result.Path).Should(Equal("abc/1.0.5/file/2.3.4"))
			Ω(result.Version.String()).Should(Equal("1.0.5"))
			Ω(result.VersionNumber).Should(Equal("1.0.5"))
		})
	})
})
