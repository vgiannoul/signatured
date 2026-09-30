package google

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/vgiannoul/signatured/internal/models"
	directory "google.golang.org/api/admin/directory/v1"
)

// orgUnitPathPattern matches valid Google Workspace org unit paths
// (letters, digits, spaces, and /_.- separators). Rejecting anything
// else prevents breaking out of the quoted Directory API query below.
var orgUnitPathPattern = regexp.MustCompile(`^[A-Za-z0-9 _./-]+$`)

// CompanyConfig holds company-wide configuration for all users.
type CompanyConfig struct {
	Website string
	Logo    string
	Phone   string
	Address string

	// PhoneLabel is shown next to a user's work phone number.
	// InternalPhoneLabel is used instead when that phone is marked with a
	// "custom" type whose customType contains "internal" in Directory
	// (e.g. an internal extension rather than an externally-dialable number).
	// Both are empty by default; a template must use {{phoneLabel}} for
	// either to have any effect.
	PhoneLabel         string
	InternalPhoneLabel string
}

// DirectoryClient handles fetching user data from Google Workspace Directory API.
type DirectoryClient struct {
	service       *directory.Service
	domain        string
	companyConfig CompanyConfig
}

// NewDirectoryClient creates a new Directory API client.
func NewDirectoryClient(service *directory.Service, domain string, companyConfig CompanyConfig) *DirectoryClient {
	return &DirectoryClient{
		service:       service,
		domain:        domain,
		companyConfig: companyConfig,
	}
}

// GetUser fetches a single user by email address.
func (d *DirectoryClient) GetUser(ctx context.Context, email string) (*models.User, error) {
	user, err := d.service.Users.Get(email).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %w", email, err)
	}

	return d.convertUser(user), nil
}

// ListUsers fetches all users in the domain.
func (d *DirectoryClient) ListUsers(ctx context.Context) ([]*models.User, error) {
	return d.listUsersWithQuery(ctx, "")
}

// ListUsersByOrgUnit fetches all users in a specific organizational unit.
func (d *DirectoryClient) ListUsersByOrgUnit(ctx context.Context, orgUnitPath string) ([]*models.User, error) {
	if !orgUnitPathPattern.MatchString(orgUnitPath) {
		return nil, fmt.Errorf("invalid org unit path %q: must contain only letters, digits, spaces, and /_.- characters", orgUnitPath)
	}
	query := fmt.Sprintf("orgUnitPath='%s'", orgUnitPath)
	return d.listUsersWithQuery(ctx, query)
}

// listUsersWithQuery fetches users matching the given query with pagination.
func (d *DirectoryClient) listUsersWithQuery(ctx context.Context, query string) ([]*models.User, error) {
	var users []*models.User
	pageToken := ""

	for {
		call := d.service.Users.List().
			Customer("my_customer").
			MaxResults(500). // Maximum allowed by API
			Context(ctx)

		if query != "" {
			call = call.Query(query)
		}

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		for _, u := range resp.Users {
			users = append(users, d.convertUser(u))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return users, nil
}

// Phone represents a user's phone number from the Directory API.
type Phone struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Organization represents a user's organization from the Directory API.
type Organization struct {
	Primary bool   `json:"primary"`
	Name    string `json:"name"`
	Title   string `json:"title"`
}

// isInternalPhone reports whether a Directory phone entry is marked as an
// internal number - a "custom" type whose customType mentions "internal"
// (e.g. an office extension rather than an externally-dialable number).
func isInternalPhone(phoneMap map[string]interface{}) bool {
	customType, _ := phoneMap["customType"].(string)
	return strings.Contains(strings.ToLower(customType), "internal")
}

// convertUser converts a Directory API user to our internal User model.
func (d *DirectoryClient) convertUser(u *directory.User) *models.User {
	user := &models.User{
		Email:          u.PrimaryEmail,
		OrgUnit:        u.OrgUnitPath,
		CompanyWebsite: d.companyConfig.Website,
		CompanyLogo:    d.companyConfig.Logo,
		CompanyPhone:   d.companyConfig.Phone,
		CompanyAddress: d.companyConfig.Address,
	}

	// Extract name fields
	if u.Name != nil {
		user.FirstName = u.Name.GivenName
		user.LastName = u.Name.FamilyName
	}

	// Extract phone numbers (work and mobile)
	// Since Phones is interface{}, we need to handle it carefully
	if u.Phones != nil {
		if phonesData, ok := u.Phones.([]interface{}); ok && len(phonesData) > 0 {
			var phoneIsInternal bool

			for _, p := range phonesData {
				phoneMap, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				phoneType, _ := phoneMap["type"].(string)
				phoneValue, _ := phoneMap["value"].(string)
				if phoneValue == "" {
					continue
				}

				switch {
				case phoneType == "work" && user.Phone == "":
					user.Phone = phoneValue
					phoneIsInternal = false
				case phoneType == "mobile" && user.PhoneMobile == "":
					user.PhoneMobile = phoneValue
				case phoneType == "custom" && user.Phone == "" && isInternalPhone(phoneMap):
					user.Phone = phoneValue
					phoneIsInternal = true
				}
			}

			// Fallback to the first non-mobile phone if no work/internal phone
			// was found. Mobile numbers are excluded here since they're already
			// represented via PhoneMobile - falling back to one would show the
			// same number twice in the signature (as both the work and mobile
			// phone).
			if user.Phone == "" {
				for _, p := range phonesData {
					phoneMap, ok := p.(map[string]interface{})
					if !ok {
						continue
					}
					phoneType, _ := phoneMap["type"].(string)
					phoneValue, _ := phoneMap["value"].(string)
					if phoneValue == "" || phoneType == "mobile" {
						continue
					}
					user.Phone = phoneValue
					phoneIsInternal = isInternalPhone(phoneMap)
					break
				}
			}

			if user.Phone != "" {
				user.PhoneIsInternal = phoneIsInternal
				if phoneIsInternal {
					user.PhoneLabel = d.companyConfig.InternalPhoneLabel
				} else {
					user.PhoneLabel = d.companyConfig.PhoneLabel
				}
			}
		}
	}

	// Extract organization details
	// Since Organizations is interface{}, we need to handle it carefully
	if u.Organizations != nil {
		if orgsData, ok := u.Organizations.([]interface{}); ok && len(orgsData) > 0 {
			// Try to find primary organization
			for _, o := range orgsData {
				if orgMap, ok := o.(map[string]interface{}); ok {
					isPrimary, _ := orgMap["primary"].(bool)
					if isPrimary {
						user.JobTitle, _ = orgMap["title"].(string)
						user.Organization, _ = orgMap["name"].(string)
						break
					}
				}
			}
			// Fallback to first organization if no primary
			if user.JobTitle == "" && user.Organization == "" {
				if orgMap, ok := orgsData[0].(map[string]interface{}); ok {
					user.JobTitle, _ = orgMap["title"].(string)
					user.Organization, _ = orgMap["name"].(string)
				}
			}
		}
	}

	return user
}
