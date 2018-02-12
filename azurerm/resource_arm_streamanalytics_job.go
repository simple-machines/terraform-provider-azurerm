package azurerm

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/streamanalytics/mgmt/2016-03-01/streamanalytics"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

const thingsToGet = "inputs,transformation,outputs"

func resourceArmStreamAnalyticsJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmStreamAnalyticsJobCreate,
		Read:   resourceArmStreamAnalyticsJobRead,
		Update: resourceArmStreamAnalyticsJobCreate,
		Delete: resourceArmStreamAnalyticsJobDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			//
			// Mandatory and built-in fields
			//

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			"tags": tagsSchema(),

			"sku": {
				Description:  "(Standard)",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSku,
			},

			//
			// Status fields
			//

			"job_id": {
				Description: "GUID uniquely identifying a Stream Analytics Job.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provisioning_state": {
				Type:        schema.TypeString,
				Description: "(Succeeded, Failed, Canceled)",
				Computed:    true,
			},
			"created_date": {
				Type:        schema.TypeString,
				Description: "ISO-8601 UTC timestamp when the job was created.",
				Computed:    true,
			},
			"job_state": {
				Description: "(Created)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"deployed_version": {
				Description: "ETag of the currently deployed job.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			//
			// Optional fields
			//

			"data_locale": {
				Type:        schema.TypeString,
				Description: "Name of a locale: [en-US], en-AU, fr-FR, ...",
				Default:     "en-US",
				Optional:    true,
				// TODO Validate
			},

			"late_arrival_max_delay": {
				Description:  "Maximum time to wait for a late arrival in seconds. -1 means wait forever.",
				Type:         schema.TypeInt,
				Default:      5,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1, 1814399),
			},

			"out_of_order_max_delay": {
				Type:        schema.TypeInt,
				Description: "seconds",
				Optional:    true,
				// TODO Validate
			},

			"out_of_order_policy": {
				// TODO: Document the default behaviour.
				Type:             schema.TypeString,
				Description:      "(Adjust, Drop)",
				Optional:         true,
				ValidateFunc:     validateOutOfOrderPolicy,
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},

			"output_start_mode": {
				// TODO: This is an/the way to start a job automatically. We should
				// document this.
				Type:             schema.TypeString,
				Description:      "(CustomTime, [JobStartTime], LastOutputEventTime)",
				Optional:         true,
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
				ValidateFunc:     validateOutputStartMode,
			},

			"output_start_time": {
				// TODO: required when output_start_mode==CustomTime, must be absent when output_start_mode==JobStartTime
				Type:         schema.TypeString,
				Description:  "ISO-8601 timestamp to start.",
				Optional:     true,
				ValidateFunc: validateRFC3339Date,
			},

			"last_output_event_time": {
				Type:        schema.TypeString,
				Description: "Timestamp of last output event.",
				Computed:    true,
			},

			"output_error_policy": {
				Type:             schema.TypeString,
				Description:      "(Stop, Drop)",
				Optional:         true,
				Default:          "Stop",
				ValidateFunc:     validateOutputErrorPolicy,
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},

			"input": {
				Type:        schema.TypeList,
				Description: "Datasources to be used by queries.",
				Optional:    true,
				// TODO: When Terraform supports validating lists and sets.
				// ValidateFunc: validateInput,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateFunc:     validateInputType,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
						"serialization": {
							Type:             schema.TypeString,
							Description:      "(AVRO|CSV|JSON)",
							Required:         true,
							ValidateFunc:     validateSerializationType,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
						"encoding": {
							Type:             schema.TypeString,
							Description:      "(UTF8)",
							Optional:         true,
							ValidateFunc:     validateEncoding,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
						"delimiter": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"json_format": {
							Type:             schema.TypeString,
							Description:      "(Array,LineSeparated)",
							Optional:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
						"datasource": {
							Type:         schema.TypeString,
							Description:  "Azure datasource type.",
							Required:     true,
							ValidateFunc: validateArmStreamAnalyticsInputDatasource,
						},
						// Storage account fields (for both stream and reference inputs)
						"storage_account_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_account_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"storage_container": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_path_pattern": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_date_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_time_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_source_partition_count": {
							// For Stream
							Type:     schema.TypeInt,
							Optional: true,
						},

						// Shared Event and IoT Hub fields
						"shared_access_policy_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"shared_access_policy_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"consumer_group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// Event Hub fields
						"service_bus_namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"eventhub_name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// IoT Hub fields
						"iot_hub_namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"iot_hub_endpoint": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			// To avoid having to split the single transformation when we
			// read.
			"transformation": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems:    1,
				Description: "SQL query to execute.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
						"streaming_units": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},

			"output": {
				Type:     schema.TypeList,
				Optional: true,
				Description: "Data outputs to be updated by queries.",
				// TODO: When Terraform supports validating lists and sets.
				// ValidateFunc: validateOutput,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							// TODO: Validate
						},
						"serialization": {
							Type:             schema.TypeString,
							Description:      "(AVRO|CSV|JSON)",
							Required:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
							ValidateFunc:     validation.StringInSlice([]string{"Avro", "CSV", "JSON"}, true),
						},
						"encoding": {
							// When serialization==CSV or serialization==JSON
							Type:             schema.TypeString,
							Description:      "Required when using CSV or JSON serialization (UTF8)",
							Optional:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
							ValidateFunc:     validation.StringInSlice([]string{"UTF8"}, true),
						},
						"delimiter": {
							// When serialization==CSV.
							Type:     schema.TypeString,
							Optional: true,
						},
						"json_format": {
							Type:             schema.TypeString,
							Description:      "Required when using JSON serialization (Array, LineSeparated)",
							Optional:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
							ValidateFunc:     validation.StringInSlice([]string{"Array", "LineSeparated"}, true),
						},
						"datasource": {
							Type:         schema.TypeString,
							Description:  "Azure datasource type",
							Required:     true,
							ValidateFunc: validateArmStreamAnalyticsOutputDatasource,
						},
						// Fields to support Blob storage
						// Storage account fields (for both stream and reference inputs)
						"storage_account_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_account_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"storage_container": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_path_pattern": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_date_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_time_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						// Fields to support SQL Server
						"database_server": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"database_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"database_table": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"database_user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"database_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						// Fields to support EventHub
						"shared_access_policy_name": {
							Type:        schema.TypeString,
							Description: "Required when using Event Hub output.",
							Optional:    true,
						},
						"shared_access_policy_key": {
							Type:        schema.TypeString,
							Description: "Required when using Event Hub output.",
							Optional:    true,
							Sensitive:   true,
						},
						"service_bus_namespace": {
							Type:        schema.TypeString,
							Description: "Required when using Event Hub output.",
							Optional:    true,
						},
						"eventhub_name": {
							Type:        schema.TypeString,
							Description: "Required when using Event Hub output.",
							Optional:    true,
						},
						"eventhub_partition_key": {
							Type:        schema.TypeString,
							Description: "Optional when using Event Hub output.",
							Optional:    true,
						},
						// Fields to support Data Lake Store
						"data_lake_refresh_token": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"data_lake_token_user_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_token_display_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_account_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_tenant_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_path_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_date_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_time_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceArmStreamAnalyticsJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).streamAnalyticsClient
	ctx := meta.(*ArmClient).StopContext

	resourceGroupName := d.Get("resource_group_name").(string)
	jobName := d.Get("name").(string)
	location := d.Get("location").(string)
	tags := d.Get("tags").(map[string]interface{})

	sku := d.Get("sku").(string)

	job := streamanalytics.StreamingJob{
		Name:     &jobName,
		Location: &location,
		Tags:     expandTags(tags),
		StreamingJobProperties: &streamanalytics.StreamingJobProperties{
			Sku: &streamanalytics.Sku{Name: streamanalytics.SkuName(sku)},
		},
	}

	// Should we specify this?
	if v, ok := d.GetOk("deployed_version"); ok {
		etag := v.(string)
		job.StreamingJobProperties.Etag = &etag
	}

	if v, ok := d.GetOk("data_locale"); ok {
		locale := v.(string)
		job.StreamingJobProperties.DataLocale = &locale
	}

	if v, ok := d.GetOk("late_arrival_max_delay"); ok {
		delay := int32(v.(int))
		job.StreamingJobProperties.EventsLateArrivalMaxDelayInSeconds = &delay
	}

	if v, ok := d.GetOk("out_of_order_max_delay"); ok {
		delay := int32(v.(int))
		job.StreamingJobProperties.EventsOutOfOrderMaxDelayInSeconds = &delay
	}

	if v, ok := d.GetOk("out_of_order_policy"); ok {
		policy := v.(string)
		job.StreamingJobProperties.EventsOutOfOrderPolicy = streamanalytics.EventsOutOfOrderPolicy(policy)
	}

	if v, ok := d.GetOk("output_start_mode"); ok {
		mode := v.(string)
		job.StreamingJobProperties.OutputStartMode = streamanalytics.OutputStartMode(mode)

		switch strings.ToLower(mode) {
		case "customtime":
			if v, ok := d.GetOk("output_start_time"); ok {
				dateString := v.(string)
				time, err := date.ParseTime(time.RFC3339, dateString)
				if err != nil {
					return fmt.Errorf("Invalid date for output_start_time: %q", err)
				}
				job.StreamingJobProperties.OutputStartTime = &date.Time{Time: time}
			}
		case "jobstarttime":
		case "lastoutputeventtime":
		default:
			return fmt.Errorf("Unknown value %s for output_start_mode", mode)
		}
	}

	if v, ok := d.GetOk("output_error_policy"); ok {
		policy := v.(string)
		job.StreamingJobProperties.OutputErrorPolicy = streamanalytics.OutputErrorPolicy(policy)
	}

	if v := d.Get("input"); v != nil {
		inputs := v.([]interface{})
		log.Printf("[INFO] Processing %d input configurations", len(inputs))
		inputConfigs := make([]streamanalytics.Input, 0, len(inputs))
		for _, configRaw := range inputs {
			config := configRaw.(map[string]interface{})
			input, err := resourceArmStreamAnalyticsParseInput(config)
			if err != nil {
				return err
			}
			inputConfigs = append(inputConfigs, *input)
		}
		job.StreamingJobProperties.Inputs = &inputConfigs
	}

	if v := d.Get("transformation"); v != nil {
		transformations := v.([]interface{})
		// NB: Validation ensures that there is one, and yet.
		if len(transformations) > 1 {
			return fmt.Errorf("Cannot process job with %d transformations", len(transformations))
		}
		for _, configRaw := range transformations {
			config := configRaw.(map[string]interface{})
			transform, err := resourceArmStreamAnalyticsParseTransformation(config)
			if err != nil {
				return err
			}
			job.StreamingJobProperties.Transformation = transform
		}
	}

	if v := d.Get("output"); v != nil {
		outputs := v.([]interface{})
		log.Printf("[INFO] Processing %d output configurations", len(outputs))
		outputConfigs := make([]streamanalytics.Output, 0, len(outputs))
		for _, configRaw := range outputs {
			config := configRaw.(map[string]interface{})
			output, err := resourceArmStreamAnalyticsParseOutput(config)
			if err != nil {
				return err
			}
			outputConfigs = append(outputConfigs, *output)
		}
		job.StreamingJobProperties.Outputs = &outputConfigs
	}

	// The structure of the async polling for long-running operations means
	// we can't assume that the response to this call is valid. In the
	// interest of simplicity we ignore the result here and immediately do
	// a get.
	ifMatch := ""     // TODO: Etag if resource to update.
	ifNoneMatch := "" // TODO: "*" to create but not update.
	_, err := client.CreateOrReplace(ctx, job, resourceGroupName, jobName, ifMatch, ifNoneMatch)
	if err != nil {
		return fmt.Errorf("Error issuing AzureRM create request for StreamAnalytics Job %q: %+v",
			jobName, err)
	}

	read, err := client.Get(ctx, resourceGroupName, jobName, thingsToGet)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Stream Analytics Job %s (resource group %s) ID", jobName, resourceGroupName)
	}

	d.SetId(*read.ID)

	return resourceArmStreamAnalyticsJobRead(d, meta)
}

func resourceArmStreamAnalyticsJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).streamAnalyticsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroupName := id.ResourceGroup
	jobName := id.Path["streamingjobs"]

	res, err := client.Get(ctx, resourceGroupName, jobName, thingsToGet)

	if err != nil {
		if utils.ResponseWasNotFound(res.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on AzureRM Stream Analytics Job '%s': %+v", jobName, err)
	}

	d.Set("name", jobName)
	d.Set("resource_group_name", resourceGroupName)
	d.Set("location", azureRMNormalizeLocation(*res.Location))
	flattenAndSetTags(d, res.Tags)

	properties := res.StreamingJobProperties
	d.Set("sku", properties.Sku.Name)
	d.Set("job_id", *properties.JobID)
	d.Set("provisioning_state", *properties.ProvisioningState)
	d.Set("job_state", *properties.JobState)
	d.Set("output_start_mode", string(properties.OutputStartMode))
	if v := properties.OutputStartTime; v != nil {
		d.Set("output_start_time", string(v.Format(time.RFC3339)))
	}
	if v := properties.LastOutputEventTime; v != nil {
		d.Set("last_output_event_time", string(v.Format(time.RFC3339)))
	}
	d.Set("out_of_order_policy", string(properties.EventsOutOfOrderPolicy))
	d.Set("output_error_policy", string(properties.OutputErrorPolicy))

	if v := properties.EventsOutOfOrderMaxDelayInSeconds; v != nil {
		d.Set("out_of_order_max_delay", int(*properties.EventsOutOfOrderMaxDelayInSeconds))
	}
	d.Set("late_arrival_max_delay", int(*properties.EventsLateArrivalMaxDelayInSeconds))
	d.Set("data_locale", *properties.DataLocale)
	d.Set("created_date", string(properties.CreatedDate.Format(time.RFC3339)))
	if v := properties.Etag; v != nil {
		d.Set("deployed_version", *v)
	}

	log.Printf("[INFO] Received %d inputs, %d outputs", len(*properties.Inputs), len(*properties.Outputs))

	if properties.Inputs != nil && len(*properties.Inputs) > 0 {
		inputConfigs := make([]map[string]interface{}, 0, len(*properties.Inputs))
		log.Printf("[INFO] Parsing %d inputs: %+v", len(*properties.Inputs), *properties.Inputs)

		for _, input := range *properties.Inputs {
			conf, err := flattenStreamAnalyticsJobInput(&input)
			if err != nil {
				return err
			}
			inputConfigs = append(inputConfigs, conf)
		}

		d.Set("input", inputConfigs)
	} else {
		log.Printf("[WARN] Received no inputs")
	}

	if properties.Transformation != nil {
		transformConfigs := make([]map[string]interface{}, 1)

		conf, err := flattenStreamAnalyticsJobTransformation(properties.Transformation)
		if err != nil {
			return err
		}
		transformConfigs[0] = conf

		d.Set("transformation", transformConfigs)
	} else {
		log.Printf("[WARN] Received no transformation")
	}

	if properties.Outputs != nil && len(*properties.Outputs) > 0 {
		outputConfigs := make([]map[string]interface{}, 0, len(*properties.Outputs))
		log.Printf("[INFO] Parsing %d outputs: %+v", len(*properties.Outputs), *properties.Outputs)

		for _, output := range *properties.Outputs {
			conf, err := flattenStreamAnalyticsJobOutput(&output)
			if err != nil {
				return err
			}
			outputConfigs = append(outputConfigs, *conf)
		}

		d.Set("output", outputConfigs)
	} else {
		log.Printf("[WARN] Received no outputs")
	}

	return nil
}

func resourceArmStreamAnalyticsJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).streamAnalyticsClient
	ctx := meta.(*ArmClient).StopContext

	resourceGroupName := d.Get("resource_group_name").(string)
	jobName := d.Get("name").(string)

	_, err := client.Delete(ctx, resourceGroupName, jobName)
	if err != nil {
		return fmt.Errorf("Error issuing AzureRM delete request for StreamAnalytics Job %q: %+v",
			jobName, err)
	}

	return nil
}

// Convert an `input{ ... }` section into an Input structure.
func resourceArmStreamAnalyticsParseInput(data map[string]interface{}) (*streamanalytics.Input, error) {
	inputName := data["name"].(string)
	inputType := data["type"].(string)

	input := streamanalytics.Input{
		Name: &inputName,
		// Type:
	}

	serialization, err := parseArmStreamAnalyticsSerialization(data)
	if err != nil {
		return nil, err
	}

	switch inputType {
	case string(streamanalytics.TypeReference):
		ds, err := parseArmStreamAnalyticsReferenceDatasource(data)
		if err != nil {
			return nil, err
		}
		input.Properties = &streamanalytics.ReferenceInputProperties{
			Serialization: serialization,
			Type:          streamanalytics.TypeReference,
			Datasource:    ds,
		}

	case string(streamanalytics.TypeStream):
		ds, err := parseArmStreamAnalyticsStreamDatasource(data)
		if err != nil {
			return nil, err
		}
		input.Properties = &streamanalytics.StreamInputProperties{
			Serialization: serialization,
			Type:          streamanalytics.TypeStream,
			Datasource:    ds,
		}
	}

	return &input, nil
}

