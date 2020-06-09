package models

const (
	AuthTypeDB     = "db"
	AuthTypeOAuth2 = "oauth2"
	AuthTypeLDAP   = "ldap"
	AuthTypeSso    = "sso"
)

// AuthModel holds information used to authenticate.
type AuthModel struct {
	Username string
	Password string
	// oauth2 name ex. google github
	OAuth2Name string
	OAuth2Code string
}
