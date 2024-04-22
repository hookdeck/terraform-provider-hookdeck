---
page_title: "Getting Started"
description: "Getting Started with Hookdeck Provider"
---

# Getting Started with Hookdeck Provider

## Hookdeck

The Hookdeck Event Gateway enables engineering teams to build, deploy, observe, and scale event-driven applications.

Once integrated, Hookdeck unlocks an entire suite of tools: [alerting](https://hookdeck.com/docs/notifications), [rate limiting](https://hookdeck.com/docs/set-a-rate-limit), [automatic retries](https://hookdeck.com/docs/automatically-retry-events), [one-to-many delivery](https://hookdeck.com/docs/create-a-destination), [payload transformations](https://hookdeck.com/docs/transformations), local testing via the [CLI](https://hookdeck.com/docs/using-the-cli), a feature-rich [API](https://hookdeck.com/docs/using-the-api), and more. It acts as a proxy – routing webhooks from any [source](https://hookdeck.com/docs/sources) to a specified [destination](destinations).

Visit the [Hookdeck documentation](https://hookdeck.com/docs/introduction) to learn more.

## Terraform

[Terraform](https://developer.hashicorp.com/terraform/intro) is an open-source Infrastructure as Code (IaC) tool that allows you to define and manage infrastructure resources using HashiCorp Configuration Language (HCL). It can be used to manage a wide range of resources, including servers, storage, networks, and cloud services. Terraform is a popular choice for infrastructure automation because it is easy to use, flexible, and powerful.

The Hookdeck Terraform provider helps you utilize Terraform to configure your workspace declaratively instead of relying on the Hookdeck dashboard. You can run Terraform in your CI/CD pipeline and maintain Hookdeck workspace configuration programmatically as part of your deployment workflow.


## Tutorial

Before you begin, make sure you have [Terraform CLI](https://developer.hashicorp.com/terraform/downloads) installed locally and a Hookdeck API Key obtained from [the dashboard](https://dashboard/hookdeck.com/workspace/secrets).

### Initialize Terraform

In a directory of your choice, create a Terraform config file `main.tf`:

```hcl
# main.tf

terraform {
  required_providers {
    hookdeck = {
      source = "hookdeck/hookdeck"
      version = "~> 0.1"
    }
  }
}

provider "hookdeck" {
  api_key = "<YOUR_API_KEY>"
}
```

Replace `<YOUR_API_KEY>` with your Hookdeck workspace API key.

After creating your basic configuration in HCL, initialize Terraform and ask it to apply the configuration to Hookdeck.

```sh
$ terraform init
```

Running `terraform init` will download the plugins required in the configuration file, such as the Hookdeck Terraform provider, to a local `.terraform` directory.

Afterward, you can run `terraform plan` to confirm that you have Terraform properly installed. As you haven't added any resources for Terraform to manage yet, it will indicate that there are no planned changes with your infrastructure's current state.

```
$ terraform plan
```

### Source

First, let's create a source resource with Terraform. You can add this resource block to the end of your Terraform configuration file

```hcl
resource "hookdeck_source" "my_source" {
  name = "my_source"
}
```

Now, try `terraform plan` again to see what Terraform may suggest

```sh
$ terraform plan
```

```
Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # hookdeck_source.my_source will be created
  + resource "hookdeck_source" "my_source" {
      + allowed_http_methods = [
          + "POST",
          + "PUT",
          + "PATCH",
          + "DELETE",
        ]
      + created_at           = (known after apply)
      + disabled_at          = (known after apply)
      + id                   = (known after apply)
      + name                 = "my_source"
      + team_id              = (known after apply)
      + updated_at           = (known after apply)
      + url                  = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.

─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't guarantee to take exactly these actions if you run "terraform apply" now.
```

You can execute this change and create a source in your workspace like so.

```sh
$ terraform apply
```

You can check the dashboard to confirm that a new source was created in your workspace.

To learn more about what options you have with Hookdeck's Source on Terraform, check out the [`hookdeck_source` docs](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/resources/source).

### Destination

Next, add a simple destination resource:

```hcl
resource "hookdeck_destination" "my_destination" {
  name = "my_destination"
  url  = "https://mock.hookdeck.com"
}
```

This is a Mock destination which will accepts all of your events so you can inspect on Hookdeck's dashboard.

Now, run `terraform apply` to create your new destination. Terraform will show the plan and ask for your confirmation before executing it, so you don't need to run `terraform plan` beforehand.

To learn more about what options you have with Hookdeck's Destination on Terraform, check out the [`hookdeck_destination` docs](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/resources/destination).

### Connection

Lastly, define a Hookdeck connection to connect the source and destination:

```hcl
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
}
```

As before, run `terraform apply` to review the plan and apply the changes.

To learn more about what options you have with Hookdeck's Connection on Terraform, check the [`hookdeck_connection` docs](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/resources/connection).

### Summary

In this tutorial, you have:

- Installed Terraform CLI locally and initialized a Terraform project with Hookdeck provider with `terraform init`
- Written the configuration code for a Hookdeck source, destination, and connection using Terraform's declarative programming language, HCL
- Reviewed and executed the Terraform plan with `terraform plan` and `terraform apply`

Here's the final `main.tf` file:

```
terraform {
  required_providers {
    hookdeck = {
      source  = "hookdeck/hookdeck"
      version = "~> 0.1"
    }
  }
}

provider "hookdeck" {
  api_key = "<YOUR_API_KEY>"
}

resource "hookdeck_source" "my_source" {
  name = "my_source"
}

resource "hookdeck_destination" "my_destination" {
  name = "my_destination"
  url  = "https://mock.hookdeck.com"
}

resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
}
```
