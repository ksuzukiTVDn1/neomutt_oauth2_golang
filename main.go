package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log/syslog"
	"math/rand"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	// The URL root for accessing Google Accounts.
	GOOGLE_ACCOUNTS_BASE_URL = "https://accounts.google.com"
	GOOGLE_OAUTH_BASE_URL    = "https://oauth2.googleapis.com"

	// Hardcoded dummy redirect URI for non-web apps.
	// oob uri is abolition
	// REDIRECT_URI = "urn:ietf:wg:oauth:2.0:oob"
	REDIRECT_URI = "http://localhost"
)

// letters are used to generate a random value for the authorization code state
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString is generate random string of specified length
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// OAuth2Process is MuttOAuth's actual OAuth process
type OAuthProcess struct {
	ctx   context.Context
	oauth oauth2.Config
	token oauth2.Token
}

// GenerateOAuth2AuthorizationString generates OAuth2 authorization header string
// If plain is specified, it will be output in plain text.
// If not specified, the value encoded in base64 will be returned.
func (o OAuthProcess) GenerateOAuth2AuthorizationString(user string, plain bool) string {
	auth_string := fmt.Sprintf("user=%s\nauth=Bearer %s\n\n", user, o.token.AccessToken)
	if plain {
		return auth_string
	}
	return base64.StdEncoding.EncodeToString([]byte(auth_string))
}

// GenerateOAuth2AuthorizationToken is a function for performing OAuth authorization code flow,
// it implements authorization code flow in CLI and provides OAuth access token and refresh token acquisition function.
func (o OAuthProcess) GenerateOAuth2AuthorizationToken(quiet bool) error {
	state := RandomString(16)
	fmt.Println("To authorize token, visit this url and follow the directions:")
	fmt.Printf("  %s\n", o.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline))
	fmt.Println("Enter verification code:")
	var inputCode string
	fmt.Scan(&inputCode)

	ctx := context.Background()
	t, err := o.oauth.Exchange(ctx, inputCode)
	if err != nil {
		return err
	}

	PrintToken(t, quiet)
	return nil
}

// RefreshToken uses the refresh token to refresh the OAuth2 token
func (o *OAuthProcess) RefreshToken(quiet bool) error {
	ctx := context.Background()
	tokenSource := o.oauth.TokenSource(ctx, &o.token)
	t, err := tokenSource.Token()
	if err != nil {
		return err
	}
	PrintToken(t, quiet)
	return nil
}

// PrintToken prints the current OAuth token information
func PrintToken(token *oauth2.Token, quiet bool) {
	if quiet {
		fmt.Println(token.AccessToken)
	} else {
		fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
		fmt.Printf("Access Token: %s\n", token.AccessToken)
		fmt.Printf("Access Token Expiration Seconds: %s\n", token.Expiry)
	}
}

// MuttOAuth is structure that receives the startup parameters
type MuttOAuth struct {
	genOAuth2Token bool // Token generate switch
	genOAuth2Str   bool // Header generate switch
	testIMAPAuth   bool // test imap auth switch
	testSMTPAuth   bool // test smtp auth switch

	clientID     string // OAuth2 client ID
	clientSecret string // OAuth2 client secret
	accessToken  string // OAuth2 access token
	refreshToken string // OAuth2 refresh token
	user         string // IMAP account name
	quiet        bool   // quiet mode
}

func (mo MuttOAuth) Exec() {
	process := OAuthProcess{
		ctx: context.Background(),
		oauth: oauth2.Config{
			ClientID:     mo.clientID,
			ClientSecret: mo.clientSecret,
			Scopes:       []string{gmail.MailGoogleComScope},
			Endpoint:     google.Endpoint,
			RedirectURL:  REDIRECT_URI,
		},
		token: oauth2.Token{
			AccessToken:  mo.accessToken,
			RefreshToken: mo.refreshToken,
		},
	}

	if mo.refreshToken != "" {
		err := process.RefreshToken(mo.quiet)
		if err != nil {
			syslog.New(syslog.LOG_DEBUG, err.Error())
			panic(1)
		}
		return
	}

	if mo.genOAuth2Str {
		tmp := process.GenerateOAuth2AuthorizationString(mo.user, false)
		fmt.Printf("%s", tmp)
		return
	}

	if mo.genOAuth2Token {
		err := process.GenerateOAuth2AuthorizationToken(mo.quiet)
		if err != nil {
			syslog.New(syslog.LOG_DEBUG, err.Error())
			panic(1)
		}
		return
	}

	if mo.testIMAPAuth && mo.testSMTPAuth {
		tmp := process.GenerateOAuth2AuthorizationString(mo.user, true)
		fmt.Printf("%s", tmp)
		return
	}

	fmt.Println("Nothing to do.")
}

func main() {
	var (
		genOAuth2Token = flag.Bool("generate_oauth2_token", false, "")
		genOAuth2Str   = flag.Bool("generate_oauth2_string", false, "")

		clientID     = flag.String("client_id", "", "Client ID of the application that is authenticating. See OAuth2 documentation for details.")
		clientSecret = flag.String("client_secret", "", "Client secret of the application that is authenticating. See OAuth2 documentation for details.")
		accessToken  = flag.String("access_token", "", "OAuth2 access token")
		refreshToken = flag.String("refresh_token", "", "OAuth2 refresh token")
		testIMAPAuth = flag.Bool("test_imap_authentication", false, "")
		testSMTPAuth = flag.Bool("test_smtp_authentication", false, "")
		user         = flag.String("user", "None", "'email address of user whose account is being accessed")
		quiet        = flag.Bool("quiet", false, "Omit verbose descriptions and only print machine-readable outputs.")
	)
	flag.Parse()

	o := MuttOAuth{
		genOAuth2Token: *genOAuth2Token,
		genOAuth2Str:   *genOAuth2Str,
		clientID:       *clientID,
		clientSecret:   *clientSecret,
		accessToken:    *accessToken,
		refreshToken:   *refreshToken,
		testIMAPAuth:   *testIMAPAuth,
		testSMTPAuth:   *testSMTPAuth,
		user:           *user,
		quiet:          *quiet,
	}
	o.Exec()
}
