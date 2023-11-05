#!/bin/bash

set -e

docker tag davidwashere/webecho:1.2 davidwashere/webecho:latest

docker push davidwashere/webecho:1.2
docker push davidwashere/webecho:latest