// Convert a `transformation{ ... }` section in to a Transformation structure.
func resourceArmStreamAnalyticsParseTransformation(data map[string]interface{}) (*streamanalytics.Transformation, error) {
	name := data["name"].(string)
	streamingUnits := int32(data["streaming_units"].(int))
	query := data["query"].(string)

	result := streamanalytics.Transformation{
		Name: &name,
		TransformationProperties: &streamanalytics.TransformationProperties{
			StreamingUnits: &streamingUnits,
			Query:          &query,
		},
	}

	return &result, nil
}

// Convert an `input{ ... }` section into an Input structure.
func resourceArmStreamAnalyticsParseOutput(data map[string]interface{}) (*streamanalytics.Output, error) {
	name := data["name"].(string)

	serialization, err := parseArmStreamAnalyticsSerialization(data)
	if err != nil {
		return nil, err
	}

	ds, err := parseArmStreamAnalyticsOutputDatasource(data)
	if err != nil {
		return nil, err
	}

	return &streamanalytics.Output{
		Name: &name,
		OutputProperties: &streamanalytics.OutputProperties{
			Datasource:    ds,
			Serialization: serialization,
		},
	}, nil
}

func parseArmStreamAnalyticsOutputDatasource(data map[string]interface{}) (streamanalytics.BasicOutputDataSource, error) {
	datasource := data["datasource"].(string)

	extractAndValidateProp := extractAndValidateRequiredProperty(data)
	extractOptionalProp :=  extractOptionalProperty(data)

	switch datasource {
	case string(streamanalytics.TypeMicrosoftDataLakeAccounts):
		// TODO Validate inputs
		refreshToken := data["data_lake_refresh_token"].(string)
		userName := data["data_lake_token_user_name"].(string)
		displayName := data["data_lake_token_display_name"].(string)
		accountName := data["data_lake_account_name"].(string)
		tenantID := data["data_lake_tenant_id"].(string)
		pathPrefix := data["data_lake_path_prefix"].(string)
		dateFormat := data["data_lake_date_format"].(string)
		timeFormat := data["data_lake_time_format"].(string)

		result := streamanalytics.AzureDataLakeStoreOutputDataSource{
			Type: streamanalytics.TypeMicrosoftDataLakeAccounts,
			AzureDataLakeStoreOutputDataSourceProperties: &streamanalytics.AzureDataLakeStoreOutputDataSourceProperties{
				RefreshToken:           &refreshToken,
				TokenUserPrincipalName: &userName,
				TokenUserDisplayName:   &displayName,
				AccountName:            &accountName,
				TenantID:               &tenantID,
				FilePathPrefix:         &pathPrefix,
				DateFormat:             &dateFormat,
				TimeFormat:             &timeFormat,
			},
		}

		return &result, nil

	case string(streamanalytics.TypeMicrosoftServiceBusEventHub):
		namespace, err := extractAndValidateProp("service_bus_namespace")
		if err != nil {
			return nil, err
		}
		eventHub, err := extractAndValidateProp("eventhub_name")
		if err != nil {
			return nil, err
		}
		policyName, err := extractAndValidateProp("shared_access_policy_name")
		if err != nil {
			return nil, err
		}
		policyKey, err := extractAndValidateProp("shared_access_policy_key")
		if err != nil {
			return nil, err
		}
		partitionKey := extractOptionalProp("eventhub_partition_key")

		result := streamanalytics.EventHubOutputDataSource{
			Type: streamanalytics.TypeMicrosoftServiceBusEventHub,
			EventHubOutputDataSourceProperties: &streamanalytics.EventHubOutputDataSourceProperties{
				ServiceBusNamespace:    namespace,
				SharedAccessPolicyName: policyName,
				SharedAccessPolicyKey:  policyKey,
				EventHubName:           eventHub,
				PartitionKey:		partitionKey,
			},
		}

		return &result, nil

	case string(streamanalytics.TypeMicrosoftSQLServerDatabase):
		databaseServer, err := extractAndValidateProp("database_server")
		if err != nil {
			return nil, err
		}
		databaseName, err := extractAndValidateProp("database_name")
		if err != nil {
			return nil, err
		}
		databaseTable, err := extractAndValidateProp("database_table")
		if err != nil {
			return nil, err
		}
		databaseUser, err := extractAndValidateProp("database_user")
		if err != nil {
			return nil, err
		}
		databasePassword, err := extractAndValidateProp("database_password")
		if err != nil {
			return nil, err
		}

		result := streamanalytics.AzureSQLDatabaseOutputDataSource{
			Type: streamanalytics.TypeMicrosoftSQLServerDatabase,
			AzureSQLDatabaseOutputDataSourceProperties: &streamanalytics.AzureSQLDatabaseOutputDataSourceProperties{
				Server:   databaseServer,
				Database: databaseName,
				User:     databaseUser,
				Password: databasePassword,
				Table:    databaseTable,
			},
		}

		return &result, nil

	case string(streamanalytics.TypeMicrosoftStorageBlob):
		accountName, err := extractAndValidateProp("storage_account_name")
		if err != nil {
			return nil, err
		}
		accountKey, err := extractAndValidateProp("storage_account_key")
		if err != nil {
			return nil, err
		}
		container, err := extractAndValidateProp("storage_container")
		if err != nil {
			return nil, err
		}
		pathPattern, err := extractAndValidateProp("storage_path_pattern")
		if err != nil {
			return nil, err
		}
		dateFormat := extractOptionalProp("storage_date_format")
		timeFormat := extractOptionalProp("storage_time_format")

		accounts := make([]streamanalytics.StorageAccount, 1)
		accounts[0].AccountName = accountName
		accounts[0].AccountKey = accountKey

		result := streamanalytics.BlobOutputDataSource{
			Type: streamanalytics.TypeMicrosoftStorageBlob,
			BlobOutputDataSourceProperties: &streamanalytics.BlobOutputDataSourceProperties{
				StorageAccounts: &accounts,
				Container:       container,
				PathPattern:     pathPattern,
				DateFormat:      dateFormat,
				TimeFormat:      timeFormat,
			},
		}

		return &result, nil

	default:
		return nil, fmt.Errorf("Unknown output datasource type: %q", datasource)
	}
}

