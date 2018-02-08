package azurerm

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
	"time"
)

type identityProperty struct {
	identity          string
	seed              string
	increment         string
	notForReplication string
}

type columnProperty struct {
	name       string
	columnType string
	size       string
	null       string
	collation  string
}

type constraintProperty struct {
	constraintType string
	constraintName string
	// We don't care if its a list, Let the user pass comma separated string if there are multiple values
	constraintKeys string
}

type constraintProperties []constraintProperty

// TODO; add computed columns too
// SELECT * FROM sys.computed_columns WHERE object_id = OBJECT_ID('tablename')

type columnToSqlQuery map[string]string

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

			"constraint": {
				Type:        schema.TypeMap,
				Description: "Define the constraint name and the value as a sql string. Ex: constraint(). Currently only primary keys is supported",
				Optional:    true,
			},

			"resource_group_name": resourceGroupNameSchema(),

			"tablename": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"columns": {
				Type:        schema.TypeMap,
				Description: "Columns in a table with its corresponding sql statements",
				Required:    true,
			},
		},
	}
}

func resourceArmSqlTableCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Into the table creation logic")
	columns := d.Get("columns").(map[string]interface{})
	constraints := d.Get("constraints").(map[string]interface{})
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

	constraintSlices := make([]string, 0, len(constraints))

	for key, v := range constraints {
		value := v.(string)
		constraintSlices = append(constraintSlices, fmt.Sprintf("CONSTRAINT %s %s", key, value))
	}

	columnString := strings.Join(querySlices, ", ")

	if len(constraintSlices) > 0 {
		columnString += strings.Join(constraintSlices, ", ")
	}

	query := fmt.Sprintf("CREATE TABLE %s ( %s );", tablename, columnString)

	log.Printf("[INFO] The query built is %s", query)

	rows, err := conn.Query(query)

	defer closeRows(rows)

	log.Printf("%v", rows)

	if err != nil {
		return fmt.Errorf("Cannot run the query for some reason %v ", err.Error())
	}

	return resourceArmSqlTableRead(d, meta)
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

	resourceGroup := d.Get("resource_group_name").(string)

	log.Printf("resource group %v", resourceGroup)
	tablename := d.Get("tablename").(string)
	databases := d.Get("database").([]interface{})
	database := databases[0]
	config := databases[0].(map[string]interface{})
	name := config["name"].(string)
	server := config["server"].(string)

	username := config["username"].(string)
	password := config["password"].(string)

	dsn := "server=" + server + ";user id=" + username + ";password=" + password + ";database=" + name

	conn, err := sql.Open("mssql", dsn)

	err = conn.Ping()

	if err != nil {
		return fmt.Errorf("Unable to connect to the server/database: %v ", err.Error())
	}

	defer conn.Close()

	spHelpRows, err := conn.Query(fmt.Sprintf("sp_help %s", tablename))

	defer closeRows(spHelpRows)

	if err != nil {
		// If user creates through tf, and deletes manually. The fallback is to clear the tfstate as it is invalid now.
		return fmt.Errorf("Unable to query the description of the table %s: %v", tablename, err.Error())
	}

	err, defaultTableProperties := getTableProperties(spHelpRows)

	if err != nil {
		return fmt.Errorf("Unable to obtain the table properties from the table description: %s", err.Error())
	}

	if !spHelpRows.NextResultSet() {
		return fmt.Errorf("Could not read the column properties from the table %s", tablename)
	}

	err, columnProperties := getColumnProperties(spHelpRows)

	if err != nil {
		return fmt.Errorf("Unable to obtain the column properties from the table description: %s", err.Error())
	}

	if !spHelpRows.NextResultSet() {
		return fmt.Errorf("Could not read the identity properties from the table %s", tablename)
	}

	err, identityProperties := getIdentityProperties(spHelpRows)

	if err != nil {
		return fmt.Errorf("Unable to obtain the indentity properties from the table description: %s", err.Error())
	}

	if err != nil {
		return fmt.Errorf("Unable to obtain the details from the index properties: %s", err.Error())
	}

	spHelpConstraintsRows, err := conn.Query(fmt.Sprintf("sp_helpconstraint %s", tablename))

	defer closeRows(spHelpConstraintsRows)

	if err != nil {
		return fmt.Errorf("Unable to query the constraints of the table: %s", err.Error())
	}

	err, constraintProperties := getConstraintProperties(spHelpConstraintsRows)

	if err != nil {
		return fmt.Errorf("Unable to obtain the details from tha table constraints: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft.sql/servers/%s/databases/%s/tables/%s", subscriptionId, resourceGroup, server, database, tablename))

	output := make(map[string]string, 0)

	cQuery := columnToSqlQuery(output)

	cQuery.appendColumnProperties(columnProperties)
	cQuery.appendIdentityProperties(identityProperties)

	log.Printf("the constraint properties are %v", constraintProperties.toConstraintMap())

	d.Set("columns", cQuery)
	d.Set("constraints", constraintProperties.toConstraintMap())

	log.Printf("the default table properties is %v", defaultTableProperties)
	log.Printf("the column map is %v", d.Get("columns"))
	log.Printf("the constraint map is %v", d.Get("constraints"))

	return nil
}

