terraform {
  required_providers {
    sodium = {
      source = "vdbe/sodium"
    }
    github = {
      source = "integrations/github"
    }
  }
}

variable "repo_name" {
  type    = string
  default = "example"
}

data "github_actions_public_key" "example" {
  repository = var.repo_name
}

resource "sodium_encrypt" "example" {
  public_key_base64 = data.github_actions_public_key.example.key
  value             = "secret"
}

resource "github_actions_secret" "exmaple" {
  repository      = var.repo_name
  secret_name     = "example"
  encrypted_value = resource.sodium_encrypt.example.encrypted_base64
}

