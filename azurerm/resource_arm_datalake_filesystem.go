package azurerm

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

// Manage Data Lake Store directories
//
// Manage directories and their ACLs on Data Lake Store file systems. We use
// 'mode' to set traditional user/group/other permissions and ACLs for other
// permissions.
//
// NB: In this file we're doing something of our own thing w.r.t. resource IDs.
// We are tacking some extra information onto the Azure Resource ID for a Data
// Lake Store Account to represent a directory.

func resourceArmDataLakeFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmDataLakeFileCreate,
		Read:   resourceArmDataLakeFileRead,
		Update: resourceArmDataLakeFileCreate,
		Delete: resourceArmDataLakeFileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"path": {
				Description: "Absolute path to the directory to manage.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"owner": {
				Description: "ID of the user who owns this path.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"group": {
				Description: "ID of the group who owns this path.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"mode": {
				Description:  "Octal permissions for owner, group, and others. E.g: 750",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateUnixPermissions,
			},

			"mask": {
				Description:  "Symbolic permissions mask for group and other entries. E.g: r-x",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateSymbolicPermissions,
			},

			"access_acl": {
				Description: "Additional access control rules: (user|group|other):entityid:(r|-)(w|-)(x|-)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateACLEntry,
				},
			},

			"default_acl": {
				Description: "Additional default access control rules: (user|group|other):entityid:(r|-)(w|-)(x|-)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateACLEntry,
				},
			},
		},
	}
}

// When completed: the path exists, and has the given mode and ACL rules.
func resourceArmDataLakeFileCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	dataLakeClient := client.dataLakeFsClient

	accountName := d.Get("account_name").(string)
	path := d.Get("path").(string)
	mode := int32(d.Get("mode").(int))
	perm := fmt.Sprintf("%03d", mode)

	status, statusErr := dataLakeClient.GetFileStatus(accountName, path, "GETFILESTATUS", nil)

	if status.Response.StatusCode == http.StatusNotFound {
		if path == "/" {
			_, err := dataLakeClient.SetPermission(accountName, path[1:], perm)
			if err != nil {
				return err
			}
		} else {
			_, err := dataLakeClient.Mkdirs(accountName, path, &mode)

			if err != nil {
				return err
			}
		}
	} else if statusErr != nil {
		return fmt.Errorf("Couldn't check existance of %s: %q", path, statusErr)
	} else {
		_, err := dataLakeClient.SetPermission(accountName, path[1:], perm)
		if err != nil {
			return fmt.Errorf("Error setting permissions on %s to %s: %q", path, perm, err)
		}
	}

	var rules []string

	if v, ok := d.GetOk("mask"); ok {
		rules = append(rules, "mask::"+v.(string))
	}

	// Add rules derived from permission.
	//
	// TODO: determine if and how the permissions should interact with default
	// ACL entries ("not" seems like it might be a sensible position here). If so
	// it seems like we should allow default user, group, other entries in the
	// default ACLs field?

	userPerm, _ := modeToPerm(mode / 100)
	rules = append(rules, "user::"+userPerm)
	// rules = append(rules, "default:user::"+userPerm)

	groupPerm, _ := modeToPerm((mode % 100) / 10)
	rules = append(rules, "group::"+groupPerm)
	// rules = append(rules, "default:group::"+groupPerm)

	otherPerm, _ := modeToPerm(mode % 10)
	rules = append(rules, "other::"+otherPerm)

	// TODO If our rule set does not contain any default rules, Azure will not
	// modify any default rules (this includes deleting the existing ones we no
	// longer mention). Including other seems like a not-too-terrible way to do
	// this?
	//
	// We'll want to fix this somehow. Probably by using d.GetChange("default_acl")
	// and managing the default rules directly.
	rules = append(rules, "default:other::"+otherPerm)

	if v, ok := d.GetOk("access_acl"); ok {
		for _, r := range v.([]interface{}) {
			rules = append(rules, r.(string))
		}
	}

	if v, ok := d.GetOk("default_acl"); ok {
		for _, r := range v.([]interface{}) {
			rules = append(rules, "default:"+r.(string))
		}
	}

	if len(rules) > 0 {
		aclSpec := strings.Join(rules, ",")
		resp, err := dataLakeClient.SetACL(accountName, path, aclSpec)

		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Failed to set ACL with code %d; expected %d", resp.StatusCode, http.StatusOK)
		}
	}

	id := fmt.Sprintf("adl://%s.%s/%s", accountName, dataLakeClient.ManagementClient.AdlsFileSystemDNSSuffix, path)
	d.SetId(id)

	return resourceArmDataLakeFileRead(d, meta)
}

func resourceArmDataLakeFileRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	datalakeFsClient := client.dataLakeFsClient

	// TODO Parse these from the ID?
	accountName := d.Get("account_name").(string)
	path := d.Get("path").(string)

	acl, err := datalakeFsClient.GetACLStatus(accountName, path, nil)

	if err != nil {
		if utils.ResponseWasNotFound(acl.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making read request on Data Lake Store path: %s", err)
	}

	d.Set("owner", acl.ACLStatus.Owner)
	d.Set("group", acl.ACLStatus.Group)

	var mode int
	if _, err := fmt.Sscanf(*acl.ACLStatus.Permission, "%d", &mode); err != nil {
		return fmt.Errorf("Couldn't parse int from mode '%s': %q", *acl.ACLStatus.Permission, err)
	}
	log.Printf("[DEBUG] Reading mode of %s as %d", path, mode)
	d.Set("mode", mode)

	accessACL := []string{}
	defaultACL := []string{}

	for _, r := range *acl.ACLStatus.Entries {
		parts := strings.Split(r, ":")

		switch len(parts) {
		case 4:
			rule := strings.Join(parts[1:], ":")
			if parts[0] == "default" {
				// We ignore the rules which are reflections of permission.
				if parts[2] != "" {
					defaultACL = append(defaultACL, rule)
				}
			} else {
				return fmt.Errorf("Error reading malformed ACL: %s", r)
			}
		case 3:
			if parts[1] != "" {
				accessACL = append(accessACL, r)
			}
		default:
			return fmt.Errorf("Error reading malformed ACL: %s", r)
		}
	}

	d.Set("access_acl", accessACL)
	d.Set("default_acl", defaultACL)

	return nil
}

func resourceArmDataLakeFileDelete(d *schema.ResourceData, meta interface{}) error {
	datalakeFsClient := meta.(*ArmClient).dataLakeFsClient

	accountName := d.Get("account_name").(string)
	path := d.Get("path").(string)

	// If it isn't empty, let's fail.
	recursive := false

	resp, err := datalakeFsClient.Delete(accountName, path, &recursive)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error issuing Azure ARM delete request of DataLake FileSystem: %s", err)
	}

	return nil
}

func validateSymbolicPermissions(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) != 3 {
		errors = append(errors, fmt.Errorf("%s is not a valid symbolic; expected (---|r--|-w-|--x|rw-|r-x|-wx|rwx)", value))
		return
	}

	if c := value[0]; c != '-' && c != 'r' {
		errors = append(errors, fmt.Errorf("%s is not a valid symbolic; expected - or r but saw %c", value, c))
	}
	if c := value[1]; c != '-' && c != 'w' {
		errors = append(errors, fmt.Errorf("%s is not a valid symbolic; expected - or w but saw %c", value, c))
	}
	if c := value[2]; c != '-' && c != 'x' {
		errors = append(errors, fmt.Errorf("%s is not a valid symbolic; expected - or x but saw %c", value, c))
	}

	return
}

// NB: We expect an integer who's decimal digits are, when interpreted as octal
// digits, the mode.
// E.g. The decimal value 751 (01357, b0010 1110 1111) is interpreted as rwxr-x--x
// while the octal value 0751 (489, b0001 1110 1001) has too many bits and is
// invalid!
func validateUnixPermissions(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)

	str := fmt.Sprintf("%03d", value)
	if len(str) != 3 {
		errors = append(errors, fmt.Errorf("%d does not appear to be a mode; expected: (0-7)(0-7)(0-7)", value))
	}

	for _, c := range str {
		if c < '0' || '7' < c {
			errors = append(errors, fmt.Errorf("%d does not appear to be a mode; %c should be (0-7)", value, c))
		}
	}

	return
}

func validateACLEntry(v interface{}, k string) (warnings []string, errors []error) {
	value := strings.Split(v.(string), ":")

	if len(value) != 3 {
		errors = append(errors, fmt.Errorf("%s: ACL entry must have three fields expected: (user|group|kind):[user-id]:(r|-)(w|-)(x|-)", k))
		return
	}

	kind := value[0]
	entity := value[1]
	perm := value[2]

	switch kind {
	case "user":
	case "group":
	case "other":
	default:
		errors = append(errors, fmt.Errorf("%s: unknown entity kind %s; expected user, group, other", k, value[0]))
	}

	if len(entity) == 0 {
		errors = append(errors, fmt.Errorf("%s: %s identifier is required in ACL entries; please use the mode field instead", k, kind))
	}

	ws, es := validateSymbolicPermissions(perm, k)
	warnings = append(warnings, ws...)
	errors = append(errors, es...)

	return
}

func modeToPerm(mode int32) (string, error) {
	m := mode
	perm := ""
	if m >= 4 {
		perm = perm + "r"
		m = m - 4
	} else {
		perm = perm + "-"
	}
	if m >= 2 {
		perm = perm + "w"
		m = m - 2
	} else {
		perm = perm + "-"
	}
	if m >= 1 {
		perm = perm + "x"
		m = m - 1
	} else {
		perm = perm + "-"
	}
	if m != 0 {
		return "", fmt.Errorf("Too many bits in %d, cannot convert to permissions", mode)
	}
	return perm, nil
}
