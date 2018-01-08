package azurerm

import (
	"fmt"
	"log"
	"regexp"
	"time"

	dataLakeStore "github.com/Azure/azure-sdk-for-go/arm/datalake-store/account"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmDataLakeStoreAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmDataLakeStoreAccountCreate,
		Read:   resourceArmDataLakeStoreAccountRead,
		Update: resourceArmDataLakeStoreAccountUpdate,
		Delete: resourceArmDataLakeStoreAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArmDataLakeStoreAccountName,
			},
			"tier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: validateArmDataLakeStoreTierType,
			},
			"location":            locationSchema(),
			"tags":                tagsSchema(),
			"resource_group_name": resourceGroupNameDiffSuppressSchema(),
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateArmDataLakeStoreAccountName(v interface{}, k string) (ws []string, es []error) {
	input := v.(string)

	// TODO
	if !regexp.MustCompile(`\A([a-z0-9]{3,24})\z`).MatchString(input) {
		es = append(es, fmt.Errorf("name can only consist of lowercase letters and numbers, and must be between 3 and 24 characters long"))
	}

	return
}

func validateArmDataLakeStoreTierType(v interface{}, k string) (ws []string, es []error) {
	input := dataLakeStore.TierType(v.(string))

	validAccountTypes := []dataLakeStore.TierType{
		dataLakeStore.Consumption,
		dataLakeStore.Commitment1TB,
		dataLakeStore.Commitment10TB,
		dataLakeStore.Commitment100TB,
		dataLakeStore.Commitment500TB,
		dataLakeStore.Commitment1PB,
		dataLakeStore.Commitment5PB,
	}

	for _, valid := range validAccountTypes {
		if valid == input {
			return
		}
	}

	es = append(es, fmt.Errorf("Invalid data lake store tier %q", input))
	return
}

func resourceArmDataLakeStoreAccountCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	dls := client.dataLakeStoreClient

	resourceGroupName := d.Get("resource_group_name").(string)
	dlsAccountName := d.Get("name").(string)

	location := d.Get("location").(string)
	tier := d.Get("tier").(string)
	tags := d.Get("tags").(map[string]interface{})

	parameters := dataLakeStore.DataLakeStoreAccount{
		Name:     &dlsAccountName,
		Location: &location,
		Tags:     expandTags(tags),
		// TODO: Identity: EncryptionIdentity{}
		DataLakeStoreAccountProperties: &dataLakeStore.DataLakeStoreAccountProperties{
			NewTier: dataLakeStore.TierType(tier),
		},
	}

	_, createError := dls.Create(resourceGroupName, dlsAccountName, parameters, make(chan struct{}))
	createErr := <-createError

	// The only way to get the ID back apparently is to read the resource again
	read, err := dls.Get(resourceGroupName, dlsAccountName)

	// Set the ID right away if we have one
	if err == nil && read.ID != nil {
		log.Printf("[INFO] Data Lake Store Account %q ID: %q", dlsAccountName, *read.ID)
		d.SetId(*read.ID)
	}

	// If we had a create error earlier then we return with that error now.
	// We do this later here so that we can grab the ID above is possible.
	if createErr != nil {
		return fmt.Errorf(
			"Error creating Azure Data Lake Store Account: %q: %+v",
			dlsAccountName, createErr)
	}

	// Check the read error now that we know it would exist without a create err
	if err != nil {
		return err
	}

	// If we got no ID then the resource group doesn't yet exist
	if read.ID == nil {
		return fmt.Errorf("Cannot read Data Lake Store Account %q (resource group %q) ID",
			dlsAccountName, resourceGroupName)
	}

	log.Printf("[DEBUG] Waiting for Data Lake Store Account (%s) to become available", dlsAccountName)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Updating", "Creating"},
		Target:     []string{"Succeeded"},
		Refresh:    dlsAccountStateRefreshFunc(client, resourceGroupName, dlsAccountName),
		Timeout:    30 * time.Minute,
		MinTimeout: 15 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for Data Lake Store Account (%s) to become available: %s", dlsAccountName, err)
	}

	return resourceArmDataLakeStoreAccountRead(d, meta)
}

func resourceArmDataLakeStoreAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).dataLakeStoreClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	dlsAccountName := id.Path["accounts"]
	resourceGroupName := id.ResourceGroup

	resp, err := client.Get(resourceGroupName, dlsAccountName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading the state of Data Lake Store Account %q: %+v", dlsAccountName, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroupName)
	d.Set("location", azureRMNormalizeLocation(*resp.Location))
	// TODO Identity

	if props := resp.DataLakeStoreAccountProperties; props != nil {
		d.Set("tier", props.CurrentTier)

		if endpoint := props.Endpoint; endpoint != nil {
			d.Set("endpoint", endpoint)
		}

		if accountid := props.AccountID; accountid != nil {
			d.Set("account_id", accountid.String)
		}
		// TODO All the other stuff in DataLakeStoreAccountProperties?
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmDataLakeStoreAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).dataLakeStoreClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	dlsAccountName := id.Path["accounts"]
	resourceGroupName := id.ResourceGroup

	d.Partial(true)

	if d.HasChange("tier") {
		tier := dataLakeStore.TierType(d.Get("tier").(string))

		opts := dataLakeStore.DataLakeStoreAccountUpdateParameters{
			UpdateDataLakeStoreAccountProperties: &dataLakeStore.UpdateDataLakeStoreAccountProperties{
				NewTier: tier,
			},
		}

		_, updateError := client.Update(resourceGroupName, dlsAccountName, opts, make(chan struct{}))
		if err := <-updateError; err != nil {
			return fmt.Errorf("Error updating Azure Data Lake Store Account tier %q: %+v", dlsAccountName, err)
		}

		d.SetPartial("tier")
	}

	if d.HasChange("tags") {
		tags := d.Get("tags").(map[string]interface{})

		opts := dataLakeStore.DataLakeStoreAccountUpdateParameters{
			Tags: expandTags(tags),
		}
		_, updateError := client.Update(resourceGroupName, dlsAccountName, opts, make(chan struct{}))
		if err := <-updateError; err != nil {
			return fmt.Errorf("Error updating Azure Data Lake Store Account tags %q: %+v", dlsAccountName, err)
		}

		d.SetPartial("tags")
	}

	d.Partial(false)
	return nil
}

func resourceArmDataLakeStoreAccountDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).dataLakeStoreClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	dlsAccountName := id.Path["accounts"]
	resourceGroupName := id.ResourceGroup

	_, deleteErr := client.Delete(resourceGroupName, dlsAccountName, make(chan struct{}))
	if err := <-deleteErr; err != nil {
		return fmt.Errorf("Error issuing AzureRM delete request for Data Lake Store Account %q: %+v",
			dlsAccountName, err)
	}

	return nil
}

func dlsAccountStateRefreshFunc(client *ArmClient, resourceGroupName string, storageAccountName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.dataLakeStoreClient.Get(resourceGroupName, storageAccountName)
		if err != nil {
			return nil, "", fmt.Errorf("Error issuing read request in dlsAccountStateRefreshFunc to Azure ARM for Data Lake Store Account '%s' (RG: '%s'): %s", storageAccountName, resourceGroupName, err)
		}

		return res, string(res.DataLakeStoreAccountProperties.ProvisioningState), nil
	}
}
