package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpenProjectUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenProjectUserCreate,
		Delete: resourceOpenProjectUserDelete,
		Read:   resourceOpenProjectUserRead,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"firstname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lastname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: resourceOpenProjectUserImport,
		},
	}
}

func resourceOpenProjectUserCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	appURL := config.AppURL
	apikey := config.APIKey

	url := fmt.Sprintf("%s/api/v3/users", appURL)
	credentials := fmt.Sprintf("apikey:%s", apikey)
	auth := base64.StdEncoding.EncodeToString([]byte(credentials))

	body := map[string]interface{}{
		"login":     d.Get("username").(string),
		"password":  d.Get("password").(string),
		"firstName": d.Get("firstname").(string),
		"lastName":  d.Get("lastname").(string),
		"email":     d.Get("email").(string),
		"admin":     false,
		"status":    "active",
		"language":  "en",
	}

	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error creating user: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create user: status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	d.SetId(fmt.Sprintf("%v", response["id"]))
	return nil
}

func resourceOpenProjectUserDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	appURL := config.AppURL
	apikey := config.APIKey
	userID := d.Id()

	url := fmt.Sprintf("%s/api/v3/users/%s", appURL, userID)
	credentials := fmt.Sprintf("apikey:%s", apikey)
	auth := base64.StdEncoding.EncodeToString([]byte(credentials))

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting user: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusAccepted {
		return nil
	}

	var errResp struct {
		Message string `json:"message"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &errResp)

	return fmt.Errorf("failed to delete user (status %d): %s", resp.StatusCode, errResp.Message)
}

func resourceOpenProjectUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	appURL := config.AppURL
	apikey := config.APIKey
	userID := d.Id()

	url := fmt.Sprintf("%s/api/v3/users/%s", appURL, userID)
	credentials := fmt.Sprintf("apikey:%s", apikey)
	auth := base64.StdEncoding.EncodeToString([]byte(credentials))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating GET request: %s", err)
	}
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending GET request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to read user (status %d): %s", resp.StatusCode, string(body))
	}

	var user struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("error decoding user JSON: %s", err)
	}

	d.Set("username", user.Login)
	d.Set("email", user.Email)
	d.Set("firstname", user.FirstName)
	d.Set("lastname", user.LastName)

	return nil
}

func resourceOpenProjectUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	id := d.Id()
	if id == "" {
		return nil, fmt.Errorf("no ID provided for import")
	}

	d.SetId(id)

	if err := resourceOpenProjectUserRead(d, meta); err != nil {
		return nil, fmt.Errorf("error reading user during import: %s", err)
	}

	return []*schema.ResourceData{d}, nil
}
