package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// terraformSourceURL represents generic git repository address https://www.terraform.io/docs/language/modules/sources.html#generic-git-repository
type terraformSourceURL string

func (u terraformSourceURL) repoURL() string {
	parsed, _ := url.Parse(u.trimProtocol())
	split := strings.Split(parsed.Path, "/")
	return fmt.Sprintf("git@%s:%s/%s", parsed.Host, split[1], split[2])
}

func (u terraformSourceURL) ref() (string, error) {
	parsed, err := url.Parse(u.trimProtocol())
	if err != nil {
		return "", err
	}
	return parsed.Query().Get("ref"), nil
}

func (u terraformSourceURL) trimProtocol() string {
	return strings.TrimPrefix(string(u), gitURLPrefix)
}

type moduleCall struct {
	*tfconfig.ModuleCall
}

func (mc moduleCall) gitURL() terraformSourceURL {
	return terraformSourceURL(mc.Source)
}
