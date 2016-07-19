variable "image_version" {
    type = "string"
}

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

variable "ecs_cluster_sg" {
    type = "string"
    default = "sg-82d5a7e6"
}

variable "vpn_sg" {
    type = "string"
    default = "sg-ded4a6ba"
}

variable "vpc_subnet_ids" {
    type = "string"
    default = "subnet-b233aeeb,subnet-58f9dd3d,subnet-50760c27"
}

variable "vpc_id" {
    type = "string"
    default = "vpc-b5f3a4d0"
}