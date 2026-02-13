package models

// User represents a Google Workspace user with fields needed for signature rendering.
type User struct {
	Email        string
	FirstName    string
	LastName     string
	JobTitle     string
	Organization string
	Phone        string
	OrgUnit      string
}

// PlaceholderData returns a map of placeholder keys to their values for template rendering.
func (u *User) PlaceholderData() map[string]string {
	return map[string]string{
		"email":        u.Email,
		"firstName":    u.FirstName,
		"lastName":     u.LastName,
		"jobTitle":     u.JobTitle,
		"organization": u.Organization,
		"phone":        u.Phone,
		"orgUnit":      u.OrgUnit,
	}
}
