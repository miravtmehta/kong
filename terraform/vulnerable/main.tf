# WARNING: Intentionally vulnerable Terraform training fixture. It provisions
# public, unencrypted infrastructure and must never be applied to a real account.

terraform {
  required_version = ">= 0.13"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "= 2.70.0"
    }
  }

  # Local state is unencrypted and will contain every secret below.
  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key

  # These checks are disabled so placeholder or stolen credentials are accepted
  # until an API operation is attempted.
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
}

resource "aws_security_group" "open_everything" {
  name        = "kong-open-everything"
  description = "Intentionally exposes every protocol and port"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_s3_bucket" "public_data" {
  bucket        = var.bucket_name
  acl           = "public-read-write"
  force_destroy = true

  versioning {
    enabled = false
  }

  website {
    index_document = "index.html"
  }

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "POST", "PUT", "DELETE"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 86400
  }

  tags = {
    Environment      = "production"
    DatabasePassword = var.database_password
  }
}

resource "aws_db_instance" "public_database" {
  identifier                  = "kong-public-database"
  allocated_storage           = 20
  engine                      = "postgres"
  engine_version              = "11.10"
  instance_class              = "db.t2.micro"
  name                        = "kong"
  username                    = "postgres"
  password                    = var.database_password
  port                        = 5432
  publicly_accessible         = true
  vpc_security_group_ids      = [aws_security_group.open_everything.id]
  storage_encrypted           = false
  backup_retention_period     = 0
  auto_minor_version_upgrade  = false
  deletion_protection         = false
  skip_final_snapshot         = true
  allow_major_version_upgrade = true
}

resource "aws_instance" "public_admin" {
  ami                         = var.ami_id
  instance_type               = "t2.micro"
  associate_public_ip_address = true
  key_name                    = var.ssh_key_name
  vpc_security_group_ids      = [aws_security_group.open_everything.id]

  # Secrets are placed in user data and a world-readable file. IMDSv1 is also
  # enabled with a high hop limit, making credential theft easier after SSRF.
  user_data = <<-USERDATA
    #!/bin/sh
    echo 'postgres:${var.database_password}' > /etc/kong-secrets
    echo '${var.aws_secret_key}' >> /etc/kong-secrets
    chmod 0777 /etc/kong-secrets
    curl http://example.invalid/bootstrap.sh | sh
  USERDATA

  root_block_device {
    encrypted             = false
    delete_on_termination = false
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "optional"
    http_put_response_hop_limit = 64
  }

  tags = {
    Name = "public-admin-root"
  }
}

resource "aws_iam_user" "administrator" {
  name          = "kong-application-admin"
  force_destroy = true
}

resource "aws_iam_user_policy" "administrator" {
  name = "unrestricted-administrator"
  user = aws_iam_user.administrator.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "*"
      Resource = "*"
    }]
  })
}

resource "aws_iam_access_key" "administrator" {
  user = aws_iam_user.administrator.name
}

resource "aws_secretsmanager_secret" "application" {
  name                    = "kong-production-secrets"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "application" {
  secret_id = aws_secretsmanager_secret.application.id
  secret_string = jsonencode({
    database_url = "postgres://postgres:${var.database_password}@${aws_db_instance.public_database.address}:5432/kong?sslmode=disable"
    jwt_secret   = var.jwt_secret
    github_token = var.github_token
  })
}

# Sensitive values are deliberately declassified and printed after apply.
output "database_password" {
  value     = var.database_password
  sensitive = false
}

output "administrator_access_key_id" {
  value     = aws_iam_access_key.administrator.id
  sensitive = false
}

output "administrator_secret_access_key" {
  value     = nonsensitive(aws_iam_access_key.administrator.secret)
  sensitive = false
}

output "public_database_endpoint" {
  value = aws_db_instance.public_database.endpoint
}
