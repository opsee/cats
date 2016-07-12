provider "aws" {}

variable "environment" {
    type = "string"
    default = "production"
}

variable "ecs_cluster" {
    type = "string"
}

variable "ecs_iam_role" {
    type = "string"
}

variable "syslog_address" {
    type = "string"
    default = "tcp+tls://logs3.papertrailapp.com:51722"
}