//TODO expand to handle non-string types
//Include resource that is broken in error message
func extractAndValidateRequiredProperty(props map[string]interface{}) func(string) (*string, error) {
	return func(k string) (*string, error) {
		value := props[k].(string)
		if len(value) > 0 {
			return utils.String(value), nil
		}
		return nil, fmt.Errorf("Missing value for required property `%s`", k)
	}
}

func extractOptionalProperty(props map[string]interface{}) func(string) *string {
	return func (k string) *string {
		value := props[k].(string)
		if len(value) > 0 {
			return utils.String(value)
		}

		return nil
	}
}

func parseArmStreamAnalyticsReferenceDatasource(data map[string]interface{}) (streamanalytics.BasicReferenceInputDataSource, error) {
	name := data["name"].(string)
	datasource := data["datasource"].(string)

	switch datasource {
	case string(streamanalytics.TypeBasicReferenceInputDataSourceTypeMicrosoftStorageBlob):

		if data["storage_account_name"] == nil {
			return nil, fmt.Errorf("Reference input %s missing storage_account_name field", name)
		}
		accountName := data["storage_account_name"].(string)
		if data["storage_account_key"] == nil {
			return nil, fmt.Errorf("Reference input %s missing storage_account_key field", name)
		}
		accountKey := data["storage_account_key"].(string)

		if data["storage_container"] == nil {
			return nil, fmt.Errorf("Reference input %s missing storage_container field", name)
		}
		container := data["storage_container"].(string)

		if data["storage_path_pattern"] == nil {
			return nil, fmt.Errorf("Reference input %s missing storage_path_pattern field", name)
		}
		pathPattern := data["storage_path_pattern"].(string)
		if data["storage_date_format"] == nil {
			return nil, fmt.Errorf("Reference input %s missing storage_date_format field", name)
		}
		dateFormat := data["storage_date_format"].(string)
		if data["storage_time_format"] == nil {
			return nil, fmt.Errorf("Reference input %s missing storage_time_format field", name)
		}
		timeFormat := data["storage_time_format"].(string)

		accounts := make([]streamanalytics.StorageAccount, 1)
		accounts[0].AccountName = &accountName
		accounts[0].AccountKey = &accountKey

		return &streamanalytics.BlobReferenceInputDataSource{
			Type: streamanalytics.TypeBasicReferenceInputDataSourceTypeMicrosoftStorageBlob,
			BlobReferenceInputDataSourceProperties: &streamanalytics.BlobReferenceInputDataSourceProperties{
				StorageAccounts: &accounts,
				Container:       &container,
				PathPattern:     &pathPattern,
				DateFormat:      &dateFormat,
				TimeFormat:      &timeFormat,
			},
		}, nil

	default:
		return nil, fmt.Errorf("Reference input %s has unknown datasource type: %q", name, datasource)
	}
}

