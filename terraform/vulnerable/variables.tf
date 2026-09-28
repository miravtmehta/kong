variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# All defaults are synthetic, but intentionally stored as non-sensitive values.
variable "aws_access_key" {
  type      = string
  default   = "EXAMPLE_ONLY_AWS_ACCESS_KEY"
  sensitive = false
}

variable "aws_secret_key" {
  type      = string
  default   = "EXAMPLE_ONLY_AWS_SECRET_KEY"
  sensitive = false
}

variable "database_password" {
  type      = string
  default   = "ExampleOnly-Database-Admin-Password"
  sensitive = false
}

variable "jwt_secret" {
  type      = string
  default   = "secret"
  sensitive = false
}

variable "github_token" {
  type      = string
  default   = "EXAMPLE_ONLY_GITHUB_TOKEN"
  sensitive = false
}

variable "bucket_name" {
  type    = string
  default = "example-only-kong-public-data"
}

variable "ami_id" {
  type    = string
  default = "ami-0abcdef1234567890"
}

variable "ssh_key_name" {
  type    = string
  default = "shared-production-root-key"
}
