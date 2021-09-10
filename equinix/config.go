package equinix

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/equinix/ecx-go/v2"
	"github.com/equinix/ne-go"
	"github.com/equinix/oauth2-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	xoauth2 "golang.org/x/oauth2"
)

//Config is the configuration structure used to instantiate the Equinix
//provider.
type Config struct {
	BaseURL        string
	ClientID       string
	ClientSecret   string
	Token          string
	RequestTimeout time.Duration
	PageSize       int

	ecx ecx.Client
	ne  ne.Client
}

//Load function validates configuration structure fields and configures
//all required API clients.
func (c *Config) Load(ctx context.Context) error {
	if c.BaseURL == "" {
		return fmt.Errorf("baseURL cannot be empty")
	}

	var authClient *http.Client

	if c.Token != "" {
		tokenSource := xoauth2.StaticTokenSource(&xoauth2.Token{AccessToken: c.Token})
		oauthTransport := &xoauth2.Transport{
			Source: tokenSource,
		}
		authClient = &http.Client{
			Transport: oauthTransport,
		}
	} else {
		if c.ClientID == "" {
			return fmt.Errorf("clientId cannot be empty")
		}
		if c.ClientSecret == "" {
			return fmt.Errorf("clientSecret cannot be empty")
		}
		authConfig := oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			BaseURL:      c.BaseURL,
		}
		authClient = authConfig.New(ctx)
	}
	authClient.Timeout = c.requestTimeout()
	authClient.Transport = logging.NewTransport("Equinix", authClient.Transport)
	ecxClient := ecx.NewClient(ctx, c.BaseURL, authClient)
	neClient := ne.NewClient(ctx, c.BaseURL, authClient)
	if c.PageSize > 0 {
		ecxClient.SetPageSize(c.PageSize)
		neClient.SetPageSize(c.PageSize)
	}
	c.ecx = ecxClient
	c.ne = neClient
	return nil
}

func (c *Config) requestTimeout() time.Duration {
	if c.RequestTimeout == 0 {
		return 5 * time.Second
	}
	return c.RequestTimeout
}
