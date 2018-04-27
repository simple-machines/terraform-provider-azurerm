package azurerm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/streamanalytics/mgmt/2016-03-01/streamanalytics"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
)

type validationcase struct {
	Value interface{}
	Valid bool
}

func testValidationFunction(t *testing.T, f schema.SchemaValidateFunc, cases []validationcase) {
	for k, v := range cases {
		name := fmt.Sprintf("item-%d", k+1)
		_, es := f(v.Value, name)

		if v.Valid {
			if len(es) > 0 {
				t.Fatalf("Validation rejected %s, expected it to accept", name)
			}
		} else {
			if len(es) == 0 {
				t.Fatalf("Validation accepted %s, expected it to reject", name)
			}
		}
	}
}

func TestValidation_sku(t *testing.T) {
	testValidationFunction(t, validateSku, []validationcase{
		{"example", false},
		{"example", false},
		{"standard", false},
		{"Standard", true},
	})
}

func TestValidation_outputStartMode(t *testing.T) {
	testValidationFunction(t, validateOutputStartMode, []validationcase{
		{"", false},
		{"example", false},
		{"jobstarttime", false},
		{"JobStartTime", true},
		{"customtime", false},
		{"CustomTime", true},
		{"lastoutputeventtime", false},
		{"LastOutputEventTime", true},
	})
}

func TestValidation_encoding(t *testing.T) {
	testValidationFunction(t, validateEncoding, []validationcase{
		{"", false},
		{"UTF-8", false},
		{"utf-8", false},
		{"ASCII", false},
		{"UTF8", true},
	})
}

func TestValidation_serializationType(t *testing.T) {
	testValidationFunction(t, validateSerializationType, []validationcase{
		{"avro", false},
		{"AVRO", false},
		{"Avro", true},
	})
}

func TestValidation_Output_Serialization(t *testing.T) {
	testValidationFunction(t, validateOutputSerialization, []validationcase{
		{map[string]interface{}{}, false},
		{map[string]interface{}{
			"serialization": "Avro",
		}, true},
		{map[string]interface{}{
			"serialization": "Json",
			"encoding":      "UTF8",
			"json_format":   "LineSeparated",
		}, true},
		{map[string]interface{}{
			"serialization": "Json",
			"encoding":      "UTF8",
			"json_format":   "Array",
		}, true},
		{map[string]interface{}{
			"serialization":   "Csv",
			"field_delimiter": "|",
			"encoding":        "UTF8",
		}, true}, {map[string]interface{}{
			"serialization":   "Csv",
			"field_delimiter": ",",
			"encoding":        "UTF8",
		}, true},
	})
}

func TestValidation_outputDataSource(t *testing.T) {
	testValidationFunction(t, validateOutputDatasource, []validationcase{
		{map[string]interface{}{}, false},
		{map[string]interface{}{
			"datasource": "I am not a number",
		}, false},
		{map[string]interface{}{
			"datasource": string(streamanalytics.TypeMicrosoftSQLServerDatabase),
		}, false},
		{map[string]interface{}{
			"datasource":        string(streamanalytics.TypeMicrosoftSQLServerDatabase),
			"database_server":   "server.example.com",
			"database_name":     "example",
			"database_table":    "sheep",
			"database_user":     "root",
			"database_password": "hunter2",
		}, true},
	})
}

