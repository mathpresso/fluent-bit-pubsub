# fluent-bit output plugin for google pubsub

<p align="left">    
  <a href="https://circleci.com/gh/gjbae1212/fluent-bit-pubsub/tree/master"><img src="https://circleci.com/gh/gjbae1212/fluent-bit-pubsub/tree/master.svg?style=svg"/></a>
  <a href="https://hits.seeyoufarm.com"/><img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fgjbae1212%2Ffluent-bit-pubsub"/></a>
  <a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-GREEN.svg" alt="license" /></a>
  <a href="https://goreportcard.com/report/github.com/gjbae1212/fluent-bit-pubsub"><img src="https://goreportcard.com/badge/github.com/gjbae1212/fluent-bit-pubsub" alt="Go Report Card" /></a>
  <a href="https://codecov.io/gh/gjbae1212/fluent-bit-pubsub"><img src="https://codecov.io/gh/gjbae1212/fluent-bit-pubsub/branch/master/graph/badge.svg"/></a>        
</p>

This plugin is used to publish data to queue in google pubsub. 

You could easily use it.

## Build
A bin directory already has been made binaries for mac, linux.

If you should directly make binaries for mac, linux
```bash
# local machine binary
$ bash make.sh build

# Your machine is mac, and if you should do to retry cross compiling for linux.
# A command in below is required a docker.  
$ bash make.sh build_linux

# qanda-for-fluent-bit
$ docker build -f docker/Dockerfile . -t qanda-for-fluent-bit:beta
```


## Usage
### configuration options for fluent-bit.conf
| Key           | Description                                    | Default        |
| ----------------|------------------------------------------------|----------------|
| Project         | google cloud project id | NONE(required) |
| Topic           | google pubsub topic name | NONE(required) |
| JwtPath         | jwt file path for accessible google cloud project | NONE(required) |
| Debug           | print debug log | false(optional) |
| Timeout         | the maximum time that the client will attempt to publish a bundle of messages. (millsecond) | 60000 (optional)|
| DelayThreshold  | publish a non-empty batch after this delay has passed. (millsecond) | 1  |
| ByteThreshold   | publish a batch when its size in bytes reaches this value. | 1000000 |
| CountThreshold  | publish a batch when it has been reached count of messages. | 100  |

### Example fluent-bit.conf
```conf
[Output]
    Name pubsub
    Match *
    Project your-project(custom)
    Topic your-topic-name(custom)
    Jwtpath your-jwtpath(custom)    
```

### Example exec
```bash
$ fluent-bit -c [your config file] -e pubsub.so 
```

### GCP Cloudbuild 배포 ###
1. 이미지 배포는 GCP CloudBuild를 통해서 배포하고 Cloudbuild Trigger는 다음과 같습니다.
[GCP CloudBduil Trigger](https://console.cloud.google.com/cloud-build/triggers;region=global/edit/bd531b25-6238-4c3f-82f7-87f615ab3322?invt=AbuM7A&project=qanda-dev-bakery-f5e1&supportedpurview=project)

2. 빌드 후 Artifact-registry 저장됩니다.
[Fluent-Bit 이미지](https://console.cloud.google.com/artifacts/docker/mp-artifact-registry-aa49/asia-northeast3/devops/qanda%2Ffluent-bit?invt=AbuM7g&project=mp-artifact-registry-aa49&supportedpurview=project)

3. Melon Fluent-Bit 배포 예시 yaml 정보
버전 upgrade 시 tag 정보로 업그레이드 합니다.
```conf
        fluentbit:
          enabled: true
          tag: 3.2.10-16ae37a
          mountPaths:
          - /tmp/data-pipeline
          infos:
          - logPath: /tmp/data-pipeline/data-pipeline.log
            pubsub_topic: qanda_log
            project_id: qp-data-serverlog-0a34
          resources:
            requests:
              cpu: 10m
              memory: 128Mi
            limits:
              memory: 256Mi
```
