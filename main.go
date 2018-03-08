package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	k8s        = "k8s.libsonnet"
	k          = "k.libsonnet"
	defaultURL = "http://localhost:8001/swagger.json"
)

func main() {
	var urlStr string
	flag.StringVar(&urlStr, "url", "", "URL to Kubernetes OpenAPI swagger JSON. Mutually exclusive with tag.")
	var tag string
	flag.StringVar(&tag, "tag", "", "Kubernetes tag to generate for. Mutually exclusive with url")
	var outputDir string
	flag.StringVar(&outputDir, "output", ".", "Directory to generate output")
	var force bool
	flag.BoolVar(&force, "force", false, "force overwrite of existing output")
	var traceFile string
	flag.StringVar(&traceFile, "trace", "", "create trace output")
	var profileFile string
	flag.StringVar(&profileFile, "profile", "", "create profile output")
	flag.Parse()

	var swaggerPath string

	switch {
	case urlStr == "" && tag == "":
		swaggerPath = defaultURL
	case urlStr != "":
		swaggerPath = urlStr
	case tag != "":
		path, err := genFromTag(tag)
		if err != nil {
			logrus.WithError(err).WithField("tag", tag).Fatal("fetch swagger from tag")
		}
		swaggerPath = path
		defer os.Remove(swaggerPath)
	case tag != "" && urlStr != "":
		logrus.Fatal("can't supply tag with url")
	}

	if traceFile != "" {
		f, err := os.Create(traceFile)
		if err != nil {
			logrus.WithError(err).Fatal("create trace output")
		}
		trace.Start(f)
		defer trace.Stop()
	}

	if profileFile != "" {
		f, err := os.Create(profileFile)
		if err != nil {
			logrus.WithError(err).Fatal("create pprof output")
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if force {
		for _, name := range []string{k8s, k} {
			path := filepath.Join(outputDir, name)
			logger := logrus.WithField("path", path)
			b, err := checkFile(path)
			if err != nil {
				logger.WithError(err).Fatal("check file")
			}
			if b {
				logger.Fatal("file exists")
			}
		}
	}

	lib, err := ksonnet.GenerateLib(swaggerPath)
	if err != nil {
		logrus.WithError(err).Fatal("generate lib")
	}

	if err := writeFile(filepath.Join(outputDir, k8s), lib.K8s); err != nil {
		logrus.WithError(err).Fatal("write k8s.libsonnet")
	}
	if err := writeFile(filepath.Join(outputDir, k), lib.Extensions); err != nil {
		logrus.WithError(err).Fatal("write k.libsonnet")
	}
}

func writeFile(path string, b []byte) error {
	err := ioutil.WriteFile(path, b, 0644)
	return errors.Wrapf(err, "write %q", path)
}

func checkFile(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, errors.Wrapf(err, "stat path %s", path)
	}

	return true, nil
}

func genFromTag(tag string) (string, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "raw.githubusercontent.com",
		Path:   fmt.Sprintf("/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json", tag),
	}

	resp, err := http.DefaultClient.Get(u.String())
	if err != nil {
		return "", errors.Wrapf(err, "fetch %s", u.String())
	}

	defer resp.Body.Close()

	file, err := ioutil.TempFile("", "swagger.json")
	if err != nil {
		return "", errors.Wrap(err, "create temp file")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read swagger contents")
	}

	if _, err := file.Write(b); err != nil {
		return "", errors.Wrap(err, "write swagger contents")
	}

	if err := file.Close(); err != nil {
		return "", errors.Wrap(err, "close file")
	}

	return file.Name(), nil
}