func parseArmStreamAnalyticsStreamDatasource(data map[string]interface{}) (streamanalytics.BasicStreamInputDataSource, error) {
	name := data["name"].(string)
	datasource := data["datasource"].(string)

	switch datasource {
	case string(streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftDevicesIotHubs):
		if data["iot_hub_namespace"] == nil {
			return nil, fmt.Errorf("Input %s missing iot_hub_namespace field", name)
		}
		if data["iot_hub_endpoint"] == nil {
			return nil, fmt.Errorf("Input %s missing iot_hub_endpoint field", name)
		}
		if data["shared_access_policy_name"] == nil {
			return nil, fmt.Errorf("Input %s missing shared_access_policy_name field", name)
		}
		if data["shared_access_policy_key"] == nil {
			return nil, fmt.Errorf("Input %s missing shared_access_policy_key field", name)
		}

		namespace := data["iot_hub_namespace"].(string)
		endpoint := data["iot_hub_endpoint"].(string)
		accessPolicyName := data["shared_access_policy_name"].(string)
		accessPolicyKey := data["shared_access_policy_key"].(string)

		result := streamanalytics.IoTHubStreamInputDataSource{
			Type: streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftDevicesIotHubs,
			IoTHubStreamInputDataSourceProperties: &streamanalytics.IoTHubStreamInputDataSourceProperties{
				IotHubNamespace:        &namespace,
				SharedAccessPolicyName: &accessPolicyName,
				SharedAccessPolicyKey:  &accessPolicyKey,
				Endpoint:               &endpoint,
			},
		}

		if v := data["consumer_group_name"]; v != nil {
			consumerGroupName := v.(string)
			result.IoTHubStreamInputDataSourceProperties.ConsumerGroupName = &consumerGroupName
		}

		return &result, nil

	case string(streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftServiceBusEventHub):
		if data["service_bus_namespace"] == nil {
			return nil, fmt.Errorf("Input %s missing service_bus_namespace field", name)
		}
		if data["eventhub_name"] == nil {
			return nil, fmt.Errorf("Input %s missing eventhub_name field", name)
		}
		if data["shared_access_policy_name"] == nil {
			return nil, fmt.Errorf("Input %s missing shared_access_policy_name field", name)
		}
		if data["shared_access_policy_key"] == nil {
			return nil, fmt.Errorf("Input %s missing shared_access_policy_key field", name)
		}

		serviceBusNamespace := data["service_bus_namespace"].(string)
		eventHubName := data["eventhub_name"].(string)
		accessPolicyName := data["shared_access_policy_name"].(string)
		accessPolicyKey := data["shared_access_policy_key"].(string)

		result := streamanalytics.EventHubStreamInputDataSource{
			Type: streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftServiceBusEventHub,
			EventHubStreamInputDataSourceProperties: &streamanalytics.EventHubStreamInputDataSourceProperties{
				SharedAccessPolicyName: &accessPolicyName,
				SharedAccessPolicyKey:  &accessPolicyKey,
				ServiceBusNamespace:    &serviceBusNamespace,
				EventHubName:           &eventHubName,
			},
		}

		if v := data["consumer_group_name"]; v != nil {
			consumerGroupName := v.(string)
			result.EventHubStreamInputDataSourceProperties.ConsumerGroupName = &consumerGroupName
		}

		return &result, nil

	case string(streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftStorageBlob):
		// TODO Doublecheck required fields, etc.
		accountName := data["storage_account_name"].(string)
		accountKey := data["storage_account_key"].(string)
		container := data["storage_container"].(string)
		pathPattern := data["storage_path_pattern"].(string)
		dateFormat := data["storage_date_format"].(string)
		timeFormat := data["storage_time_format"].(string)
		partitionCount := int32(data["storage_source_partition_count"].(int))

		accounts := make([]streamanalytics.StorageAccount, 1)
		accounts[0].AccountName = &accountName
		accounts[0].AccountKey = &accountKey

		return &streamanalytics.BlobStreamInputDataSource{
			Type: streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftStorageBlob,
			BlobStreamInputDataSourceProperties: &streamanalytics.BlobStreamInputDataSourceProperties{
				StorageAccounts:      &accounts,
				Container:            &container,
				PathPattern:          &pathPattern,
				DateFormat:           &dateFormat,
				TimeFormat:           &timeFormat,
				SourcePartitionCount: &partitionCount,
			},
		}, nil

	default:
		return nil, fmt.Errorf("Unknown stream input datasource type: %q", datasource)
	}
}

