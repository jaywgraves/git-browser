package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const VERSION = "0.0.2"

func main() {
	remoteName := "origin" // default to "origin"
	argv := os.Args[1:]
	show := false
	if len(argv) > 0 {
		if argv[0] == "--version" {
			fmt.Printf("%s\n", VERSION)
			os.Exit(0)
		}
		if argv[0] == "--show" {
			show = true
			if len(argv) > 1 {
				remoteName = argv[1]
			}
		} else if strings.HasPrefix(argv[0], "-") {
			fmt.Printf("invalid argument %s\n", argv[0])
			os.Exit(1)
		} else {
			remoteName = argv[0]
		}
	}
	// this will fail if no remote or not a git repo
	remote := gitRemote(remoteName)
	browserUrl := parseRemote(remote)
	if show {
		fmt.Printf("%s\n", browserUrl)

	} else {
		openBrowser(browserUrl)
	}
}

func gitRemote(remoteName string) string {
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error: `git remote` call failed: %v", err)
		os.Exit(1)
	}
	outStr := string(out)
	return strings.TrimRight(outStr, "\n")
}

func gitCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error: failed getting current branch: %v", err)
		os.Exit(1)
	}
	outStr := string(out)
	return strings.TrimRight(outStr, "\n")
}

func parseRemote(remoteUrl string) string {
	// sample ssh clone remote url
	// git@github.com:<some-profile/<some-repo>.git

	// sample https clone remote url
	// https://github.com/<some-profile/<some-repo>.git

	// should turn into
	// https://github.com/<some-profile/<some-repo>

	remoteUrl = strings.TrimSuffix(remoteUrl, ".git")

	if strings.HasPrefix(remoteUrl, "http") {
		// http(s) clone, nothing else to do
	} else if strings.HasPrefix(remoteUrl, "git@") {
		remoteUrl = strings.Replace(remoteUrl, ":", "/", -1)
		remoteUrl = strings.TrimPrefix(remoteUrl, "git@")
		remoteUrl = "https://" + remoteUrl
	}

	return remoteUrl
}

func codeHostFromUrl(remoteUrl string) string {
	// it's a bit unclear about when this will actually return
	// an err.  it's posslbe to just return an empty string for .Host
	url, err := url.Parse(remoteUrl)
	if err != nil {
		fmt.Printf("Error: failed parsing remoteUrl: %v", err)
		os.Exit(1)
	}
	return url.Host
}

func makeBranchUrl(remoteUrl string, branch string) string {
	codehost := codeHostFromUrl(remoteUrl)

	var urlFmtString string

	switch codehost {
	case "github.com":
		urlFmtString = "%s/tree/%s"
	case "gitlab.com":
		urlFmtString = "%s/-tree/%s"
	case "bitbucket.org":
		urlFmtString = "%s/src/%s/"
	default:
		// just return the remoteURL if we don't recognize the code host
		return remoteUrl
	}

	// this should not have a trailing / but make sure so the
	// fmt strings don't double up the /
	remoteUrl = strings.TrimSuffix(remoteUrl, "/")
	return fmt.Sprintf(urlFmtString, remoteUrl, branch)

}

func openBrowser(url string) {

	var cmd []string
	switch runtime.GOOS {
	case "linux":
		cmd = []string{"xdg-open"}
	case "darwin":
		cmd = []string{"open"}
	case "windows":
		cmd = []string{"cmd", "/c", "start"}
	default:
		fmt.Printf("Error: unable to start a browser session")
		os.Exit(1)
	}
	openCmd := exec.Command(cmd[0], append(cmd[1:], url)...)
	openCmd.Start()

}
