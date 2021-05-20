package tesla

// List of base filenames for HTML saving
const (
	htmlLoginForm   = "login-form"
	htmlAuthorize   = "login-authorize"
	htmlMFAFactorID = "mfa-factorid"
	htmlMFAVerify   = "mfa-verify"
	htmlMFAToken    = "mfa-token"
	htmlOwnerToken  = "owner-token"
)

// MFA types we support
const (
	// This is using a software app like Google Authenticator to generate token strings
	factorTypeSoftwareToken = "token:software"
)

const (
	urlVoidCallback     = "https://auth.tesla.com/void/callback"
	urlVoidCallbackBase = "auth.tesla."
	urlVoidCallbackPath = "/void/callback"

	urlOwnerAPI = "https://owner-api.teslamotors.com/oauth/token"
)

const (
	teslaClientID = "81527cff06843c8634fdc09e8ac0abefb46ac849f38fe1e431c2ef2106796384"
	grantType     = "urn:ietf:params:oauth:grant-type:jwt-bearer"
)

// Required hidden input form fields
var requiredHiddenFields = []string{
	"_csrf",
	"transaction_id",
}