// TODO: Implement a round-trip serialization property.
func TestJSONMarshalling(t *testing.T) {

	itsAnOutput := "Microsoft.StreamAnalytics/streamingjobs/outputs"

	outputName := "output"
	server := "someServer"
	database := "someDatabase"
	table := "sampleTable"
	user := "someUser"
	password := "****************"

	sqlDatabaseOutputDS := streamanalytics.AzureSQLDatabaseOutputDataSource{
		Type: streamanalytics.TypeMicrosoftSQLServerDatabase,
		AzureSQLDatabaseOutputDataSourceProperties: &streamanalytics.AzureSQLDatabaseOutputDataSourceProperties{
			Server:   &server,
			Database: &database,
			User:     &user,
			Password: &password,
			Table:    &table,
		},
	}

	outputProps := streamanalytics.OutputProperties{
		Datasource: &sqlDatabaseOutputDS,
	}

	anOutput := streamanalytics.Output{
		Name:             &outputName,
		Type:             &itsAnOutput,
		OutputProperties: &outputProps,
	}

	dsJSON, _ := json.Marshal(sqlDatabaseOutputDS)
	var parsedDS streamanalytics.AzureSQLDatabaseOutputDataSource
	json.Unmarshal(dsJSON, &parsedDS)
	dsJSON2, _ := json.Marshal(parsedDS)

	if !bytes.Equal(dsJSON, dsJSON2) {
		t.Logf("Initial Datasource JSON: %s", dsJSON)
		t.Logf("Second Datasource JSON:  %s", dsJSON2)
		t.Fatalf("marshal(OutputDataSource) == marshal-parse-marshal(OutputDataSource) failed")
	}

	opJSON, _ := json.Marshal(outputProps)
	var parsedOP streamanalytics.OutputProperties
	json.Unmarshal(opJSON, &parsedOP)
	opJSON2, _ := json.Marshal(parsedOP)

	if !bytes.Equal(opJSON, opJSON2) {
		t.Logf("Initial output properties JSON: %s", opJSON)
		t.Logf("Second output properties JSON:  %s", opJSON2)
		t.Fatalf("marshal(OutputProperties) == marshal-parse-marshal(OutputProperties) failed")
	}

	outputJSON, _ := json.Marshal(anOutput)
	var parsedOutput streamanalytics.Output
	json.Unmarshal(outputJSON, &parsedOutput)
	outputJSON2, _ := json.Marshal(parsedOutput)

	if !bytes.Equal(outputJSON2, outputJSON) {
		t.Logf("Initial Output JSON: %s", outputJSON)
		t.Logf("Second Output JSON:  %s", outputJSON2)
		t.Fatalf("marshal(Output) == marshal-parse-marshal(Output) failed")
	}

}

func TestAccAzureRMStreamAnalyticsJob_streamInput(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurerm_streamanalytics_job.test"
	config := testAccAzureRMStreamAnalytics_streamInput(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsJobExists("azurerm_streamanalytics_job.test"),
					resource.TestCheckResourceAttr(resourceName, "input.0.name", fmt.Sprintf("acctesteventhubinput-%d", ri)),
				),
			},
		},
	})
}

func TestAccAzureRMStreamAnalyticsJob_referenceInput(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurerm_streamanalytics_job.test"
	config := testAccAzureRMStreamAnalytics_referenceInput(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsJobExists("azurerm_streamanalytics_job.test"),
					resource.TestCheckResourceAttr(resourceName, "input.0.name", fmt.Sprintf("acctestblobinput-%d", ri)),
				),
			},
		},
	})
}

func TestAccAzureRMStreamAnalyticsJob_sqlDbOut(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurerm_streamanalytics_job.test"
	config := testAccAzureRMStreamAnalytics_sqlDbOutput(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsJobExists("azurerm_streamanalytics_job.test"),
					resource.TestCheckResourceAttr(resourceName, "output.0.name", fmt.Sprintf("acctestdboutput-%d", ri)),
				),
			},
		},
	})
}

func TestAccAzureRMStreamAnalyticsJob_eventHubOut(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurerm_streamanalytics_job.test"
	config := testAccAzureRMStreamAnalytics_eventHubOutput(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsJobExists("azurerm_streamanalytics_job.test"),
					resource.TestCheckResourceAttr(resourceName, "output.0.name", fmt.Sprintf("acctesteventhuboutput-%d", ri)),
				),
			},
		},
	})
}

func TestAccAzureRMStreamAnalyticsJob_blobOut(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurerm_streamanalytics_job.test"
	config := testAccAzureRMStreamAnalytics_blobOutput(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsJobExists("azurerm_streamanalytics_job.test"),
					resource.TestCheckResourceAttr(resourceName, "output.0.name", fmt.Sprintf("acctestbloboutput-%d", ri)),
				),
			},
		},
	})
}

func TestAccAzureRMStreamAnalyticsJob_transformation(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurerm_streamanalytics_job.test"
	config := testAccAzureRMStreamAnalytics_transformation(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsJobExists("azurerm_streamanalytics_job.test"),
					resource.TestCheckResourceAttr(resourceName, "transformation.0.name", fmt.Sprintf("accteststreamtransformation-%d", ri)),
				),
			},
		},
	})
}

//TODO datalake output

func testCheckAzureRMStreamAnalyticsJobDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).streamAnalyticsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_streamanalytics_job" {
			continue
		}

		jobName := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(ctx, resourceGroup, jobName, thingsToGet)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Stream analytics job %q (resource group %q) still exists: %+v", jobName, resourceGroup, resp)
		}
	}

	return nil
}

