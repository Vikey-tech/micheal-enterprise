package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"michealenterprise/internal/models"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

func Token() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func Email(s string) bool {
	a, e := mail.ParseAddress(s)
	return e == nil && a.Address == s && len(s) <= 254
}
func Password(s string) bool { return len(s) >= 12 && len(s) <= 72 }
func Can(u models.User, area string) bool {
	return u.Role == "owner" || (u.Role == "manager" && u.Permissions[area])
}

var uuid = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var image = regexp.MustCompile(`^/uploads/[0-9a-f]{64}\.webp$`)

func UUID(s string) bool  { return uuid.MatchString(s) }
func Image(s string) bool { return s == "" || image.MatchString(s) }
func Vehicle(v models.Vehicle) error {
	if len(strings.TrimSpace(v.Title)) < 3 || len(v.Title) > 150 || len(v.Make) < 1 || len(v.Make) > 60 || len(v.Model) < 1 || len(v.Model) > 60 || v.Year < 1900 || v.Year > time.Now().Year()+1 || v.Price < 1 || v.Price > 100000000000 || v.Mileage < 0 || v.Mileage > 2000000 || len(v.Description) > 10000 || len(v.Images) > 12 {
		return errors.New("check title, make, model, year, price, mileage and image limits")
	}
	if v.Status != "available" && v.Status != "reserved" && v.Status != "sold" {
		return errors.New("invalid vehicle status")
	}
	if v.Fuel != "Petrol" && v.Fuel != "Diesel" && v.Fuel != "Hybrid" && v.Fuel != "Electric" {
		return errors.New("invalid fuel")
	}
	if v.Transmission != "Automatic" && v.Transmission != "Manual" {
		return errors.New("invalid transmission")
	}
	for _, s := range v.Images {
		if s == "" || !Image(s) {
			return errors.New("use uploaded vehicle images")
		}
	}
	return nil
}
