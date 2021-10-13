package list_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gotgenes/getignore/identifiers"
	"github.com/gotgenes/getignore/list"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("GitHubLister", func() {
	var (
		ctx    context.Context
		server *ghttp.Server
		lister list.GitHubLister
	)

	BeforeEach(func() {
		ctx = context.Background()
		server = ghttp.NewServer()
		lister, _ = list.NewGitHubLister(list.WithBaseURL(server.URL()))
	})

	AfterEach(func() {
		server.Close()
	})

	Context("basic functionality", func() {
		expectedUserAgent := []string{fmt.Sprintf("getignore/%s", identifiers.Version)}
		BeforeEach(func() {
			responseBody := `{
	  "name": "master",
	  "commit": {
		"sha": "b0012e4930d0a8c350254a3caeedf7441ea286a3",
		"commit": {
		  "tree": {
			"sha": "5adf061bdde4dd26889be1e74028b2f54aabc346"
		  }
		}
	  }
	}`
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v3/repos/github/gitignore/branches/master"),
					ghttp.VerifyHeader(http.Header{
						"User-Agent": expectedUserAgent,
					}),
					ghttp.VerifyHeader(http.Header{
						"Accept": []string{"application/vnd.github.v3+json"},
					}),
					ghttp.RespondWith(http.StatusOK, responseBody),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v3/repos/github/gitignore/git/trees/5adf061bdde4dd26889be1e74028b2f54aabc346"),
					ghttp.VerifyHeader(http.Header{
						"User-Agent": expectedUserAgent,
					}),
					ghttp.VerifyHeader(http.Header{
						"Accept": []string{"application/vnd.github.v3+json"},
					}),
				),
			)

		})

		It("should send requests with the expected headers", func() {

			lister.List(ctx, "")

			Expect(server.ReceivedRequests()).Should(HaveLen(2))
		})
	})

	Context("happy path", func() {
		BeforeEach(func() {
			branchesResponseBody := `{
	  "name": "master",
	  "commit": {
		"sha": "b0012e4930d0a8c350254a3caeedf7441ea286a3",
		"commit": {
		  "tree": {
			"sha": "5adf061bdde4dd26889be1e74028b2f54aabc346"
		  }
		}
	  }
	}`
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v3/repos/github/gitignore/branches/master"),
					ghttp.RespondWith(http.StatusOK, branchesResponseBody),
				),
			)
		})

		When("the tree response is empty", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/repos/github/gitignore/git/trees/5adf061bdde4dd26889be1e74028b2f54aabc346"),
						ghttp.RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			It("should return an empty slice", func() {
				ignoreFiles, _ := lister.List(ctx, "")
				Expect(ignoreFiles).Should(BeNil())
			})

			It("should not have an error", func() {
				_, err := lister.List(ctx, "")
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("the response has gitignore files", func() {
			BeforeEach(func() {
				responseBody := `{
  "sha": "5adf061bdde4dd26889be1e74028b2f54aabc346",
  "url": "https://api.github.com/repos/github/gitignore/git/trees/5adf061bdde4dd26889be1e74028b2f54aabc346",
  "tree": [
    {
      "path": "Actionscript.gitignore",
      "mode": "100644",
      "type": "blob",
      "sha": "5d947ca8879f8a9072fe485c566204e3c2929e80",
      "size": 350,
      "url": "https://api.github.com/repos/github/gitignore/git/blobs/5d947ca8879f8a9072fe485c566204e3c2929e80"
    },
    {
      "path": "Global/Anjuta.gitignore",
      "mode": "100644",
      "type": "blob",
      "sha": "20dd42c53e6f0df8233fee457b664d443ee729f4",
      "size": 78,
      "url": "https://api.github.com/repos/github/gitignore/git/blobs/20dd42c53e6f0df8233fee457b664d443ee729f4"
    },
    {
      "path": "community/AWS/SAM.gitignore",
      "mode": "100644",
      "type": "blob",
      "sha": "dc9d020aee1ebc1a23c02d80a1c33c0cb35ebaeb",
      "size": 167,
      "url": "https://api.github.com/repos/github/gitignore/git/blobs/dc9d020aee1ebc1a23c02d80a1c33c0cb35ebaeb"
    }
  ],
  "truncated": false
}
				}`
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/repos/github/gitignore/git/trees/5adf061bdde4dd26889be1e74028b2f54aabc346"),
						ghttp.RespondWith(http.StatusOK, responseBody),
					),
				)
			})

			It("should return a list of gitignore files", func() {
				ignoreFiles, _ := lister.List(ctx, "")
				Expect(ignoreFiles).Should(Equal(
					[]string{"Actionscript.gitignore", "Global/Anjuta.gitignore", "community/AWS/SAM.gitignore"}))
			})
		})
	})
})