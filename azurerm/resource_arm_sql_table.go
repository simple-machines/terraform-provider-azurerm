package azurerm

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceArmSqlTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSqlTableCreate,
		Read:   resourceArmSqlTableRead,
		Update: resourceArmSqlTableCreate,
		Delete: resourceArmSqlTableCreate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			//
			// Mandatory fields
			//
			"database": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    1,
				Description: "The database details on which the table has to be created",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"server": {
							Type:     schema.TypeString,
							Required: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"resource_group_name": resourceGroupNameSchema(),

			"tablename": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// This will go off, as we are moving away from string comparison.
			"columns": {
				Type:        schema.TypeMap,
				Description: "Columns in a table with its corresponding sql statements",
				Required:    true,
			},

			// Another possible solution is to run the script and get the create table query.
			// The advantage with it would be more flexible queries.
			// https://stackoverflow.com/questions/706664/generate-sql-create-scripts-for-existing-tables-with-query
			// http://www.c-sharpcorner.com/UploadFile/67b45a/how-to-generate-a-create-table-script-for-an-existing-table/
			// However, let us don't prefer this as constraints with robustness is better than fragile flexibility from a user perspective, and technically.
			"table_description": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"main_properties": {
							Type:     schema.TypeList,
							MinItems: 1,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"owner": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"identity_properties": {
							Type:     schema.TypeList,
							MinItems: 1,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"identity": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"seed": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"increment": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"not_for_replication": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"index_properties": {
							Type:     schema.TypeList,
							MinItems: 1,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"index_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"index_description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"index_keys": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"constraint_properties": {
							Type:     schema.TypeList,
							MinItems: 1,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"constraint_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"constraint_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"constraint_keys": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"column_properties": {
							Type:     schema.TypeList,
							MinItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"null": {
										// NULL or NOT NULL
										Type:     schema.TypeString,
										Computed: true,
									},
									"collation": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceArmSqlTableCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Into the table creation logic")
	columns := d.Get("columns").(map[string]interface{})
	tablename := d.Get("tablename").(string)
	databases := d.Get("database").([]interface{})
	config := databases[0].(map[string]interface{})
	name := config["name"].(string)
	server := config["server"].(string)

	log.Printf("[INFO] Ther server name is %v", server)

	username := config["username"].(string)
	password := config["password"].(string)
	dsn := "server=" + server + ";user id=" + username + ";password=" + password + ";database=" + name
	log.Printf("[INFO] The connection string is %s", dsn)
	conn, err := sql.Open("mssql", dsn)

	if err != nil {
		return fmt.Errorf("Cannot connect: %v", err.Error())
	}

	err = conn.Ping()

	if err != nil {
		return fmt.Errorf("Cannot connect: %v ", err.Error())
	}

	defer conn.Close()

	querySlices := make([]string, 0, len(columns))

	log.Printf("[INFO] Building the query string using the slices %v", querySlices)

	for key, v := range columns {
		value := v.(string)
		querySlices = append(querySlices, strings.Join([]string{key, value}, " "))
	}

	query := fmt.Sprintf("CREATE TABLE %s ( %s );", tablename, strings.Join(querySlices, ","))

	log.Printf("[INFO] The query built is %s", query)

	rows, err := conn.Query(query)

	defer closeRows(rows)

	log.Printf("%v", rows)

	if err != nil {
		return fmt.Errorf("Cannot run the query for some reason %v ", err.Error())
	}

	return nil
}

// Go made me do this...
func closeRows(r *sql.Rows) error {
	var err error

	if r != nil {
		err = r.Close()
	}

	return err
}

func resourceArmSqlTableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	subscriptionId := client.subscriptionId

	resourceGroup := d.Get("resource_group_namne").(string)
	tablename := d.Get("tablename").(string)
	databases := d.Get("database").([]interface{})
	database := databases[0]
	config := databases[0].(map[string]interface{})
	name := config["name"].(string)
	server := config["server"].(string)

	log.Printf("[INFO] Ther server name is %v", server)

	username := config["username"].(string)
	password := config["password"].(string)

	dsn := "server=" + server + ";user id=" + username + ";password=" + password + ";database=" + name
	conn, err := sql.Open("mssql", dsn)

	err = conn.Ping()

	if err != nil {
		return fmt.Errorf("Cannot connect: %v ", err.Error())
	}

	err = conn.Ping()

	if err != nil {
		return fmt.Errorf("Cannot connect: %v ", err.Error())
	}

	defer conn.Close()

	rows, err := conn.Query(fmt.Sprintf("sp_help %s", tablename))
	if err != nil {
		return fmt.Errorf("Cannot read the of the description of the table %s in database %s", tablename, database)
	}

	_, defaultTableProperties := getTableProperties(rows)
	_, columnProperties := getColumnProperties(rows, tablename)
	_, identityProperties := getIdentityProperties(rows, tablename)

	// Skip this few result set for the time being
	rows.NextResultSet()

	// Skip this result set for the time being
	rows.NextResultSet()

	_, indexProperties := getIndexProperties(rows, tablename)
	_, constraintProperties := getConstraintProperties(rows, tablename)

	value := map[string]interface{}{}

	value["main_properties"] = defaultTableProperties
	value["column_properties"] = columnProperties
	value["identity_properties"] = identityProperties
	value["constaint_properties"] = constraintProperties
	value["index properties"] = indexProperties

	d.SetId(fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft.sql/servers/%s/databases/%s/tables/%s", subscriptionId, resourceGroup, server, database, tablename))
	d.Set("table_description", value)

	defer closeRows(rows)

	return nil
}

func getTableProperties(rows *sql.Rows) (error, []interface{}) {
	var name string
	var owner string
	var tableType string
	var createdDateTime string

	output := make([]interface{}, 1)

	for rows.Next() {
		err := rows.Scan(&name, &owner, &tableType, &createdDateTime)
		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		tableProperties := make(map[string]string, 0)
		tableProperties["name"] = name
		tableProperties["owner"] = owner

		output = append(output, tableProperties)

		return nil, output

	}

	return nil, nil
}

func getIdentityProperties(rows *sql.Rows, tableName string) (error, []interface{}) {
	if !rows.NextResultSet() {
		return fmt.Errorf("Could not read the identity properties from the table %s", tableName), nil
	}

	var identity string
	var seed string
	var increment string
	var notForReplication string

	output := make([]interface{}, 1)

	for rows.Next() {
		err := rows.Scan(&identity, &seed, &increment, &notForReplication)
		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		identityProperties := make(map[string]string, 0)
		identityProperties["identity"] = identity
		identityProperties["seed"] = seed
		identityProperties["increment"] = increment
		identityProperties["not_for_replication"] = notForReplication

		output = append(output, identityProperties)

		return nil, output

	}

	return nil, nil
}

func getIndexProperties(rows *sql.Rows, tableName string) (error, []interface{}) {
	if !rows.NextResultSet() {
		return fmt.Errorf("Could not read the identity properties from the table %s", tableName), nil
	}

	var indexName string
	var indexDescription string
	var indexKeys string

	output := make([]interface{}, 1)

	for rows.Next() {
		err := rows.Scan(&indexName, &indexDescription, &indexKeys)
		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		indexProperties := make(map[string]string, 0)
		indexProperties["index_name"] = indexName
		indexProperties["index_description"] = indexDescription
		indexProperties["index_keys"] = indexKeys

		output = append(output, indexProperties)

		return nil, output

	}

	return nil, nil
}

func getConstraintProperties(rows *sql.Rows, tableName string) (error, []interface{}) {
	if !rows.NextResultSet() {
		return fmt.Errorf("Could not read the identity properties from the table %s", tableName), nil
	}

	var constraintType string
	var constraintName string
	var constraintKeys string

	output := make([]interface{}, 1)

	for rows.Next() {
		err := rows.Scan(&constraintType, &constraintName, &constraintKeys)
		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		constraintProperties := make(map[string]string, 0)
		constraintProperties["constraint_type"] = constraintType
		constraintProperties["constraint_name"] = constraintName
		constraintProperties["constraint_keys"] = constraintKeys

		output = append(output, constraintProperties)

		return nil, output

	}

	return nil, nil
}

func getColumnProperties(rows *sql.Rows, tableName string) (error, []interface{}) {
	if !rows.NextResultSet() {
		return fmt.Errorf("Could not read the column properties from the table %s", tableName), nil
	}

	cols, err := rows.Columns()

	log.Printf("[INFO] The metadata columns while fetching column metadata are %v", cols)

	if err != nil || cols == nil {
		return fmt.Errorf("Could not retrieve the metadata columns of the data %s", cols), nil
	}

	vals := make([]interface{}, len(cols))

	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
	}

	output := make([]interface{}, 1)

	for rows.Next() {
		err := rows.Scan(vals...)
		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		columnProperties := make(map[string]string, 0)
		// A safer approach when compared to the above approach.
		for i := 0; i < len(vals); i++ {
			switch cols[i] {
			case "column name":
				columnProperties["name"] = vals[i].(string)
			case "type":
				columnProperties["type"] = vals[i].(string)
			case "size":
				columnProperties["size"] = vals[i].(string)
			case "nullable":
				if vals[i] == "no" {
					columnProperties["null"] = "NULL"
				} else {
					columnProperties["null"] = "NOT NULL"
				}

			case "collation":
				columnProperties["collation"] = vals[i].(string)
			default:
				continue

			}
		}

		output = append(output, columnProperties)
	}

	return nil, output
}