// Parse the serialization parameters out of an `input{...}` or `output{...}`
// section.
func parseArmStreamAnalyticsSerialization(data map[string]interface{}) (streamanalytics.BasicSerialization, error) {
	serialization := data["serialization"].(string)

	switch strings.ToUpper(serialization) {
	case "AVRO":
		return &streamanalytics.AvroSerialization{
			Type: streamanalytics.TypeAvro,
		}, nil

	case "JSON":
		if data["json_format"] == nil {
			return nil, fmt.Errorf("Serialization %s requires json_format field", serialization)
		}
		if data["encoding"] == nil {
			return nil, fmt.Errorf("Serialization %s requires encoding field", serialization)
		}

		encoding := streamanalytics.Encoding(data["encoding"].(string))
		format := streamanalytics.JSONOutputSerializationFormat(data["json_format"].(string))

		return &streamanalytics.JSONSerialization{
			Type: streamanalytics.TypeJSON,
			JSONSerializationProperties: &streamanalytics.JSONSerializationProperties{
				Encoding: encoding,
				Format:   format,
			},
		}, nil

	case "CSV":
		if data["delimiter"] == nil {
			return nil, fmt.Errorf("Serialization %s requires delimiter field", serialization)
		}
		if data["encoding"] == nil {
			return nil, fmt.Errorf("Serialization %s requires encoding field", serialization)
		}

		delimiter := data["delimiter"].(string)
		encoding := streamanalytics.Encoding(data["encoding"].(string))

		return &streamanalytics.CsvSerialization{
			Type: streamanalytics.TypeCsv,
			CsvSerializationProperties: &streamanalytics.CsvSerializationProperties{
				Encoding:       encoding,
				FieldDelimiter: &delimiter,
			},
		}, nil

	default:
		return nil, fmt.Errorf("Unknown serialization format %q; expected: Avro, Json, Csv", serialization)
	}
}

func flattenAndSetArmStreamAnalyticsSerialization(d *map[string]interface{}, input streamanalytics.BasicSerialization) error {
	result := *d

	if _, ok := input.AsAvroSerialization(); ok {
		result["serialization"] = string(streamanalytics.TypeAvro)
	} else if ser, ok := input.AsCsvSerialization(); ok {
		result["serialization"] = string(streamanalytics.TypeCsv)
		result["encoding"] = string(ser.CsvSerializationProperties.Encoding)
		result["delimiter"] = ser.CsvSerializationProperties.FieldDelimiter
	} else if ser, ok := input.AsJSONSerialization(); ok {
		result["serialization"] = string(streamanalytics.TypeJSON)
		result["encoding"] = string(ser.JSONSerializationProperties.Encoding)
		result["json_format"] = string(ser.JSONSerializationProperties.Format)
	} else {
		return fmt.Errorf("Unknown serialization description: %q", input)
	}

	return nil
}

// De-structure an Input into a map of parameters.
func flattenStreamAnalyticsJobInput(config *streamanalytics.Input) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	name := *config.Name
	result["name"] = name

	if properties, ok := config.Properties.AsStreamInputProperties(); ok {
		result["type"] = string(streamanalytics.TypeStream)

		if err := flattenAndSetArmStreamAnalyticsSerialization(&result, properties.Serialization); err != nil {
			return nil, err
		}

		if ds, ok := properties.Datasource.AsBlobStreamInputDataSource(); ok {
			props := ds.BlobStreamInputDataSourceProperties
			accounts := *props.StorageAccounts
			account := accounts[0]
			result["storage_account_name"] = account.AccountName
			if account.AccountKey != nil {
				result["storage_account_key"] = account.AccountKey
			}
			result["storage_container"] = props.Container
			result["storage_path_pattern"] = props.PathPattern
			result["storage_date_format"] = props.DateFormat
			result["storage_time_format"] = props.TimeFormat
			result["storage_source_partition_count"] = props.SourcePartitionCount
		} else if ds, ok := properties.Datasource.AsEventHubStreamInputDataSource(); ok {
			props := ds.EventHubStreamInputDataSourceProperties
			result["shared_access_policy_name"] = props.SharedAccessPolicyName
			result["shared_access_policy_key"] = props.SharedAccessPolicyKey
			result["consumer_group_name"] = props.ConsumerGroupName
			result["service_bus_namespace"] = props.ServiceBusNamespace
			result["eventhub_name"] = props.EventHubName
		} else if ds, ok := properties.Datasource.AsIoTHubStreamInputDataSource(); ok {
			props := ds.IoTHubStreamInputDataSourceProperties
			result["shared_access_policy_name"] = props.SharedAccessPolicyName
			result["shared_access_policy_key"] = props.SharedAccessPolicyKey
			result["consumer_group_name"] = props.ConsumerGroupName
			result["iot_hub_namespace"] = props.IotHubNamespace
			result["iot_hub_endpoint"] = props.Endpoint
		} else {
			return nil, fmt.Errorf("Stream input %s has unknown datasource; expected configuration for blob, event hub, or IoT hub", name)
		}
	} else if properties, ok := config.Properties.AsReferenceInputProperties(); ok {
		result["type"] = string(streamanalytics.TypeReference)

		if err := flattenAndSetArmStreamAnalyticsSerialization(&result, properties.Serialization); err != nil {
			return nil, err
		}

		if ds, ok := properties.Datasource.AsBlobReferenceInputDataSource(); ok {
			props := ds.BlobReferenceInputDataSourceProperties
			if props.StorageAccounts != nil && len(*props.StorageAccounts) > 0 {
				accounts := *props.StorageAccounts
				account := accounts[0]
				result["storage_account_name"] = account.AccountName
				if account.AccountKey != nil {
					result["storage_account_key"] = account.AccountKey
				}
			}
			result["storage_container"] = props.Container
			result["storage_path_pattern"] = props.PathPattern
			result["storage_date_format"] = props.DateFormat
			result["storage_time_format"] = props.TimeFormat
		} else {
			return nil, fmt.Errorf("Reference input %s has unknown datasource; expected configuration for blob", name)
		}
	} else {
		return nil, fmt.Errorf("Input %s has unknown input type; expected %s, %s",
			name, streamanalytics.TypeStream, streamanalytics.TypeReference)
	}

	return result, nil
}

func flattenStreamAnalyticsJobTransformation(config *streamanalytics.Transformation) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	result["name"] = *config.Name

	result["query"] = *config.Query

	result["streaming_units"] = *config.StreamingUnits

	return result, nil
}

