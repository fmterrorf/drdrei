package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

const (
	gitURLPrefix = "git::"
)

var (
	refTagsRe          = regexp.MustCompile(`refs/tags/(.*)-(\d\.\d\.\d.*)`)
	gitRefRe           = regexp.MustCompile(`(.*)-(\d\.\d\.\d.*)`)
	defaultIgnorePaths = []string{".terraform", ".git"}
)

func main() {
	targetPath := os.Args[1:][0]
	dirs, err := getDirs(targetPath, defaultIgnorePaths)
	if err != nil {
		log.Fatalf("Failed to list dirs %v", err)
	}
	mcs := getExternalModuleCalls(dirs)
	gitURLKv := make(map[string]bool)
	for _, mc := range mcs {
		gitURLKv[mc.gitURL().repoURL()] = true
	}
	// Represents a map of
	// 	{
	//		[gitURL]: {
	//			[featureName]: latestVersion
	// 		}
	//	}
	latestTagByGitURL := make(map[string]map[string]string)
	for k := range gitURLKv {
		tags, err := fetchLatestTag(k)
		if err != nil {
			log.Fatalf("Failed to fetch latest tags %v", err)
		}
		latestTagByGitURL[k] = tags
	}
	found := false
	for _, mc := range mcs {
		gitURL := mc.gitURL()
		ref, err := gitURL.ref()
		if err != nil {
			log.Fatalf("Failed to parse git URL ref %v", err)
		}
		matches := gitRefRe.FindAllStringSubmatch(ref, -1)
		if len(matches) == 0 {
			continue
		}
		if !found {
			found = true
		}
		match := matches[0]
		feature := match[1]
		version := match[2]
		latestVer := latestTagByGitURL[gitURL.repoURL()][feature]
		if latestVer != version {
			fmt.Printf("%s %s:%d:0 \nusing: %s-%s \nlatest: %s-%s\n\n", mc.Name, mc.Pos.Filename, mc.Pos.Line, feature, version, feature, latestVer)
		}
	}

	if !found {
		fmt.Printf("You are all up to date")
	}
}

// getDirs collects all the children dir of sourceDir ignoring ignorePaths
func getDirs(sourceDir string, ignorePaths []string) ([]string, error) {
	dirs := make([]string, 0)
	ignorePathsKV := make(map[string]bool)
	for _, v := range ignorePaths {
		ignorePathsKV[v] = true
	}
	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if _, ok := ignorePathsKV[info.Name()]; ok {
			return filepath.SkipDir
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return dirs, nil
}

// getExternalModuleCalls collects all ModuleCalls that uses git:: prefix
func getExternalModuleCalls(dirs []string) []moduleCall {
	mcs := make([]moduleCall, 0)
	for _, dirname := range dirs {
		module, _ := tfconfig.LoadModule(dirname)
		for _, mc := range module.ModuleCalls {
			if strings.HasPrefix(mc.Source, gitURLPrefix) {
				mcs = append(mcs, moduleCall{mc})
			}
		}
	}
	return mcs
}

// fetchLatestTag returns a map of
// 	{
// 		[featureName]: latestVersion
// 	}
func fetchLatestTag(url string) (map[string]string, error) {
	output, err := exec.Command("git", "ls-remote", "--tags", "--ref", url).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	versionsByFeatures := make(map[string][]*semver.Version)
	for _, line := range lines {
		matches := refTagsRe.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		match := matches[0]
		feature := match[1]
		version := match[2]
		v, err := semver.NewVersion(version)
		if err != nil {
			return nil, err
		}
		if versions, ok := versionsByFeatures[feature]; ok {
			versionsByFeatures[feature] = append(versions, v)
		} else {
			versionsByFeatures[feature] = []*semver.Version{v}
		}
	}
	latestVersionByFeature := make(map[string]string)
	for k, v := range versionsByFeatures {
		c := semver.Collection(v)
		sort.Sort(c)
		latestVersionByFeature[k] = v[len(v)-1].Original()
	}

	return latestVersionByFeature, nil
}