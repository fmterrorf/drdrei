drdrei
------

A tool that helps you detect outdated module sources

**Note**

This tool only currently works if your module sources fits the following

* Uses a [generic git repository](https://www.terraform.io/docs/language/modules/sources.html#generic-git-repository)
* The source module is using a tag with the following format `<feature-name>-<semver>` e.g. `gcs-1.0.0`

## Install

* Using go get 

        go get github.com/fmterrorf/drdrei

OR 

* See [releases](https://github.com/fmterrorf/drdrei/releases) for compiled executables

## Usage

```sh
  drdrei paths... [flags]

Flags:
      --config string             config file (default is $HOME/.drdrei.yaml)
  -h, --help                      help for drdrei
  -i, --ignorePaths stringArray   Path to ignore (default [.terraform,.git])
      --json                      Show result as JSON
  -r, --recursive                 Run audit recursively to given paths
```