func getTableProperties(rows *sql.Rows) (error, []interface{}) {
	log.Printf("Into getting the table properties")
	var name interface{}
	var owner interface{}
	var tableType interface{}
	var createdDateTime interface{}

	tableProperties := make(map[string]string, 0)

	for rows.Next() {
		err := rows.Scan(&name, &owner, &tableType, &createdDateTime)
		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		tableProperties["name"] = convertColumnValueToString(&name)
		tableProperties["owner"] = convertColumnValueToString(&owner)

	}

	return nil, []interface{}{tableProperties}

}

func getIdentityProperties(rows *sql.Rows) (error, identityProperty) {
	var identity interface{}
	var seed interface{}
	var increment interface{}
	var notForReplication interface{}

	for rows.Next() {
		err := rows.Scan(&identity, &seed, &increment, &notForReplication)

		if err != nil {
			return fmt.Errorf("Cannot scan for identity details of the table %v. This might be because of the conflicts in SQL version", err.Error()), identityProperty{}
		}

		// Better not to have multiple identity properties
		break
	}

	id := identityProperty{
		// Further abstraction on setting struct fields requires reflection.
		// Hence, lets be versbose and atomic in updating struct fields.
		convertColumnValueToString(&identity),
		convertColumnValueToString(&seed),
		convertColumnValueToString(&increment),
		convertColumnValueToString(&notForReplication),
	}

	return nil, id
}

func getConstraintProperties(rows *sql.Rows) (error, constraintProperties) {
	var objName interface{}

	for rows.Next() {

		err := rows.Scan(&objName)

		if err != nil {
			return fmt.Errorf("Cannot obtain the object name as part of scanning the constraint properties: %v", err.Error()), nil
		}
	}

	if !rows.NextResultSet() {
		return nil, nil
	}

	var constraintType interface{}
	var constraintName interface{}
	var deleteAction interface{}
	var updateAction interface{}
	var statusEnabled interface{}
	var statusForReplication interface{}
	var constraintKeys interface{}

	output := make([]constraintProperty, 0)
	//constraintProperties := make(map[string]string, 0)

	for rows.Next() {
		err := rows.Scan(&constraintType, &constraintName, &deleteAction, &updateAction, &statusEnabled, &statusForReplication, &constraintKeys)
		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		constraintProperties := constraintProperty{
			convertColumnValueToString(&constraintType),
			convertColumnValueToString(&constraintName),
			convertColumnValueToString(&constraintKeys),
		}

		output = append(output, constraintProperties)
	}

	return nil, constraintProperties(output)
}

