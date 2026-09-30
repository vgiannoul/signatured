package models

// User represents a Google Workspace user with fields needed for signature rendering.
type User struct {
	Email           string
	FirstName       string
	LastName        string
	JobTitle        string
	Organization    string
	Phone           string
	PhoneLabel      string
	PhoneIsInternal bool
	OrgUnit         string
	PhoneMobile     string
	CompanyWebsite  string
	CompanyLogo     string
	CompanyPhone    string
	CompanyAddress  string
}

// PlaceholderData returns a map of placeholder keys to their values for template rendering.
func (u *User) PlaceholderData() map[string]string {
	// phoneIsInternal is a conditional-friendly flag: {{#if phoneIsInternal}}
	// is only truthy (non-empty) when the work phone is an internal
	// extension, so templates can prefix it with the company's main number:
	//   {{#if phoneIsInternal}}{{companyPhone}} / {{/if}}{{phoneLabel}} {{phone}}
	phoneIsInternal := ""
	if u.PhoneIsInternal {
		phoneIsInternal = "true"
	}

	return map[string]string{
		"email":           u.Email,
		"firstName":       u.FirstName,
		"lastName":        u.LastName,
		"jobTitle":        u.JobTitle,
		"organization":    u.Organization,
		"phone":           u.Phone,
		"phoneLabel":      u.PhoneLabel,
		"phoneIsInternal": phoneIsInternal,
		"orgUnit":         u.OrgUnit,
		"phoneMobile":     u.PhoneMobile,
		"companyWebsite":  u.CompanyWebsite,
		"companyLogo":     u.CompanyLogo,
		"companyPhone":    u.CompanyPhone,
		"companyAddress":  u.CompanyAddress,
	}
}
