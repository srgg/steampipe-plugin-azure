
variable "resource_name" {
  type        = string
  default     = "turbot-test-20200125-create-update"
  description = "Name of the resource used throughout the test."
}

variable "azure_environment" {
  type        = string
  default     = "public"
  description = "Azure environment used for the test."
}

variable "azure_subscription" {
  type        = string
  default     = "3510ae4d-530b-497d-8f30-53b9616fc6c1"
  description = "Azure subscription used for the test."
}

provider "azurerm" {
  # Cannot be passed as a variable
  environment     = var.azure_environment
  subscription_id = var.azure_subscription
  features {}
}

data "azurerm_client_config" "current" {}

data "null_data_source" "resource" {
  inputs = {
    scope = "azure:///subscriptions/${data.azurerm_client_config.current.subscription_id}"
  }
}

resource "azurerm_resource_group" "named_test_resource" {
  name     = var.resource_name
  location = "East US"
}

resource "azurerm_kusto_cluster" "named_test_resource" {
  name                = var.resource_name
  location            = azurerm_resource_group.named_test_resource.location
  resource_group_name = azurerm_resource_group.named_test_resource.name

  sku {
    name     = "Standard_D13_v2"
    capacity = 2
  }

  tags = {
    Name = var.resource_name
  }
}

output "resource_aka" {
  value = "azure://${azurerm_kusto_cluster.named_test_resource.id}"
}

output "resource_aka_lower" {
  value = "azure://${lower(azurerm_kusto_cluster.named_test_resource.id)}"
}

output "resource_name" {
  value = var.resource_name
}

output "resource_id" {
  value = azurerm_kusto_cluster.named_test_resource.id
}

output "cluster_uri" {
  value = azurerm_kusto_cluster.named_test_resource.uri
}

output "location" {
  value = lower(azurerm_kusto_cluster.named_test_resource.location)
}

output "subscription_id" {
  value = var.azure_subscription
}