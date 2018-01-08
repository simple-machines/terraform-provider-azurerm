package azurerm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/datalake-store/filesystem"
)

func TestUnmarshalACL(t *testing.T) {

	var f filesystem.ACLStatus

	b := []byte(
		`{
		   "entries":["user:Microsoft.EventHubs:rwx", "group::rwx"],
		   "group":"bla",
		   "owner":"john",
		   "permission":"rwx",
		   "stickyBit":true
		 }`)

	json.Unmarshal(b, &f)

	fmt.Printf("the value of the json unmarshal is %s", f)

	if *(f.Owner) == "" ||
		*(f.Permission) == "" ||
		len(*(f.Entries)) == 0 ||
		*(f.Group) == "" ||
		*(f.StickyBit) == false {
		t.Error("ACL Status Unmarshal unsuccessful")
	}
}

func TestValidate_ACL(t *testing.T) {
	// Accept
	if _, es := validateACLEntry("user:1:rwx", ""); len(es) > 0 {
		t.Fatalf("Validation rejected user:1:rwx")
	}
	if _, es := validateACLEntry("group:2:rwx", ""); len(es) > 0 {
		t.Fatalf("Validation rejected group:2:rwx")
	}
	if _, es := validateACLEntry("other:asdln:rwx", ""); len(es) > 0 {
		t.Fatalf("Validation rejected other:asdln:rwx")
	}
	if _, es := validateACLEntry("user:42:rwx", ""); len(es) > 0 {
		t.Fatalf("Validation rejected user:42:rwx")
	}

	// Reject
	if _, es := validateACLEntry("none::rwx", ""); len(es) == 0 {
		t.Fatalf("Validation accepted none::rwx")
	}
	if _, es := validateACLEntry("user::c--", ""); len(es) == 0 {
		t.Fatalf("Validation accepted user::c--")
	}
	if _, es := validateACLEntry("extra:user::rwx", ""); len(es) == 0 {
		t.Fatalf("Validation accepted extra:user::rwx")
	}
	if _, es := validateACLEntry("extra:rwx", ""); len(es) == 0 {
		t.Fatalf("Validation accepted extra:rwx")
	}
	if _, es := validateACLEntry("user::rwx", ""); len(es) == 0 {
		t.Fatalf("Validation accepted user::rwx")
	}
}

func TestValidate_SymbolicPermissions(t *testing.T) {
	// Accept
	if _, es := validateSymbolicPermissions("---", ""); len(es) > 0 {
		t.Fatalf("Validation rejected ---")
	}
	if _, es := validateSymbolicPermissions("r--", ""); len(es) > 0 {
		t.Fatalf("Validation rejected r--")
	}
	if _, es := validateSymbolicPermissions("-w-", ""); len(es) > 0 {
		t.Fatalf("Validation rejected -w-")
	}
	if _, es := validateSymbolicPermissions("--x", ""); len(es) > 0 {
		t.Fatalf("Validation rejected --x")
	}
	if _, es := validateSymbolicPermissions("rwx", ""); len(es) > 0 {
		t.Fatalf("Validation rejected rwx")
	}

	// Reject
	if _, es := validateSymbolicPermissions("c--", ""); len(es) == 0 {
		t.Fatalf("Validation accepted c--")
	}
	if _, es := validateSymbolicPermissions("-", ""); len(es) == 0 {
		t.Fatalf("Validation accepted -")
	}
}

func TestValidate_UnixMode(t *testing.T) {
	// Accept
	if _, es := validateUnixPermissions(000, "000"); len(es) > 0 {
		t.Fatalf("Validation rejected 000")
	}
	if _, es := validateUnixPermissions(777, "777"); len(es) > 0 {
		t.Fatalf("Validation rejected 777")
	}
	if _, es := validateUnixPermissions(111, "111"); len(es) > 0 {
		t.Fatalf("Validation rejected 111")
	}

	// Reject
	if _, es := validateUnixPermissions(1000, "1000"); len(es) == 0 {
		t.Fatalf("Validation accepted 1000")
	}
	if _, es := validateUnixPermissions(-1, "-1"); len(es) == 0 {
		t.Fatalf("Validation accepted -1")
	}
}

func Test_modeToPerm(t *testing.T) {
	if v, _ := modeToPerm(7); v != "rwx" {
		t.Fatalf("modeToPerm(7) != rwx; %s", v)
	}
	if v, _ := modeToPerm(6); v != "rw-" {
		t.Fatalf("modeToPerm(6) != rw-; %s", v)
	}
	if v, _ := modeToPerm(5); v != "r-x" {
		t.Fatalf("modeToPerm(5) != r-x; %s", v)
	}
	if v, _ := modeToPerm(4); v != "r--" {
		t.Fatalf("modeToPerm(4) != r--; %s", v)
	}
	if v, _ := modeToPerm(3); v != "-wx" {
		t.Fatalf("modeToPerm(3) != -wx; %s", v)
	}
	if v, _ := modeToPerm(2); v != "-w-" {
		t.Fatalf("modeToPerm(2) != -w-; %s", v)
	}
	if v, _ := modeToPerm(1); v != "--x" {
		t.Fatalf("modeToPerm(1) != --x; %s", v)
	}
	if v, _ := modeToPerm(0); v != "---" {
		t.Fatalf("modeToPerm(0) != ---; %s", v)
	}
}
