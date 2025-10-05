# Terraform configuration
terraform {
  required_version = ">= 1.0"

  required_providers {
    zitadel = {
      source  = "zitadel/zitadel"
      version = "~> 0.1"
    }
  }

  # Local backend for development
  # Production'da S3 veya Terraform Cloud kullanın
  backend "local" {
    path = "terraform.tfstate"
  }
}