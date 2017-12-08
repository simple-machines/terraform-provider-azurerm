package azurerm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/streamanalytics"
	"github.com/hashicorp/terraform/helper/schema"
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
