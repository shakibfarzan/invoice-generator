package config

import "testing"

// sampleCompany is the expected result of parsing sampleConfig.
func sampleCompany() Company {
	return Company{
		Name:    "Acme Industries",
		Logo:    "./assets/logo.png",
		Phone:   "+98 21 1234 5678",
		Email:   "sales@acme.test",
		Address: "Tehran, Iran",
		Website: "https://acme.test",
		TaxID:   "1234567890",
	}
}

const sampleConfig = `
# seller configuration
company:
  name: "Acme Industries"
  logo: ./assets/logo.png
  phone: "+98 21 1234 5678"
  email: sales@acme.test
  address: Tehran, Iran
  website: https://acme.test
  tax_id: "1234567890"

# reserved for roadmap step 22; ignored for now
invoice:
  currency: IRR
  tax_rate: 10
`

func TestParseCompany(t *testing.T) {
	cfg, err := Parse(sampleConfig)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.Company != sampleCompany() {
		t.Errorf("company = %+v, want %+v", cfg.Company, sampleCompany())
	}
}

func TestParseMissingFileIsEmpty(t *testing.T) {
	cfg, err := Load("this-file-does-not-exist.yaml")
	if err != nil {
		t.Fatalf("Load of a missing file returned error: %v", err)
	}
	if cfg.Company != (Company{}) {
		t.Errorf("expected zero company for a missing file, got %+v", cfg.Company)
	}
}

func TestParseIgnoresUnknownSections(t *testing.T) {
	cfg, err := Parse("other:\n  foo: bar\ncompany:\n  name: X\n")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.Company.Name != "X" {
		t.Errorf("expected company name %q, got %q", "X", cfg.Company.Name)
	}
}

func TestParseToleratesInlineComments(t *testing.T) {
	cfg, err := Parse("company:\n  name: Acme # trailing comment\n  phone: \"+1 2 3\" # note\n")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.Company.Name != "Acme" {
		t.Errorf("expected name %q without comment, got %q", "Acme", cfg.Company.Name)
	}
	if cfg.Company.Phone != "+1 2 3" {
		t.Errorf("expected phone %q, got %q", "+1 2 3", cfg.Company.Phone)
	}
}

func TestParseRejectsBadLine(t *testing.T) {
	if _, err := Parse("company:\n  this line is malformed\n"); err == nil {
		t.Error("expected an error for a key without a colon")
	}
}
