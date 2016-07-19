provider "aws" {}

resource "aws_security_group" "results_cache" {
    name = "results elasticache"
    description = "Allow access from the ECS cluster to elasticache"
    vpc_id = "${var.vpc_id}"

    ingress {
        from_port = 11211
        to_port = 11211
        protocol = "tcp"
        security_groups = ["${var.ecs_cluster_sg}", "${var.vpn_sg}"]
    }

    egress {
        from_port = 0
        to_port = 0
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_subnet_group" "results_cache" {
    name = "results-cache-subnets"
    description = "subnets available for results cache cluster"
    subnet_ids = ["${split(",", var.vpc_subnet_ids)}"]
}

resource "aws_elasticache_parameter_group" "results_cache" {
    name = "results-cache-params"
    family = "memcached1.4"
    description = "memcached settings for the results cache"

    ## Results specific settings

    # The maximum size of a value in memcached. If we consider that check
    # responses are 128k max, then this means that we can hold at most 8
    # check responses in a single result.
    parameter {
        name = "max_item_size"
        value = "4194304"
    }

    parameter {
        name = "cas_disabled"
        value = "0"
    }

    parameter {
        name = "chunk_size"
        value = "48"
    }

    parameter {
        name = "chunk_size_growth_factor"
        value = "1.25"
    }

    parameter {
        name = "disable_flush_all"
        value = "0"
    }

    parameter {
        name = "error_on_memory_exhausted"
        value = "1"
    }
}

resource "aws_elasticache_cluster" "results_cache" {
    cluster_id = "results-cache"
    engine = "memcached"
    node_type = "cache.t2.small"
    port = 11211
    num_cache_nodes = 3
    parameter_group_name = "${aws_elasticache_parameter_group.results_cache.name}"
    subnet_group_name = "${aws_elasticache_subnet_group.results_cache.name}"
    security_group_ids = ["${aws_security_group.results_cache.id}"]
    az_mode = "cross-az"

    tags {
        Name = "results cache"
        Environment = "${var.environment}"
    }
}