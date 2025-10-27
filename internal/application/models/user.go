package models

// User represents the authenticated user from Authentik OIDC
type User struct {
	Sub               string `json:"sub"` // Authentik user ID
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Nickname          string `json:"nickname"`
}

// GetID returns the user's unique identifier from Authentik
func (u *User) GetID() string {
	return u.Sub
}

// GetUsername returns the preferred username
func (u *User) GetUsername() string {
	if u.PreferredUsername != "" {
		return u.PreferredUsername
	}
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Email
}
