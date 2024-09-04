package selfupdate

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v64/github"
	"golang.org/x/oauth2"

	"github.com/joanlopez/go-selfupdate-poc/kit/gitconfig"
)

func listReleases(
	ctx context.Context, owner, repo string, opts *github.ListOptions,
) ([]*github.RepositoryRelease, *github.Response, error) {
	client := githubClient()
	releases, response, err := client.Repositories.ListReleases(ctx, owner, repo, opts)

	// By default, we try to use the GitHub API with the token loaded from
	// the environment (or .gitconfig), in order to avoid rate limiting.
	//
	// However, a user may have a token that is either invalid or expired.
	//
	// So, in case the GitHub API returns a "401 Unauthorized", we try again
	// but using the default client, with no token.
	// In theory, it shouldn't reach the rate limit, but it's a possibility.
	var gErr *github.ErrorResponse
	if err != nil && errors.As(err, &gErr) && gErr.Response.StatusCode == http.StatusUnauthorized {
		client = github.NewClient(httpClient(ctx, ""))
		releases, response, err = client.Repositories.ListReleases(ctx, owner, repo, opts)
	}
	return releases, response, err
}

func downloadReleaseAsset(
	ctx context.Context, owner, repo string, id int64, followRedirectsClient *http.Client,
) (rc io.ReadCloser, redirectURL string, err error) {
	client := githubClient()
	asset, url, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, id, followRedirectsClient)

	// Same as with [listReleases], we try to use the GitHub API without the token.
	var gErr *github.ErrorResponse
	if err != nil && errors.As(err, &gErr) && gErr.Response.StatusCode == http.StatusUnauthorized {
		client = github.NewClient(httpClient(ctx, ""))
		asset, url, err = client.Repositories.DownloadReleaseAsset(ctx, owner, repo, id, followRedirectsClient)
	}
	return asset, url, err
}

func githubClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// We ignore the error, because it's not a big deal if we can't get the token.
		// In the worst case, the token will remain empty and the user will be rate limited.
		token, _ = gitconfig.GithubToken()
	}

	client := httpClient(context.Background(), token)
	return github.NewClient(client)
}

func httpClient(ctx context.Context, token string) *http.Client {
	if token == "" {
		return http.DefaultClient
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, src)
}
