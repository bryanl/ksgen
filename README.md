# ksgen

Generate `k8s.libsonnet` and `k.libsonnet` given a path to a Kubernetes swagger file.

```text
Usage of ksgen:
  -force
    	force overwrite of existing output
  -k string
    	path output k.libsonnet (default "k.libsonnet")
  -k8s string
    	path output k8s.libsonnet (default "k8s.libsonnet")
  -profile string
    	create profile output
  -trace
    	create trace output
  -url string
    	URL to Kubernetes OpenAPI swagger JSON (default "http://localhost:8001/swagger.json")
 ```
