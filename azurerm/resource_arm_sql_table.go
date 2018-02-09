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

// columnToSqlQuery is a receiver to custom functions such as append columnProperties, identityProperties etc.
type columnsToSqlStatements map[string]string

func resourceArmSqlTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSqlTableCreate,
		Read:   resourceArmSqlTableRead,
		Update: resourceArmSqlUpdate,
		Delete: resourceArmSqlDelete,
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

			"constraints": {
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
	columns := d.Get("columns").(map[string]interface{})
	constraints := d.Get("constraints").(map[string]interface{})
	tablename := d.Get("tablename").(string)

	conn, err := getConnection(d)

	defer conn.Close()

	if err != nil {
		return err
	}

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

func resourceArmSqlUpdate(d *schema.ResourceData, meta interface{}) error {
	tablename := d.Get("tablename").(string)

	conn, err := getConnection(d)

	defer conn.Close()

	if err != nil {
		return nil
	}

	if d.HasChange("columns") {
		prev, newValue := d.GetChange("columns")

		for newKey, newValue := range newValue.(map[string]interface{}) {
			oldValue := prev.(map[string]interface{})[newKey]

			if oldValue == nil {
				rows, err := conn.Query(fmt.Sprintf("ALTER TABLE %s ADD %s %s", tablename, newKey, newValue.(string)))
				closeRows(rows)

				if err != nil {
					return fmt.Errorf("Cannot add a new column into the table: %v", err.Error())
				}

			} else if newValue != oldValue {
				rows, err := conn.Query(fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s", tablename, newKey, newValue.(string)))
				closeRows(rows)

				if err != nil {
					return fmt.Errorf("Cannot alter an existing column in the table: %v", err.Error())
				}
			}
		}

	}

	return resourceArmSqlTableRead(d, meta)

}

func resourceArmSqlDelete(d *schema.ResourceData, meta interface{}) error {
	tablename := d.Get("tablename").(string)

	conn, err := getConnection(d)

	defer conn.Close()

	if err != nil {
		return nil
	}

	rows, err := conn.Query(fmt.Sprintf("DROP TABLE %s", tablename))

	if err != nil {
		return fmt.Errorf("Unable to delete the table resource from the database: %v", err.Error())
	}

	log.Printf("Succesfully Deleted the table resource %s", tablename)

	defer closeRows(rows)

	return nil
}

func resourceArmSqlTableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	subscriptionId := client.subscriptionId
	databases := d.Get("database").([]interface{})
	config := databases[0].(map[string]interface{})
	database := config["name"].(string)
	server := config["server"].(string)
	resourceGroup := d.Get("resource_group_name").(string)
	tablename := d.Get("tablename").(string)

	conn, err := getConnection(d)

	defer conn.Close()

	if err != nil {
		return err
	}

	//
	// Fetch table properties: Table name, owner, column descriptions and identity
	//
	spHelpRows, err := conn.Query(fmt.Sprintf("sp_help %s", tablename))

	defer closeRows(spHelpRows)

	if err != nil {
		return fmt.Errorf("Unable to query table description: %v", err.Error())
	}

	err, defaultTableProperties := getTablePropertiesFromRows(spHelpRows)

	if err != nil {
		return err
	}

	if !spHelpRows.NextResultSet() {
		return fmt.Errorf("No additional result set available in sp_help to scan the column properties of table %s", tablename)
	}

	err, columnProperties := getColumnPropertiesFromRows(spHelpRows)

	if err != nil {
		return err
	}

	if !spHelpRows.NextResultSet() {
		return fmt.Errorf("No additional result set available in sp_help to scan the identity properties of table %s", tablename)
	}

	err, identityProperties := getIdentityPropertiesFromRows(spHelpRows)

	if err != nil {
		return fmt.Errorf("Unable to obtain the indentity properties from the table description: %s", err.Error())
	}

	//
	// Fetch table constraints
	//
	spHelpConstraintsRows, err := conn.Query(fmt.Sprintf("sp_helpconstraint %s", tablename))

	defer closeRows(spHelpConstraintsRows)

	if err != nil {
		return fmt.Errorf("Unable to query the table constraints: %s", err.Error())
	}

	err, constraintProperties := getConstraintPropertiesFromRows(spHelpConstraintsRows)

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft.sql/servers/%s/databases/%s/tables/%s", subscriptionId, resourceGroup, server, database, tablename))

	// Build the map with column names and corresponding sql statements.
	columnsToSqlStatements := columnsToSqlStatements(make(map[string]string, 0))

	columnsToSqlStatements.appendColumnProperties(columnProperties).appendIdentityProperties(identityProperties)

	log.Printf("the constraint properties are %v", constraintProperties.toConstraintMap())

	d.Set("columns", columnsToSqlStatements)
	d.Set("constraints", constraintProperties.toConstraintMap())

	log.Printf("the default table properties is %v", defaultTableProperties)
	log.Printf("the column map is %v", d.Get("columns"))
	log.Printf("the constraint map is %v", d.Get("constraints"))

	return nil
}

func getTablePropertiesFromRows(rows *sql.Rows) (error, []interface{}) {
	var name interface{}
	var owner interface{}
	var tableType interface{}
	var createdDateTime interface{}

	tableProperties := make(map[string]string, 0)

	for rows.Next() {
		err := rows.Scan(&name, &owner, &tableType, &createdDateTime)
		if err != nil {
			return fmt.Errorf("Failed to scan/parse the column properties of the table: %v", err.Error()), nil
		}

		tableProperties["name"] = convertColumnValueToString(&name)
		tableProperties["owner"] = convertColumnValueToString(&owner)

	}

	return nil, []interface{}{tableProperties}

}

func getIdentityPropertiesFromRows(rows *sql.Rows) (error, identityProperty) {
	var identity interface{}
	var seed interface{}
	var increment interface{}
	var notForReplication interface{}

	for rows.Next() {
		err := rows.Scan(&identity, &seed, &increment, &notForReplication)

		if err != nil {
			return fmt.Errorf("Failed to scan/parse the column properties of the table: %v", err.Error()), identityProperty{}
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

func getConstraintPropertiesFromRows(rows *sql.Rows) (error, constraintProperties) {
	var objName interface{}

	for rows.Next() {

		err := rows.Scan(&objName)

		if err != nil {
			return fmt.Errorf("Failed to scan/parse the constraint properties of the table: %v", err.Error()), nil
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

func getColumnPropertiesFromRows(rows *sql.Rows) (error, []columnProperty) {
	cols, err := rows.Columns()
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
			return fmt.Errorf("Failed to scan/parse the column properties of the table: %v", err.Error()), nil
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

func (c columnsToSqlStatements) appendColumnProperties(columnProperties []columnProperty) columnsToSqlStatements {
	for _, m := range columnProperties {
		columnType := func() string {
			if m.columnType == "varchar" || m.columnType == "nvarchar" || m.columnType == "char" {
				return fmt.Sprintf("%s(%s)", m.columnType, m.size)
			} else {
				return m.columnType
			}
		}()

		if m.collation != "NULL" {
			c[m.name] = fmt.Sprintf("%s collate %s %s", columnType, m.collation, m.null)
		} else {
			c[m.name] = fmt.Sprintf("%s %s", m.columnType, m.null)
		}
	}

	return c
}

func (c columnsToSqlStatements) appendIdentityProperties(identityProperty identityProperty) columnsToSqlStatements {
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
		case "PRIMARY KEY (non-clustered)":
			constraintsMap[constraint.constraintName] = fmt.Sprintf("primary key nonclustered (%s)", constraint.constraintKeys)
		case "UNIQUE (non-clustered)":
			constraintsMap[constraint.constraintName] = fmt.Sprintf("unique nonclustered (%s)", constraint.constraintKeys)
		case "UNIQUE (clustered)":
			constraintsMap[constraint.constraintName] = fmt.Sprintf("unique clustered (%s)", constraint.constraintKeys)
		default:
			continue
		}
	}

	return constraintsMap
}

func getConnection(d *schema.ResourceData) (*sql.DB, error) {
	databases := d.Get("database").([]interface{})
	config := databases[0].(map[string]interface{})
	name := config["name"].(string)
	server := config["server"].(string)

	username := config["username"].(string)
	password := config["password"].(string)
	dsn := "server=" + server + ";user id=" + username + ";password=" + password + ";database=" + name
	conn, err := sql.Open("mssql", dsn)

	err = conn.Ping()

	if err != nil {
		return nil, fmt.Errorf("Unable to connect to the server: %v ", err.Error())
	}

	return conn, nil
}

// Go made me do this...
func closeRows(r *sql.Rows) error {
	var err error

	if r != nil {
		err = r.Close()
	}

	return err
}