func flattenStreamAnalyticsJobOutput(config *streamanalytics.Output) (*map[string]interface{}, error) {
	result := make(map[string]interface{})

	if config.Name == nil {
		return nil, fmt.Errorf("Received output with no name: %+v", *config)
	}
	name := *config.Name
	result["name"] = name

	properties := *config.OutputProperties

	if _, ok := properties.Datasource.AsAzureSQLDatabaseOutputDataSource(); !ok {
		if err := flattenAndSetArmStreamAnalyticsSerialization(&result, properties.Serialization); err != nil {
			return nil, fmt.Errorf("Nope")
		}
	}

	if ds, ok := properties.Datasource.AsAzureDataLakeStoreOutputDataSource(); ok {
		result["datasource"] = string(streamanalytics.TypeMicrosoftDataLakeAccounts)
		props := ds.AzureDataLakeStoreOutputDataSourceProperties
		result["data_lake_refresh_token"] = props.RefreshToken
		result["data_lake_token_user_name"] = props.TokenUserPrincipalName
		result["data_lake_token_display_name"] = props.TokenUserDisplayName
		result["data_lake_account_name"] = props.AccountName
		result["data_lake_tenant_id"] = props.TenantID
		result["data_lake_path_prefix"] = props.FilePathPrefix
		result["data_lake_date_format"] = props.DateFormat
		result["data_lake_time_format"] = props.TimeFormat
	} else if ds, ok := properties.Datasource.AsBlobOutputDataSource(); ok {
		result["datasource"] = string(streamanalytics.TypeMicrosoftStorageBlob)
		props := ds.BlobOutputDataSourceProperties
		account := (*props.StorageAccounts)[0]
		result["storage_account_name"] = account.AccountName
		if account.AccountKey != nil {
			result["storage_account_key"] = account.AccountKey
		}
		result["storage_container"] = props.Container
		result["storage_path_pattern"] = props.PathPattern
		result["storage_date_format"] = props.DateFormat
		result["storage_time_format"] = props.TimeFormat
	} else if ds, ok := properties.Datasource.AsAzureSQLDatabaseOutputDataSource(); ok {
		result["datasource"] = string(streamanalytics.TypeMicrosoftSQLServerDatabase)
		props := ds.AzureSQLDatabaseOutputDataSourceProperties
		result["database_server"] = props.Server
		result["database_name"] = props.Database
		result["database_table"] = props.Table
		result["database_user"] = props.User
		result["database_password"] = props.Password
	} else if ds, ok := properties.Datasource.AsEventHubOutputDataSource(); ok {
		result["datasource"] = string(streamanalytics.TypeMicrosoftServiceBusEventHub)
		props := ds.EventHubOutputDataSourceProperties
		result["shared_access_policy_name"] = props.SharedAccessPolicyName
		result["shared_access_policy_key"] = props.SharedAccessPolicyKey
		result["service_bus_namespace"] = props.ServiceBusNamespace
		result["eventhub_name"] = props.EventHubName
		result["eventhub_partition_key"] = props.PartitionKey
	} else if _, ok := properties.Datasource.AsPowerBIOutputDataSource(); ok {
		// Unsupported
		return nil, fmt.Errorf("Output %s has unsupported datasource: %s",
			name, streamanalytics.TypePowerBI)
	} else if _, ok := properties.Datasource.AsServiceBusTopicOutputDataSource(); ok {
		// Unsupported
		return nil, fmt.Errorf("Output %s has unsupported datasource: %s",
			name, streamanalytics.TypeMicrosoftServiceBusTopic)
	} else if _, ok := properties.Datasource.AsServiceBusQueueOutputDataSource(); ok {
		// Unsupported
		return nil, fmt.Errorf("Output %s has unsupported datasource: %s",
			name, streamanalytics.TypeMicrosoftServiceBusQueue)
	} else if _, ok := properties.Datasource.AsDocumentDbOutputDataSource(); ok {
		// Unsupported
		return nil, fmt.Errorf("Output %s has unsupported datasource: %s",
			name, streamanalytics.TypeMicrosoftStorageDocumentDB)
	} else if _, ok := properties.Datasource.AsAzureTableOutputDataSource(); ok {
		// Unsupported
		return nil, fmt.Errorf("Output %s has unsupported datasource: %s",
			name, streamanalytics.TypeMicrosoftStorageTable)
	} else {
		return nil, fmt.Errorf("Output %s has unknown datasource; expected %s, %s, %s, %s",
			name,
			streamanalytics.TypeMicrosoftDataLakeAccounts,
			streamanalytics.TypeMicrosoftSQLServerDatabase,
			streamanalytics.TypeMicrosoftServiceBusEventHub,
			streamanalytics.TypeMicrosoftStorageBlob)
	}

	return &result, nil
}

func validateInputType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case string(streamanalytics.TypeReference):
		return
	case string(streamanalytics.TypeStream):
		return
	default:
		errors = append(errors, fmt.Errorf("%s: Unknown input type: %q; must be %q or %q", k, value,
			streamanalytics.TypeStream, streamanalytics.TypeReference))
	}

	return
}

func validateArmStreamAnalyticsInputDatasource(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case string(streamanalytics.TypeBasicReferenceInputDataSourceTypeMicrosoftStorageBlob):
		return
	case string(streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftDevicesIotHubs):
		return
	case string(streamanalytics.TypeBasicStreamInputDataSourceTypeMicrosoftServiceBusEventHub):
		return
	default:
		errors = append(errors, fmt.Errorf("Unknown input datasource: %q", value))
	}

	return
}

func validateArmStreamAnalyticsOutputDatasource(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case string(streamanalytics.TypeMicrosoftDataLakeAccounts):
		return
	case string(streamanalytics.TypeMicrosoftServiceBusEventHub):
		return
	case string(streamanalytics.TypeMicrosoftSQLServerDatabase):
		return
	case string(streamanalytics.TypeMicrosoftStorageBlob):
		return
	case string(streamanalytics.TypeMicrosoftServiceBusQueue):
		errors = append(errors, fmt.Errorf("Output datasource not currently supported: %q", value))
	case string(streamanalytics.TypeMicrosoftServiceBusTopic):
		errors = append(errors, fmt.Errorf("Output datasource not currently supported: %q", value))
	case string(streamanalytics.TypeMicrosoftStorageDocumentDB):
		errors = append(errors, fmt.Errorf("Output datasource not currently supported: %q", value))
	case string(streamanalytics.TypeMicrosoftStorageTable):
		errors = append(errors, fmt.Errorf("Output datasource not currently supported: %q", value))
	case string(streamanalytics.TypePowerBI):
		errors = append(errors, fmt.Errorf("Output datasource not currently supported: %q", value))
	default:
		errors = append(errors, fmt.Errorf("Unknown output datasource: %q", value))
	}

	return
}

