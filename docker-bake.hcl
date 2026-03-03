variable "REGISTRY" {
  default = "docker.io"
}

variable "IMAGE_PREFIX" {
  default = "messaging"
}

variable "VERSION" {
  default = "latest"
}

group "default" {
  targets = ["dev"]
}

group "dev" {
  targets = ["server-dev", "client-dev"]
}

group "prod" {
  targets = ["server-prod", "client-prod"]
}

target "server-dev" {
  context = "."
  dockerfile = "Dockerfile.server"
  target = "dev"
  tags = [
    "${REGISTRY}/${IMAGE_PREFIX}-server:dev",
    "${REGISTRY}/${IMAGE_PREFIX}-server:dev-${VERSION}"
  ]
  platforms = ["linux/amd64"]
  cache-from = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-server:buildcache-dev"]
  cache-to = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-server:buildcache-dev,mode=max"]
}

target "server-prod" {
  context = "."
  dockerfile = "Dockerfile.server"
  target = "prod"
  tags = [
    "${REGISTRY}/${IMAGE_PREFIX}-server:prod",
    "${REGISTRY}/${IMAGE_PREFIX}-server:${VERSION}",
    "${REGISTRY}/${IMAGE_PREFIX}-server:latest"
  ]
  platforms = ["linux/amd64", "linux/arm64"]
  cache-from = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-server:buildcache-prod"]
  cache-to = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-server:buildcache-prod,mode=max"]
}

target "client-dev" {
  context = "."
  dockerfile = "Dockerfile.client"
  target = "dev"
  tags = [
    "${REGISTRY}/${IMAGE_PREFIX}-client:dev",
    "${REGISTRY}/${IMAGE_PREFIX}-client:dev-${VERSION}"
  ]
  platforms = ["linux/amd64"]
  cache-from = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-client:buildcache-dev"]
  cache-to = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-client:buildcache-dev,mode=max"]
}

target "client-prod" {
  context = "."
  dockerfile = "Dockerfile.client"
  target = "prod"
  tags = [
    "${REGISTRY}/${IMAGE_PREFIX}-client:prod",
    "${REGISTRY}/${IMAGE_PREFIX}-client:${VERSION}",
    "${REGISTRY}/${IMAGE_PREFIX}-client:latest"
  ]
  platforms = ["linux/amd64", "linux/arm64"]
  cache-from = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-client:buildcache-prod"]
  cache-to = ["type=registry,ref=${REGISTRY}/${IMAGE_PREFIX}-client:buildcache-prod,mode=max"]
}
