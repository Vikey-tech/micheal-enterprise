package service

import (
	"michealenterprise/internal/models"
	"testing"
)

func TestRoles(t *testing.T) {
	for _, tc := range []struct {
		u    models.User
		area string
		want bool
	}{
		{models.User{Role: "customer", Permissions: map[string]bool{"vehicles": true}}, "vehicles", false},
		{models.User{Role: "manager", Permissions: map[string]bool{"vehicles": true}}, "vehicles", true},
		{models.User{Role: "manager", Permissions: map[string]bool{"vehicles": true}}, "owner-only", false},
		{models.User{Role: "manager", Permissions: map[string]bool{}}, "chat", false},
		{models.User{Role: "owner"}, "owner-only", true},
	} {
		if Can(tc.u, tc.area) != tc.want {
			t.Errorf("role %s area %s", tc.u.Role, tc.area)
		}
	}
}
func TestVehicleValidation(t *testing.T) {
	v := models.Vehicle{Title: "Toyota Camry", Make: "Toyota", Model: "Camry", Year: 2022, Price: 25000000, Fuel: "Petrol", Transmission: "Automatic", Status: "available", Images: []string{"/uploads/" + Token() + ".webp"}}
	if e := Vehicle(v); e != nil {
		t.Fatal(e)
	}
	v.Price = -1
	if Vehicle(v) == nil {
		t.Fatal("negative price accepted")
	}
	v.Price = 100
	v.Images = []string{"https://attacker.test/a.svg"}
	if Vehicle(v) == nil {
		t.Fatal("external image accepted")
	}
}
func TestIdentifiers(t *testing.T) {
	if UUID("not-a-uuid") || Image("/uploads/../../secret") || Image("/uploads/x.svg") {
		t.Fatal("unsafe identifier accepted")
	}
	a, b := Token(), Token()
	if len(a) != 64 || a == b {
		t.Fatal("bad tokens")
	}
	if !Email("person@example.com") || Email("Name <person@example.com>") || Password("short") {
		t.Fatal("bad validation")
	}
}
