terraform {
  cloud {
    organization = "gtis"

    workspaces {
      name = "cloudflare-scanner"
    }
  }
}
