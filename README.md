drdrie
------

A tool that helps you detect outdated module sources

**Note**

This tool only currently works if your module sources fits the following

* Uses a generic git url
* Versioned using a mono repo format of `<feature-name>-<semver>` e.g. `gcs-1.0.0`

## Install

Using go get 

    go get github.com/fmterrorf/drdrie

Precompiled executables

* See [releases](https://github.com/fmterrorf/drdrie/releases)

## Usage

    drdrie path/to/tfproject
