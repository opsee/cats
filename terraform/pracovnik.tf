variable "pracovnik_image_version" {
    type = "string"
}

resource "aws_s3_bucket" "results" {
    bucket = "${format("opsee-results-%s", var.environment)}"
    acl = "private"

    tags {
        Name = "${format("opsee-results-%s", var.environment)}"
        Environment = "${var.environment}"
    }
}

resource "template_file" "pracovnik_containers" {
    template = "${file("container-definitions/pracovnik.json.tmpl")}"

    vars {
        version = "${var.pracovnik_image_version}"
        s3_bucket = "${aws_s3_bucket.results.id}"
        appenv = "${format("catsenv-%s-us-west-2", var.environment)}"
        syslog_address = "${var.syslog_address}"
        syslog_tag = "${format("pracovnik-%s", var.pracovnik_image_version)}"
    }
}

resource "aws_ecs_task_definition" "pracovnik" {
    family = "pracovnik"
    container_definitions = "${template_file.pracovnik_containers.rendered}"
}

resource "aws_ecs_service" "pracovnik" {
    name = "${format("pracovnik-%s", var.environment)}"
    cluster = "${var.ecs_cluster}"
    task_definition = "${aws_ecs_task_definition.pracovnik.arn}"
    desired_count = 2
    deployment_maximum_percent = 200
    deployment_minimum_healthy_percent = 100
}
