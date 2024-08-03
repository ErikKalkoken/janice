// package github contains features for accessing repos on Github.
package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-version"
)

var ErrHttpError = errors.New("HTTP error")

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// AvailableUpdate returns the version of the latest release and reports wether the update is newer.
func AvailableUpdate(owner, repo, current string) (string, bool, error) {
	v1, err := version.NewVersion(current)
	if err != nil {
		return "", false, err
	}
	latest, err := fetchLatest(owner, repo)
	if err != nil {
		return "", false, err
	}
	v2, err := version.NewVersion(latest)
	if err != nil {
		return "", false, err
	}
	isNewer := v1.LessThan(v2)
	return latest, isNewer, nil
}

func fetchLatest(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	if r.StatusCode >= 400 {
		return "", fmt.Errorf("%s: %w", r.Status, ErrHttpError)
	}
	var info githubRelease
	if err := json.Unmarshal(data, &info); err != nil {
		return "", err
	}
	return info.TagName, nil
}
