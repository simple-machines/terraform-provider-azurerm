
provider "azurerm" {}

resource "azurerm_resource_group" "stuff" {
    name = "testingthings"
    location = "Central US"

}

resource "azurerm_resource_group" "morestuff" {
    name = "testingthingsmore"
    location = "Central US"

    tags {
        hello = "${azurerm_resource_group.stuff.name}"
    }
}

resource "azurerm_eventhub_namespace" "inputs" {
    name = "inputs"
    resource_group_name = "${azurerm_resource_group.stuff.name}"
    location = "${azurerm_resource_group.stuff.location}"

    sku = "Standard"
}

resource "azurerm_eventhub" "input" {
    name = "input"

    resource_group_name = "${azurerm_resource_group.stuff.name}"
    namespace_name = "${azurerm_eventhub_namespace.inputs.name}"

    partition_count = 8
    message_retention = 1
}

resource "azurerm_eventhub" "output" {
    name = "output"

    resource_group_name = "${azurerm_resource_group.stuff.name}"
    namespace_name = "${azurerm_eventhub_namespace.inputs.name}"

    partition_count = 8
    message_retention = 1
}

resource "azurerm_eventhub_consumer_group" "query" {
    name = "Hello"

    resource_group_name = "${azurerm_resource_group.stuff.name}"

    namespace_name = "${azurerm_eventhub_namespace.inputs.name}"
    eventhub_name = "${azurerm_eventhub.input.name}"
}

resource "azurerm_eventhub_authorization_rule" "readinput" {
    name = "inputsreadinput"

    resource_group_name = "${azurerm_resource_group.stuff.name}"
    namespace_name = "${azurerm_eventhub_namespace.inputs.name}"
    eventhub_name = "${azurerm_eventhub.input.name}"

    listen = true
}

resource "azurerm_eventhub_authorization_rule" "writeoutput" {
    name = "inputswriteoutput"

    resource_group_name = "${azurerm_resource_group.stuff.name}"
    namespace_name = "${azurerm_eventhub_namespace.inputs.name}"
    eventhub_name = "${azurerm_eventhub.output.name}"

    send = true
}

resource "azurerm_storage_account" "store" {
    name = "terratestbythomas"

    resource_group_name = "${azurerm_resource_group.stuff.name}"
    location = "${azurerm_resource_group.stuff.location}"

    account_kind = "BlobStorage"
    account_tier = "Standard"
    account_replication_type = "LRS"

    enable_https_traffic_only = true
}

resource "azurerm_storage_container" "data" {
    name = "blobbyblobblob"
    resource_group_name = "${azurerm_resource_group.stuff.name}"
    storage_account_name = "${azurerm_storage_account.store.name}"
}

resource "azurerm_streamanalytics_job" "hello3" {
    name = "goodbye2"
    resource_group_name = "${azurerm_resource_group.stuff.name}"
    location = "${azurerm_resource_group.stuff.location}"

    sku = "Standard"

    data_locale = "en-AU"

    out_of_order_policy = "adjust"
    out_of_order_max_delay = 10
    late_arrival_max_delay = 5

    output_start_mode = "JobStartTime"

    output_start_time = "2017-12-14T23:34:19Z"


    input {
        name = "MyEventHubInput"
        type = "Stream"
        serialization = "JSON"
        encoding = "UTF8"

        datasource = "Microsoft.ServiceBus/EventHub"
        service_bus_namespace = "${azurerm_eventhub_namespace.inputs.name}"
        eventhub_name = "${azurerm_eventhub.input.name}"
        shared_access_policy_name = "${azurerm_eventhub_authorization_rule.readinput.name}"
        shared_access_policy_key = "${azurerm_eventhub_authorization_rule.readinput.primary_key}"
        consumer_group_name = "${azurerm_eventhub_consumer_group.query.name}"
    }

    input {
        name = "Uploads"
        type = "Stream"
        serialization = "JSON"
        encoding = "UTF8"

        datasource = "Microsoft.Storage/Blob"
        storage_account_name = "${azurerm_storage_account.store.name}"
        storage_account_key = "${azurerm_storage_account.store.primary_access_key}"
        storage_container = "${azurerm_storage_container.data.name}"
        storage_path_pattern = "/uploads/{date}"
        storage_date_format = "yyyy/MM/dd"
        storage_time_format = "HH"
        storage_source_partition_count = 32
    }

    transformation {
        name = "ProcessRubbish"
        streaming_units = 1
        query = "select id, name, address, temperature, weight into [MyEventHubOutput] from [MyEventHubInput]"
    }

    output {
        name = "MyEventHubOutput"
        serialization = "JSON"
        encoding = "UTF8"
        json_format = "Array"

        datasource = "Microsoft.ServiceBus/EventHub"
        service_bus_namespace = "${azurerm_eventhub_namespace.inputs.name}"
        eventhub_name = "${azurerm_eventhub.output.name}"
        shared_access_policy_name = "${azurerm_eventhub_authorization_rule.writeoutput.name}"
        shared_access_policy_key = "${azurerm_eventhub_authorization_rule.writeoutput.primary_key}"
        eventhub_partition_key = "id"
    }
}
