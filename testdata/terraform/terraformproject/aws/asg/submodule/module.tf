resource "random_string" "password" {
  count   = var.module_enabled ? 1 : 0
  length  = 16
  special = true
}

data "template_file" "startup-script" {
  template = file("${path.module}/files/${var.service_name}_bootstrap.sh")

  vars = {
    mmsGroupId = var.mmsGroupId
    mmsApiKey  = var.mmsApiKey
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name = "name"

    values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

module "asg" {
  create_lc                    = var.module_enabled
  create_asg                   = var.module_enabled
  source                       = "terraform-aws-modules/autoscaling/aws"
  version                      = "~> 3.0"
  name                         = "${var.deploy_name}-${var.region}-${var.service_name}"
  lc_name                      = "${var.deploy_name}-${var.region}-${var.service_name}"
  image_id                     = var.asg_specified_ami == "" ? data.aws_ami.ubuntu.id : var.asg_specified_ami
  instance_type                = var.machine_type
  security_groups              = aws_security_group.default.*.id
  associate_public_ip_address  = var.public_ip
  recreate_asg_when_lc_changes = true
  user_data                    = data.template_file.startup-script.rendered
  key_name                     = var.key_name

  ebs_block_device = [
    {
      device_name           = "/dev/xvdz"
      volume_type           = "gp2"
      volume_size           = "100"
      delete_on_termination = true
    },
  ]

  root_block_device = [
    {
      volume_size           = "50"
      volume_type           = "gp2"
      delete_on_termination = true
    },
  ]

  # Auto scaling group
  asg_name                  = "${var.deploy_name}-${var.region}-${var.service_name}"
  vpc_zone_identifier       = var.subnets
  health_check_type         = "EC2"
  min_size                  = var.instance_count
  max_size                  = var.instance_count
  desired_capacity          = var.instance_count
  wait_for_capacity_timeout = 0

  tags = [
    {
      key                 = "Environment"
      value               = "dev"
      propagate_at_launch = true
    },
    {
      key                 = "Project"
      value               = "megasecret"
      propagate_at_launch = true
    },
  ]

  tags_as_map = {
    extra_tag1 = "extra_value1"
    extra_tag2 = "extra_value2"
  }
}

resource "aws_security_group" "default" {
  count  = var.module_enabled ? 1 : 0
  name   = "${var.deploy_name}-${var.region}-${var.service_name}"
  vpc_id = var.vpc_id

  ingress {
    from_port = var.port
    to_port   = var.port
    protocol  = "tcp"
    # TF-UPGRADE-TODO: In Terraform v0.10 and earlier, it was sometimes necessary to
    # force an interpolation expression to be interpreted as a list by wrapping it
    # in an extra set of list brackets. That form was supported for compatibility in
    # v0.11, but is no longer supported in Terraform v0.12.
    #
    # If the expression in the following list itself returns a list, remove the
    # brackets to avoid interpretation as a list of lists. If the expression
    # returns a single list item then leave it as-is and remove this TODO comment.
    security_groups = [var.security_groups[0]]
    self            = true
  }

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    # TF-UPGRADE-TODO: In Terraform v0.10 and earlier, it was sometimes necessary to
    # force an interpolation expression to be interpreted as a list by wrapping it
    # in an extra set of list brackets. That form was supported for compatibility in
    # v0.11, but is no longer supported in Terraform v0.12.
    #
    # If the expression in the following list itself returns a list, remove the
    # brackets to avoid interpretation as a list of lists. If the expression
    # returns a single list item then leave it as-is and remove this TODO comment.
    security_groups = [var.security_groups[1]]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.deploy_name}-${var.region}-${var.service_name}"
  }
}

