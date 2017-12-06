package azurerm

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/eventhub"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmEventHub() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmEventHubCreate,
		Read:   resourceArmEventHubRead,
		Update: resourceArmEventHubCreate,
		Delete: resourceArmEventHubDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": resourceGroupNameSchema(),

			// TODO: remove me in the next major version
			"location": deprecatedLocationSchema(),

			"partition_count": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateEventHubPartitionCount,
			},

			"message_retention": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateEventHubMessageRetentionCount,
			},

			"partition_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},

			"capture": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"encoding": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateEventHubCaptureEncoding,
						},
						"time_limit": {
							Type:        schema.TypeInt,
							Description: "Maximimum time between writes (seconds)",
							Required:    true,
						},
						"size_limit": {
							Type:        schema.TypeInt,
							Description: "Maximum buffering between writes (bytes)",
							Required:    true,
						},
						"archive_name_format": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "{Namespace}/{EventHub}/{PartitionId}/{Year}/{Month}/{Day}/{Hour}/{Minute}/{Second}",
						},
						"destination": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateEventHubCaptureDestination,
						},
						// Required when destination==AzureBlockBlob
						"storage_account_resource_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"storage_blob_container_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						// Required when destination==AzureDataLake
						"data_lake_subscription_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_account_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"data_lake_folder_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

const (
	captureDestinationPrefix = "EventHubArchive."
)

func resourceArmEventHubCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	eventhubClient := client.eventHubClient
	log.Printf("[INFO] preparing arguments for Azure ARM EventHub creation.")

	name := d.Get("name").(string)
	namespaceName := d.Get("namespace_name").(string)
	resGroup := d.Get("resource_group_name").(string)
	partitionCount := int64(d.Get("partition_count").(int))
	messageRetention := int64(d.Get("message_retention").(int))

	parameters := eventhub.Model{
		Properties: &eventhub.Properties{
			PartitionCount:         &partitionCount,
			MessageRetentionInDays: &messageRetention,
		},
	}

	if v, ok := d.GetOk("capture"); ok {
		captures := v.([]interface{})
		if len(captures) > 0 {
			capture := captures[0].(map[string]interface{})
			enabled := capture["enabled"].(bool)

			encodingName := capture["encoding"].(string)
			var encoding eventhub.EncodingCaptureDescription

			switch encodingName {
			case "Avro":
				encoding = eventhub.Avro
			case "AvroDeflat":
				encoding = eventhub.AvroDeflate
			default:
				return fmt.Errorf("Unknown Encoding Capture Description %s; expected %q, %q",
					encodingName, eventhub.Avro, eventhub.AvroDeflate)
			}

			time := int32(capture["time_limit"].(int))
			size := int32(capture["size_limit"].(int))

			archiveName := capture["archive_name_format"].(string)

			destination := capture["destination"].(string)
			destinationName := captureDestinationPrefix + destination

			props := &eventhub.DestinationProperties{
				ArchiveNameFormat: &archiveName,
			}

			switch destinationName {
			case "EventHubArchive.AzureDataLake":
				storageAccount := capture["storage_account_resource_id"].(string)
				blobContainer := capture["storage_blob_container_name"].(string)

				props.StorageAccountResourceID = &storageAccount
				props.BlobContainer = &blobContainer
			case "EventHubArchive.AzureBlockBlob":
				sub := capture["data_lake_subscription_id"].(string)
				name := capture["data_lake_account_name"].(string)
				path := capture["data_lake_folder_path"].(string)

				props.DataLakeSubscriptionID = &sub
				props.DataLakeAccountName = &name
				props.DataLakeFolderPath = &path
			default:
				return fmt.Errorf("Unknown capture destination %s; expected %q or %q",
					destination, "AzureBlockBlob", "AzureDataLake")
			}

			parameters.Properties.CaptureDescription = &eventhub.CaptureDescription{
				Enabled:           &enabled,
				Encoding:          encoding,
				IntervalInSeconds: &time,
				SizeLimitInBytes:  &size,
				Destination: &eventhub.Destination{
					Name: &destinationName,
					DestinationProperties: props,
				},
			}
		}
	}

	_, err := eventhubClient.CreateOrUpdate(resGroup, namespaceName, name, parameters)
	if err != nil {
		return err
	}

	read, err := eventhubClient.Get(resGroup, namespaceName, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read EventHub %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmEventHubRead(d, meta)
}

