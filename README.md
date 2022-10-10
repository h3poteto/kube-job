[![Build](https://github.com/h3poteto/kube-job/workflows/Build/badge.svg)](https://github.com/h3poteto/kube-job/actions?query=workflow%3ABuild)
[![E2E](https://github.com/h3poteto/kube-job/workflows/E2E/badge.svg)](https://github.com/h3poteto/kube-job/actions?query=workflow%3AE2E)
[![GitHub release](http://img.shields.io/github/release/h3poteto/kube-job.svg?style=flat)](https://github.com/h3poteto/kube-job/releases)
[![GoDoc](https://godoc.org/github.com/h3poteto/kube-job/pkg/job?status.svg)](https://godoc.org/github.com/h3poteto/kube-job/pkg/job)
[![Renovate](https://img.shields.io/badge/renovate-enabled-brightgreen.svg)](https://renovatebot.com)

# kube-job

`kube-job` is a command line tool to run one off job on Kubernetes. The feature is

- Override args argument and docker image in a kubernetes job template, and run the job.
- Wait for completion of the job execution
- Get logs from kubernetes pods and output in stream

This is a command line tool, but you can use job as a package. So when you write own job execition script for Kubernetes, you can embed job package in your golang source code and customize task recipe. Please check the [godoc](https://godoc.org/github.com/h3poteto/kube-job/pkg/job).

## Install
```
$ wget https://github.com/h3poteto/kube-job/releases/download/v0.7.0/kube-job_v0.7.0_linux_amd64.zip
$ unzip kube-job_v0.7.0_linux_amd64.zip
$ ./kube-job help
Run one off job on kubernetes

Usage:
  kube-job [command]

Available Commands:
  help        Help about any command
  run         Run a job on Kubernetes
  version     Print the version number

Flags:
      --config KUBECONFIG   Kubernetes config file path (If you don't set it, use environment variables KUBECONFIG)
  -h, --help                help for kube-job
  -v, --verbose             Enable verbose mode

Use "kube-job [command] --help" for more information about a command.
```

## Usage
### kubeconfig file
At first, you have to prepare kuberenetes config file. If you are using kubectl, you can use the same file.

```yaml
apiVersion: v1
clusters:
- cluster:
    server: https://example.com
    certificate-authority-data: certificate
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: hoge
  name: hoge
current-context: hoge
kind: Config
preferences: {}
users:
# ...
```

### Job template
Next, please write job template file. It can be the same as used for kubectl, as below:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
  namespace: default
  labels:
    app: example-job
spec:
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: alpine
        image: alpine:latest
        imagePullPolicy: Always
        args: ["env"]
        env:
          - name: HOGE
            value: fuga
      restartPolicy: Never
  backoffLimit: 2

```

`metadata.name` and `spec.template.spec.containers[0].args`, `spec.template.spec.containers[0].image` are overrided when you use `kube-job`.

#### Why override name?
Kubernetes creates a job based on the job template yaml file, so if you use `kube-job` more than once at the same time, it is failed.
Because Kubernetes can not register duplicate job name.

But, I think the usecase of `kube-job`, then I think that sometimes users want to use `kube-job` more than once at the same time for another commands.
So it is necessary to make it possible to perform multiple activation.

As a solution, `kube-job` adds random string to the name of the job.


### Run a command

Please provide Kuberenetes config file, job template yaml file, and command.
The container parameter receives which container do you want to execute the command.

```
$ ./kube-job run --config=$HOME/.kube/config --template-file=./job.yaml --args="echo fuga" --container="alpine"
fuga
```

### Specify an URL as a template file

You can specify an URL as a template file, like this:

```
$ ./kube-job run --template-file=https://raw.githubusercontent.com/h3poteto/kube-job/master/example/job.yaml --args="echo fuga" --container="alpine"
```

If your template file is located in private repository, please export personal access token of GitHub. And please use an API URL endpoint.

```
$ export GITHUB_TOKEN=hogehogefugafuga
$ ./kube-job run --template-file=https://api.github.com/repos/h3poteto/kube-job/contents/example/job.yaml --args="echo fuga" --container="alpine"
```

## Role

The user to be executed needs the following role.

```
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: job-user-role
rules:
- apiGroups: [""]
  verbs: ["get", "list", "delete", "deletecollection"]
  resources: ["pods", "pods/log"]
- apiGroups: ["batch"]
  verbs: ["create", "get", "delete"]
  resources: ["jobs", "jobs/status"]
```

## License
The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
