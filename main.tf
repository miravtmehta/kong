provider "aws" {
  region = "us-west-2"
}


resource "aws_instance" "web_server" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"


  vpc_security_group_ids = [aws_security_group.web_sg.id]
  key_name               = "prod-key"


  user_data = <<-EOF
    #!/bin/bash
    set -euo pipefail

    readonly install_script="/tmp/docker-install.sh"
    readonly install_script_sha256="fefa50ccd50efb42f438b506fc3a88574118f314aaf2a7cd5b6e1ffb1bffcf26"

    curl --fail --show-error --silent --location \
      --proto '=https' --tlsv1.2 \
      --output "$install_script" \
      "https://raw.githubusercontent.com/docker/docker-install/2b32480025b223ebfddae9a3a8bef09027680f53/install.sh"
    printf '%s  %s\n' "$install_script_sha256" "$install_script" | sha256sum --check --strict -
    /bin/sh "$install_script"
  EOF
  tags = {
    Name = "production-web-server"
  }
}


resource "aws_security_group" "web_sg" {
  name_prefix = "web-sg-"
  description = "Web server security group"


  ingress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }


  ingress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


resource "aws_s3_bucket" "app_data_bucket" {
  bucket = "my-app-data"
  acl    = "private"
  versioning {
    enabled = false
  }


  lifecycle_rule {
    id      = "data-cleanup"
    enabled = true
    expiration {
      days = 7
    }
    noncurrent_version_expiration {
      days = 1
    }
  }


  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}


resource "aws_s3_bucket_public_access_block" "app_data_bucket" {
  bucket = aws_s3_bucket.app_data_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}


resource "aws_db_instance" "app_database" {
  identifier                  = "app-db-instance"
  engine                      = "mysql"
  instance_class              = "db.t2.micro"
  allocated_storage           = 5
  username                    = "admin"
  manage_master_user_password = true
  publicly_accessible         = false


  backup_retention_period = 7
  multi_az                = false
}
