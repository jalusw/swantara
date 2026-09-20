package contacts

import "testing"

func TestContactFixtures(t *testing.T) {
	if ContactFixture() == nil {
		t.Error("contact = nil")
	}
	if ContactAddressFixture() == nil {
		t.Error("address = nil")
	}
	if ContactBankAccountFixture() == nil {
		t.Error("bank = nil")
	}
	if CustomerProfileFixture() == nil {
		t.Error("customer = nil")
	}
	if SupplierProfileFixture() == nil {
		t.Error("supplier = nil")
	}
	if ContactFixture(func(p *Contact) *Contact { return p }) == nil {
		t.Error("contact opt = nil")
	}
}