func validateOutOfOrderPolicy(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case "adjust":
		return
	case "drop":
		return
	default:
		errors = append(errors, fmt.Errorf("Unknown out of order policy: %q; expected adjust, drop", value))
	}

	return
}

func validateOutputErrorPolicy(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case "stop":
		return
	case "drop":
		return
	default:
		errors = append(errors, fmt.Errorf("Unknown output error policy: %q; expected drop, stop", value))
	}

	return
}

func validateSerializationType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case string(streamanalytics.TypeAvro):
		return
	case string(streamanalytics.TypeCsv):
		return
	case string(streamanalytics.TypeJSON):
		return
	default:
		errors = append(errors, fmt.Errorf("Unknown serialization type: %q; expected %q, %q, %q", value,
			streamanalytics.TypeAvro, streamanalytics.TypeCsv, streamanalytics.TypeJSON))
	}

	return
}

func validateEncoding(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if v != "UTF8" {
		errors = append(errors, fmt.Errorf("Unknown encoding: %q; expected UTF8", value))
	}

	return
}

func validateSku(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case "Standard":
	default:
		errors = append(errors, fmt.Errorf("Unknown SKU: %q; expected 'Standard'", value))
	}
	return
}

func validateOutputStartMode(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case string(streamanalytics.JobStartTime):
	case string(streamanalytics.CustomTime):
	case string(streamanalytics.LastOutputEventTime):
	default:
		errors = append(errors, fmt.Errorf("Unknown Output Start Mode: %q; expected %q, %q, %q", value,
			streamanalytics.JobStartTime, streamanalytics.CustomTime, streamanalytics.LastOutputEventTime))
	}

	return
}

func validateOutput(v interface{}, k string) (warnings []string, errors []error) {

	ws, es := validateOutputDatasource(v, k)
	warnings = append(warnings, ws...)
	errors = append(errors, es...)

	return
}

func validateOutputSerialization(v interface{}, k string) (ws []string, errors []error) {
	data := v.(map[string]interface{})

	serialization := data["serialization"]

	if serialization == nil {
		errors = append(errors, fmt.Errorf("required serialization format in output %s", k))
		return
	}

	switch serialization.(string) {
	case string(streamanalytics.TypeAvro):
	case string(streamanalytics.TypeCsv):
	case string(streamanalytics.TypeJSON):
	default:
		errors = append(errors, fmt.Errorf("unknown serialization format %s in output %s", serialization, k))
	}

	return
}

// ValidateFunc: the datasource-specific fields needed are present.
func validateOutputDatasource(v interface{}, k string) (warnings []string, errors []error) {
	data := v.(map[string]interface{})

	datasource := data["datasource"]
	if datasource == nil {
		errors = append(errors, fmt.Errorf("%s: datasource is a required field", k))
		return
	}

	switch datasource.(string) {
	case string(streamanalytics.TypeMicrosoftDataLakeAccounts):
		if data["data_lake_refresh_token"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_refresh_token is required when using %s", k, datasource))
		}
		if data["data_lake_token_user_name"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_token_user_name is required when using %s", k, datasource))
		}
		if data["data_lake_token_display_name"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_token_display_name is required when using %s", k, datasource))
		}
		if data["data_lake_account_name"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_account_name is required when using %s", k, datasource))
		}
		if data["data_lake_tenant_id"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_tenant_id is required when using %s", k, datasource))
		}
		if data["data_lake_path_prefix"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_path_prefix is required when using %s", k, datasource))
		}
		if data["data_lake_date_format"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_date_format is required when using %s", k, datasource))
		}
		if data["data_lake_time_format"] == nil {
			errors = append(errors, fmt.Errorf("%s: data_lake_time_format is required when using %s", k, datasource))
		}
		ws, es := validateOutputSerialization(v, k)
		warnings = append(warnings, ws...)
		errors = append(errors, es...)

	case string(streamanalytics.TypeMicrosoftServiceBusEventHub):
		if data["service_bus_namespace"] == nil {
			errors = append(errors, fmt.Errorf("%s: service_bus_namespace is required when using %s", k, datasource))
		}
		if data["eventhub_name"] == nil {
			errors = append(errors, fmt.Errorf("%s: eventhub_name is required when using %s", k, datasource))
		}
		if data["eventhub_partition_key"] == nil {
			errors = append(errors, fmt.Errorf("%s: eventhub_partition_key is required when using %s", k, datasource))
		}
		if data["shared_access_policy_name"] == nil {
			errors = append(errors, fmt.Errorf("%s: shared_access_policy_name is required when using %s", k, datasource))
		}
		if data["shared_access_policy_key"] == nil {
			errors = append(errors, fmt.Errorf("%s: shared_access_policy_key is required when using %s", k, datasource))
		}
		ws, es := validateOutputSerialization(v, k)
		warnings = append(warnings, ws...)
		errors = append(errors, es...)

	case string(streamanalytics.TypeMicrosoftSQLServerDatabase):
		if data["database_server"] == nil {
			errors = append(errors, fmt.Errorf("%s: database_server is required when using %s", k, datasource))
		}
		if data["database_name"] == nil {
			errors = append(errors, fmt.Errorf("%s: database_name is required when using %s", k, datasource))
		}
		if data["database_table"] == nil {
			errors = append(errors, fmt.Errorf("%s: database_table is required when using %s", k, datasource))
		}
		if data["database_user"] == nil {
			errors = append(errors, fmt.Errorf("%s: database_user is required when using %s", k, datasource))
		}
		if data["database_password"] == nil {
			errors = append(errors, fmt.Errorf("%s: database_password is required when using %s", k, datasource))
		}
	default:
		errors = append(errors, fmt.Errorf("%s: %s is not supported", k, datasource))
	}
	return
}
