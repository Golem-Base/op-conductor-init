variable "REGISTRY" {
  default = "docker.io"
}

variable "REPOSITORY" {
  default = "golemnetwork"
}

variable "IMAGE" {
  default = "op-conductor-init"
}

variable "TAG" {
  default = "latest"
}

variable "PLATFORMS" {
  default = ["linux/amd64"]
}

target "_common" {
  context    = "."
  dockerfile = "Dockerfile"
  platforms  = PLATFORMS
  args = {
    VERSION    = ""
    GIT_COMMIT = ""
    GIT_DATE   = ""
  }
}

target "default" {
  inherits = ["_common"]
  tags     = ["${REGISTRY}/${REPOSITORY}/${IMAGE}:${TAG}"]
}

target "dev" {
  inherits = ["_common"]
  tags     = ["${REGISTRY}/${REPOSITORY}/${IMAGE}:dev"]
}

target "release" {
  inherits = ["_common"]
  tags = [
    "${REGISTRY}/${REPOSITORY}/${IMAGE}:${TAG}",
    "${REGISTRY}/${REPOSITORY}/${IMAGE}:latest"
  ]
}
