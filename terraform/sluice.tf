resource "template_file" "sluice_containers" {
    template = "${file("container-definitions/sluice.json.tmpl")}"

    vars {
        version = "${var.image_version}"
        appenv = "${format("catsenv-%s-us-west-2", var.environment)}"
        syslog_address = "${var.syslog_address}"
        syslog_tag = "${format("sluice-%s", var.image_version)}"
    }
}

resource "aws_ecs_task_definition" "sluice" {
    family = "sluice"
    container_definitions = "${template_file.sluice_containers.rendered}"
}

resource "aws_ecs_service" "sluice" {
    name = "${format("sluice-%s", var.environment)}"
    cluster = "${var.ecs_cluster}"
    task_definition = "${aws_ecs_task_definition.sluice.arn}"
    desired_count = 1
    deployment_maximum_percent = 200
    deployment_minimum_healthy_percent = 100
    
    load_balancer {
        elb_name = "sluice"
        container_name = "sluice"
        container_port = 9107
    }
}
