package saboteur

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"golang.org/x/oauth2"
)

// AuthHeaderSource returns the value of the Authorization header for an HTTP request.
type AuthHeaderSource func(ctx context.Context) (string, error)

func SetupAuth(auth Auth) (apiTransport http.RoundTripper, gitAuthHeaderSource AuthHeaderSource, err error) {
	switch auth := auth.(type) {
	case *AuthPAT:
		var token string

		if auth.TokenFromEnv != "" {
			token = os.Getenv(auth.TokenFromEnv)
		}

		if token == "" {
			token = auth.Token
		}

		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)

		apiTransport = &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, tokenSource)}
		gitAuthHeaderSource = func(context.Context) (string, error) {
			return "Basic " + basicAuth(auth.Username, token), nil
		}
		err = nil
	case *AuthInstallation:
		transport, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, auth.AppID, auth.InstallationID, auth.KeyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("error initializing app credentials: %w", err)
		}

		apiTransport = transport
		gitAuthHeaderSource = func(ctx context.Context) (string, error) {
			token, err := transport.Token(ctx)
			if err != nil {
				return "", err
			}

			return "Basic " + basicAuth("x-access-token", token), nil
		}
		err = nil
	default:
		return nil, nil, fmt.Errorf("unexpected auth type %T", auth)
	}

	return
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
