package main

import (
	"flag"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	urlStr := flag.String("url", "http://localhost:8001/swagger.json", "URL to Kubernetes OpenAPI swagger JSON")
	pathK8s := flag.String("k8s", "k8s.libsonnet", "path output k8s.libsonnet")
	pathK := flag.String("k", "k.libsonnet", "path output k.libsonnet")
	force := flag.Bool("force", false, "force overwrite of existing output")
	enableTrace := flag.Bool("trace", false, "create trace output")
	profile := flag.String("profile", "", "create profile output")
	flag.Parse()

	if *urlStr == "" {
		logrus.Fatal("url is required")
	}

	if *enableTrace {
		trace.Start(os.Stdout)
		defer trace.Stop()
	}

	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			logrus.WithError(err).Fatal("create pprof output")
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *force {
		for _, path := range []string{*pathK8s, *pathK} {
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

	k8s, k, err := ksonnet.GenerateLib(*urlStr)
	if err != nil {
		logrus.WithError(err).Fatal("generate lib")
	}

	if err := writeFile(*pathK8s, k8s); err != nil {
		logrus.WithError(err).Fatal("write k8s.libsonnet")
	}
	if err := writeFile(*pathK, k); err != nil {
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