func getColumnProperties(rows *sql.Rows) (error, []columnProperty) {
	cols, err := rows.Columns()

	log.Printf("[INFO] The metadata columns while fetching column metadata are %v", cols)

	if err != nil || cols == nil {
		return fmt.Errorf("Could not retrieve the metadata columns of the data %s", cols), nil
	}

	var columnName interface{}
	var columnType interface{}
	var computed interface{}
	var length interface{}
	var prec interface{}
	var scale interface{}
	var nullable interface{}
	var trimTrailingBlanks interface{}
	var fixedLengthNullInSource interface{}
	var collation interface{}

	output := make([]columnProperty, 0)

	for rows.Next() {
		err := rows.Scan(&columnName, &columnType, &computed, &length, &prec, &scale, &nullable, &trimTrailingBlanks, &fixedLengthNullInSource, &collation)

		if err != nil {
			return fmt.Errorf("Cannot scan for table details %v", err.Error()), nil
		}

		nullable := convertColumnValueToString(&nullable)

		columnProperties := columnProperty{
			convertColumnValueToString(&columnName),
			convertColumnValueToString(&columnType),
			convertColumnValueToString(&length),
			// To safely handle an inconsistent representation of nullable or not in sql server output
			func() string {
				if nullable == "no" {
					return "not null"
				} else {
					return "null"
				}
			}(),
			convertColumnValueToString(&collation),
		}

		output = append(output, columnProperties)
	}

	log.Printf("the output is %s", output)
	return nil, output
}

// A canonical representation of property values as strings makes sense to sql users.
// Ex: "Error: Collation is NULL" instead of "Error: Collation is `Nil`"
// Normalising to one type makes switch cases over sql values easier than type matching interface{}.
// The advantage of not having this method doesn't bring any alternative value adds than removal of few lines.
func convertColumnValueToString(pval *interface{}) string {
	switch v := (*pval).(type) {
	case nil:
		return "NULL"
	case bool:
		if v {
			return "1"
		} else {
			return "0"
		}
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05.999")
	default:
		return fmt.Sprint(v)
	}
}

func (c columnToSqlQuery) appendColumnProperties(columnProperties []columnProperty) columnToSqlQuery {
	for _, m := range columnProperties {
		if m.collation != "NULL" {
			c[m.name] = fmt.Sprintf("%s collate %s %s", m.columnType, m.collation, m.null)
		} else {
			c[m.name] = fmt.Sprintf("%s %s", m.columnType, m.null)
		}
	}

	return c
}

func (c columnToSqlQuery) appendIdentityProperties(identityProperty identityProperty) columnToSqlQuery {
	for k, v := range c {
		if k == identityProperty.identity {
			identity := "identity" + fmt.Sprintf("(%s, %s)", identityProperty.seed, identityProperty.increment)
			c[k] = fmt.Sprintf("%s %s", v, identity)

			// A sql table can have only one column with identity properties
			break
		}
	}

	return c
}

func (c constraintProperties) toConstraintMap() map[string]string {
	constraintsMap := make(map[string]string, 0)

	for _, constraint := range c {
		switch constraint.constraintName {
		case "PRIMARY KEY (clustered)":
			constraintsMap[constraint.constraintName] = fmt.Sprintf("primary key clustered (%s)", constraint.constraintKeys)
		case "RIMARY KEY (non-clustered)":
			constraintsMap[constraint.constraintName] = fmt.Sprintf("primary key nonclustered (%s)", constraint.constraintKeys)
		case "UNIQUE (non-clustered":
			constraintsMap[constraint.constraintName] = fmt.Sprintf("unique nonclustered (%s)", constraint.constraintKeys)
		case "UNIQUE (clustered)":
			constraintsMap[constraint.constraintName] = fmt.Sprintf("unique clustered (%s)", constraint.constraintKeys)
		default:
			continue
		}
	}

	return constraintsMap
}