func testCheckAzureRMStreamAnalyticsJobExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		jobName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for Stream Analytics Job: %s", jobName)
		}

		conn := testAccProvider.Meta().(*ArmClient).streamAnalyticsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := conn.Get(ctx, resourceGroup, jobName, thingsToGet)

		if err != nil {
			return fmt.Errorf("Bad: Get on streamAnalyticsClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Stream analytics job %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testAccAzureRMStreamAnalytics_streamInput(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_eventhub_namespace" "test" {
  name                = "acctesteventhubnamespace-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"
  capacity            = 1
}

resource "azurerm_eventhub" "test" {
  name                = "acctesteventhub-%d"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  partition_count     = 2
  message_retention   = 1
}

resource "azurerm_eventhub_authorization_rule" "test" {
  name                = "acctesteventhubauth-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  eventhub_name       = "${azurerm_eventhub.test.name}"
  listen              = true
}

resource "azurerm_eventhub_consumer_group" "test" {
  name                = "acctesteventhubconsumer-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  eventhub_name       = "${azurerm_eventhub.test.name}"
}

resource "azurerm_streamanalytics_job" "test" {
  name                = "accteststreamanalyticsjob-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"

  input {
    name          = "acctesteventhubinput-%d"
    type          = "Stream"
    serialization = "Avro"
    datasource    = "Microsoft.ServiceBus/EventHub"

    service_bus_namespace     = "${azurerm_eventhub_namespace.test.name}"
    eventhub_name             = "${azurerm_eventhub.test.name}"
    shared_access_policy_name = "${azurerm_eventhub_authorization_rule.test.name}"
    shared_access_policy_key  = "${azurerm_eventhub_authorization_rule.test.primary_key}"
    consumer_group_name       = "${azurerm_eventhub_consumer_group.test.name}"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureRMStreamAnalytics_referenceInput(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "ats%d"
  resource_group_name      = "${azurerm_resource_group.test.name}"
  location                 = "${azurerm_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
}

resource "azurerm_storage_container" "test" {
  name                  = "accteststoragecontainer-%d"
  resource_group_name   = "${azurerm_resource_group.test.name}"
  storage_account_name  = "${azurerm_storage_account.test.name}"
  container_access_type = "container"
}

resource "azurerm_storage_blob" "test" {
  name                   = "acctestblob-%d"
  resource_group_name    = "${azurerm_resource_group.test.name}"
  storage_account_name   = "${azurerm_storage_account.test.name}"
  storage_container_name = "${azurerm_storage_container.test.name}"
  type                   = "block"
}

resource "azurerm_streamanalytics_job" "test" {
  name                = "accteststreamanalyticsjob-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"

  input {
    name          = "acctestblobinput-%d"
    type          = "Reference"
    serialization = "Json"
    encoding      = "UTF8"
    datasource    = "Microsoft.Storage/Blob"

    storage_account_name = "${azurerm_storage_account.test.name}"
    storage_account_key  = "${azurerm_storage_account.test.primary_access_key}"
    storage_container    = "${azurerm_storage_container.test.name}"
    storage_path_pattern = "/test/{date}/{time}/test.json"
    storage_date_format  = "yyyy-MM-dd"
    storage_time_format  = "HH"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureRMStreamAnalytics_sqlDbOutput(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_sql_server" "test" {
  name                         = "acctestsqlserver-%d"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  location                     = "${azurerm_resource_group.test.location}"
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_sql_database" "test" {
  name                = "acctestsqldb-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  server_name         = "${azurerm_sql_server.test.name}"
}

resource "azurerm_streamanalytics_job" "test" {
  name                = "accteststreamanalyticsjob-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"

  output {
    name              = "acctestdboutput-%d"
    serialization     = "Avro"
    datasource        = "Microsoft.Sql/Server/Database"
    database_server   = "${azurerm_sql_server.test.name}"
    database_name     = "${azurerm_sql_database.test.name}"
    database_table    = "something"
    database_user     = "mradministrator"
    database_password = "thisIsDog11"
  }
 }
`, rInt, location, rInt, rInt, rInt, rInt)
}

func testAccAzureRMStreamAnalytics_eventHubOutput(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_eventhub_namespace" "test" {
  name                = "acctesteventhubnamespace-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"
  capacity            = 1
}

resource "azurerm_eventhub" "test" {
  name                = "acctesteventhub-%d"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  partition_count     = 2
  message_retention   = 1
}

resource "azurerm_eventhub_authorization_rule" "test" {
  name                = "acctesteventhubauth-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  eventhub_name       = "${azurerm_eventhub.test.name}"
  send                = true
}

resource "azurerm_eventhub_consumer_group" "test" {
  name                = "acctesteventhubconsumer-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  eventhub_name       = "${azurerm_eventhub.test.name}"
}

resource "azurerm_streamanalytics_job" "test" {
  name                = "accteststreamanalyticsjob-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"

  output {
    name                      = "acctesteventhuboutput-%d"
    serialization             = "Avro"
    datasource                = "Microsoft.ServiceBus/EventHub"
    service_bus_namespace     = "${azurerm_eventhub_namespace.test.name}"
    eventhub_name             = "${azurerm_eventhub.test.name}"
    shared_access_policy_name = "${azurerm_eventhub_authorization_rule.test.name}"
    shared_access_policy_key  = "${azurerm_eventhub_authorization_rule.test.primary_key}"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureRMStreamAnalytics_blobOutput(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "ats%d"
  resource_group_name      = "${azurerm_resource_group.test.name}"
  location                 = "${azurerm_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
}

resource "azurerm_storage_container" "test" {
  name                  = "accteststoragecontainer-%d"
  resource_group_name   = "${azurerm_resource_group.test.name}"
  storage_account_name  = "${azurerm_storage_account.test.name}"
  container_access_type = "container"
}

resource "azurerm_storage_blob" "test" {
  name                   = "acctestblob-%d"
  resource_group_name    = "${azurerm_resource_group.test.name}"
  storage_account_name   = "${azurerm_storage_account.test.name}"
  storage_container_name = "${azurerm_storage_container.test.name}"
  type                   = "block"
}

resource "azurerm_streamanalytics_job" "test" {
  name                = "accteststreamanalyticsjob-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"

  output {
    name          = "acctestbloboutput-%d"
    serialization = "Json"
    encoding      = "UTF8"
    datasource    = "Microsoft.Storage/Blob"

    storage_account_name = "${azurerm_storage_account.test.name}"
    storage_account_key  = "${azurerm_storage_account.test.primary_access_key}"
    storage_container    = "${azurerm_storage_container.test.name}"
    storage_path_pattern = "/test/{date}/{time}/test.json"
    storage_time_format  = "HH"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureRMStreamAnalytics_transformation(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_eventhub_namespace" "test" {
  name                = "acctesteventhubnamespace-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"
  capacity            = 1
}

resource "azurerm_eventhub" "test" {
  name                = "acctesteventhub-%d"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  partition_count     = 2
  message_retention   = 1
}

# Create authorization rule for listening on the event hub
resource "azurerm_eventhub_authorization_rule" "test" {
  name                = "acctesteventhubauth-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  eventhub_name       = "${azurerm_eventhub.test.name}"
  listen              = true
}

# Create consumer group on the event hub
resource "azurerm_eventhub_consumer_group" "test" {
  name                = "acctesteventhubconsumer-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  namespace_name      = "${azurerm_eventhub_namespace.test.name}"
  eventhub_name       = "${azurerm_eventhub.test.name}"
}

resource "azurerm_sql_server" "test" {
  name                         = "acctestsqlserver-%d"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  location                     = "${azurerm_resource_group.test.location}"
  version                      = "12.0"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsDog11"
}

resource "azurerm_sql_database" "test" {
  name                = "acctestsqldb-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  server_name         = "${azurerm_sql_server.test.name}"
}

resource "azurerm_streamanalytics_job" "test" {
  name                = "accteststreamanalyticsjob-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"

  input {
    name                      = "accteststreaminput-%d"
    type                      = "Stream"
    serialization             = "Avro"
    datasource                = "Microsoft.ServiceBus/EventHub"
    service_bus_namespace     = "${azurerm_eventhub_namespace.test.name}"
    eventhub_name             = "${azurerm_eventhub.test.name}"
    shared_access_policy_name = "${azurerm_eventhub_authorization_rule.test.name}"
    shared_access_policy_key  = "${azurerm_eventhub_authorization_rule.test.primary_key}"
    consumer_group_name       = "${azurerm_eventhub_consumer_group.test.name}"
  }

  transformation {
    name  = "accteststreamtransformation-%d"
    query = "select id into [accteststreamoutput-%d] from [accteststreaminput-%d]"
  }

  output {
    name              = "accteststreamoutput-%d"
    datasource        = "Microsoft.Sql/Server/Database"
    database_server   = "${azurerm_sql_server.test.name}"
    database_name     = "${azurerm_sql_database.test.name}"
    database_table    = "something"
    database_user     = "mradministrator"
    database_password = "thisIsDog11"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt, rInt, rInt, rInt, rInt, rInt, rInt)
}
