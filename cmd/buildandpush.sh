#!/bin/bash
harbor=docker.io
appName=glogstash
version=v1.0.0
repo=jinyidong
imageName=${harbor}/${repo}/${appName}:${version}

echo ---------------docker build ...----------------
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./${appName} ../${appName}.go

docker build -t ${imageName} .

docker login -u jinyidong -p jyd051060

docker push ${imageName}