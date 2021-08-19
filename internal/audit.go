package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

const gitURLPrefix = "git::"

type auditResult struct {
	Pos            tfconfig.SourcePos
	FeatureName    string
	CurrentVersion string
	LatestVersion  string
}

func (ar auditResult) isUsingLatestVersion() bool {
	return ar.CurrentVersion == ar.LatestVersion
}

func printToConsole(results []auditResult, printAsJSON bool) error {
	if printAsJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		if err := enc.Encode(results); err != nil {
			return err
		}
		return nil
	}
	for _, ar := range results {
		fmt.Println(ar)
	}
	return nil
}

func (ar auditResult) String() string {
	return fmt.Sprintf("%s %s:%d:0 \nusing: %s-%s \nlatest: %s-%s\n\n",
		ar.FeatureName, ar.Pos.Filename, ar.Pos.Line, ar.FeatureName,
		ar.CurrentVersion, ar.FeatureName, ar.LatestVersion)
}

func (ar auditResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Feature string `json:"feature"`
		Pos     string `json:"pos"`
		Current string `json:"current"`
		Latest  string `json:"latest"`
	}{
		Feature: ar.FeatureName,
		Pos:     fmt.Sprintf("%s:%d", ar.Pos.Filename, ar.Pos.Line),
		Current: ar.CurrentVersion,
		Latest:  ar.LatestVersion,
	})
}

func RunAudit(targetPaths []string, recursive bool, ignorePaths []string, printAsJSON bool) {
	paths := uniqString(targetPaths)
	dirs := make([]string, 0)
	if recursive {
		for _, p := range paths {
			found, err := walkDirs(p, ignorePaths)
			if err != nil {
				log.Fatalf("Failed to list dirs %v", err)
			}
			dirs = append(dirs, found...)
		}
	} else {
		dirs = paths
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
	results := make([]auditResult, 0)
	for _, mc := range mcs {
		gitURL := mc.gitURL()
		ref, err := gitURL.ref()
		if err != nil {
			log.Fatalf("Failed to parse git URL ref %v", err)
		}
		feat := parseToFeature(ref, gitRefRe)
		if feat.isEmpty() {
			continue
		}
		ar := auditResult{
			Pos:            mc.Pos,
			FeatureName:    feat.Name,
			CurrentVersion: feat.Version,
			LatestVersion:  latestTagByGitURL[gitURL.repoURL()][feat.Name],
		}
		if ar.isUsingLatestVersion() {
			results = append(results, ar)
		}
	}
	if len(results) > 0 {
		printToConsole(results, printAsJSON)
		return
	}
	fmt.Printf("You are all up to date")
}

// walkDirs collects all the children dir of sourceDir ignoring ignorePaths
func walkDirs(sourceDir string, ignorePaths []string) ([]string, error) {
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
		feat := parseToFeature(line, refTagsRe)
		if feat.isEmpty() {
			continue
		}
		v, err := semver.NewVersion(feat.Version)
		if err != nil {
			return nil, err
		}
		if versions, ok := versionsByFeatures[feat.Name]; ok {
			versionsByFeatures[feat.Name] = append(versions, v)
		} else {
			versionsByFeatures[feat.Name] = []*semver.Version{v}
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
