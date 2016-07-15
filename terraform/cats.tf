resource "template_file" "cats_containers" {
    template = "${file("container-definitions/cats.json.tmpl")}"

    vars {
        version = "${var.image_version}"
        appenv = "${format("catsenv-%s-us-west-2", var.environment)}"
        syslog_address = "${var.syslog_address}"
        syslog_tag = "${format("cats-%s", var.image_version)}"
    }
}

resource "aws_ecs_task_definition" "cats" {
    family = "cats"
    container_definitions = "${template_file.cats_containers.rendered}"
}

resource "aws_ecs_service" "cats" {
    name = "${format("cats-%s", var.environment)}"
    cluster = "${var.ecs_cluster}"
    task_definition = "${aws_ecs_task_definition.cats.arn}"
    desired_count = 2
    deployment_maximum_percent = 200
    deployment_minimum_healthy_percent = 100
    iam_role = "${var.ecs_iam_role}"

    load_balancer {
        elb_name = "cats"
        container_name = "cats"
        container_port = 9105
    }
}
