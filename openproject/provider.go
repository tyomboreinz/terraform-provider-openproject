package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ProviderConfig struct {
	AppURL string
	APIKey string
}

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"app_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "OpenProject API base URL",
			},
			"apikey": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "API key for OpenProject",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"openproject_user": resourceOpenProjectUser(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	return &ProviderConfig{
		AppURL: d.Get("app_url").(string),
		APIKey: d.Get("apikey").(string),
	}, nil
}