func resourceArmEventHubRead(d *schema.ResourceData, meta interface{}) error {
	eventhubClient := meta.(*ArmClient).eventHubClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	namespaceName := id.Path["namespaces"]
	name := id.Path["eventhubs"]

	resp, err := eventhubClient.Get(resGroup, namespaceName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure EventHub %s: %s", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("namespace_name", namespaceName)
	d.Set("resource_group_name", resGroup)

	d.Set("partition_count", resp.Properties.PartitionCount)
	d.Set("message_retention", resp.Properties.MessageRetentionInDays)
	d.Set("partition_ids", resp.Properties.PartitionIds)

	if err := d.Set("capture", flattenAzureRmEventHubCapture(resp.Properties.CaptureDescription)); err != nil {
		return fmt.Errorf("[DEBUG] Error setting Event Hub Capture error: %#v", err)
	}

	return nil
}

func resourceArmEventHubDelete(d *schema.ResourceData, meta interface{}) error {
	eventhubClient := meta.(*ArmClient).eventHubClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	namespaceName := id.Path["namespaces"]
	name := id.Path["eventhubs"]

	resp, err := eventhubClient.Delete(resGroup, namespaceName, name)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error issuing Azure ARM delete request of EventHub'%s': %s", name, err)
	}

	return nil
}

func validateEventHubPartitionCount(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)

	if !(32 >= value && value >= 2) {
		errors = append(errors, fmt.Errorf("EventHub Partition Count has to be between 2 and 32"))
	}
	return
}

func validateEventHubMessageRetentionCount(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)

	if !(7 >= value && value >= 1) {
		errors = append(errors, fmt.Errorf("EventHub Retention Count has to be between 1 and 7"))
	}
	return
}

func validateEventHubCaptureDestination(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case "AzureDataLake":
		return
	case "AzureBlockBlob":
		return
	default:
		errors = append(errors, fmt.Errorf("EventHub Capture Destination has to be AzureBlockBlob"))
	}
	return
}

func validateEventHubCaptureEncoding(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	switch value {
	case string(eventhub.Avro):
		return
	case string(eventhub.AvroDeflate):
		return
	default:
		errors = append(errors, fmt.Errorf("Unknown Capture Encoding Description %q; expected %q, %q",
			value, eventhub.Avro, eventhub.AvroDeflate))
	}
	return
}

func flattenAzureRmEventHubCapture(capture *eventhub.CaptureDescription) []interface{} {
	result := make(map[string]interface{})

	if capture != nil {
		result["enabled"] = *capture.Enabled

		result["encoding"] = string(capture.Encoding)

		result["time_limit"] = *capture.IntervalInSeconds
		result["size_limit"] = *capture.SizeLimitInBytes

		result["archive_name_format"] = *capture.Destination.DestinationProperties.ArchiveNameFormat

		dest := strings.Split(*capture.Destination.Name, ".")[1]
		result["destination"] = dest
		if dest == "AzureDataLake" {
			result["data_lake_subscription_id"] = *capture.Destination.DestinationProperties.DataLakeSubscriptionID
			result["data_lake_account_name"] = *capture.Destination.DestinationProperties.DataLakeAccountName
			result["data_lake_folder_path"] = *capture.Destination.DestinationProperties.DataLakeFolderPath
		} else if dest == "AzureBlockBlob" {
			result["storage_account_resource_id"] = *capture.Destination.DestinationProperties.StorageAccountResourceID
			result["storage_blob_container_name"] = *capture.Destination.DestinationProperties.BlobContainer
		}
	}

	return []interface{}{result}
}
