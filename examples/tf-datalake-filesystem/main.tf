provider "azurerm" {}

resource "azurerm_resource_group" "rg" {
  name = "thomas"
  location = "East US 2"

  tags {
    test = "tf-datalake-filesystem"
  }
}

variable "microsoft_eventhubs" {
  default = "2f121587-2394-44ea-bd68-c4c4e3bfa391"
}

resource "azurerm_datalake_store_account" "analytics" {
  name = "billyjanesdatawarehouse"
  
  tier = "Consumption"

  resource_group_name = "${azurerm_resource_group.rg.name}"
  location = "${azurerm_resource_group.rg.location}"

  tags {
    test = "tf-datalake-filesystem"
  }
}

resource "azurerm_datalake_filesystem" "lake" {
  account_name = "${azurerm_datalake_store_account.analytics.name}"

  path = "/"
  mode = 770
  mask = "rwx"

  access_acl = [
    "user:${var.microsoft_eventhubs}:--x"
  ]

  default_acl = [
    "user:${var.microsoft_eventhubs}:--x"
  ]
}

resource "azurerm_datalake_filesystem" "archives" {
  account_name = "${azurerm_datalake_store_account.analytics.name}"

  path = "${azurerm_datalake_filesystem.lake.path}gator"

  mode = 700
  mask = "---"

  access_acl = [
    "user:${var.microsoft_eventhubs}:rwx",
  ]

  default_acl = [
    # "user:${var.microsoft_eventhubs}:rwx",
  ]
}
