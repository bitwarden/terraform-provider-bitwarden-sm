package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"regexp"
	"testing"
)

const (
	validProjectUUID    = "df636133-c709-4a5f-a3dc-da28790657b5"
	invalidProjectUUID1 = "df636133-c709-4a5f-a3dc-da28790xxxxx"
	invalidProjectUUID2 = "df636133-c709-4a5f-a3dc-da28790657b"
)

func TestAccProviderExpectErrorOnMissingApiUrlInProviderConfigString(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: `provider "bitwarden-sm" {
                            identity_url = "https://identity.example.com"
                            access_token = "mock_access_token"
                            organization_id = "` + validProjectUUID + `"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing URI for Bitwarden Secrets Manager API endpoint"),
			},
		},
	})
}

func TestAccProviderExpectErrorOnMissingIdentityUrlInProviderConfigString(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: `provider "bitwarden-sm" {
                            api_url      = "https://api.example.com"
                            access_token = "mock_access_token"
                            organization_id = "` + validProjectUUID + `"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing URI for Bitwarden Secrets Manager IDENTITY endpoint"),
			},
		},
	})
}

func TestAccProviderExpectErrorOnMissingApiAndIdentityUrlInProviderConfigString1(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: `provider "bitwarden-sm" {
                            access_token = "mock_access_token"
                            organization_id = "` + validProjectUUID + `"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing URI for Bitwarden Secrets Manager API endpoint"),
			},
		},
	})
}

func TestAccProviderExpectErrorOnMissingApiAndIdentityUrlInProviderConfigString2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: `provider "bitwarden-sm" {
                            access_token = "mock_access_token"
                            organization_id = "` + validProjectUUID + `"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing URI for Bitwarden Secrets Manager IDENTITY endpoint"),
			},
		},
	})
}

func TestAccProviderExpectErrorOnMissingAccessTokenInProviderConfigString(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: `provider "bitwarden-sm" {
                            api_url      = "https://api.example.com"
                            identity_url = "https://identity.example.com"
                            organization_id = "` + validProjectUUID + `"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing Bitwarden Secrets Manager Access Token"),
			},
		},
	})
}

func TestAccProviderExpectErrorOnMissingOrganizationIdInProviderConfigString(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: `provider "bitwarden-sm" {
                            api_url      = "https://api.example.com"
                            identity_url = "https://identity.example.com"
                            access_token = "mock_access_token"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing Bitwarden Secrets Manager Organization ID"),
			},
		},
	})
}

func TestAccProviderExpectErrorOnOrganizationIdNotAValidUUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: `provider "bitwarden-sm" {
                            api_url         = "https://api.example.com"
                            identity_url    = "https://identity.example.com"
                            access_token    = "mock_access_token"
                            organization_id = "` + invalidProjectUUID1 + `"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("string attribute not a valid UUID"),
			},
			{
				Config: `provider "bitwarden-sm" {
                            api_url         = "https://api.example.com"
                            identity_url    = "https://identity.example.com"
                            access_token    = "mock_access_token"
                            organization_id = "` + invalidProjectUUID2 + `"
                        }

                        data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("string attribute not a valid UUID"),
			},
		},
	})
}

func TestAccProviderExpectErrorOnMissingApiUrlInEnvVars(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			preCheckUnsetAllEnvVars()

			err := os.Setenv("BW_IDENTITY_API_URL", "https://identity.example.com")
			if err != nil {
				t.Fatal("Setting BW_IDENTITY_API_URL in acceptance tests failed")
			}
			err = os.Setenv("BW_ACCESS_TOKEN", "mock_access_token")
			if err != nil {
				t.Fatal("Setting BW_ACCESS_TOKEN in acceptance tests failed")
			}
			err = os.Setenv("BW_ORGANIZATION_ID", validProjectUUID)
			if err != nil {
				t.Fatal("Setting BW_ORGANIZATION_ID in acceptance tests failed")
			}
		},
		Steps: []resource.TestStep{
			{
				Config:      `data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing URI for Bitwarden Secrets Manager API endpoint"),
			},
		},
		CheckDestroy: checkDestroyUnsetAllEnvVars,
	})
}

func TestAccProviderExpectErrorOnMissingIdentityUrlInEnvVars(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			preCheckUnsetAllEnvVars()

			err := os.Setenv("BW_API_URL", "https://api.example.com")
			if err != nil {
				t.Fatal("Setting BW_API_URL in acceptance tests failed")
			}
			err = os.Setenv("BW_ACCESS_TOKEN", "mock_access_token")
			if err != nil {
				t.Fatal("Setting BW_ACCESS_TOKEN in acceptance tests failed")
			}
			err = os.Setenv("BW_ORGANIZATION_ID", validProjectUUID)
			if err != nil {
				t.Fatal("Setting BW_ORGANIZATION_ID in acceptance tests failed")
			}
		},
		Steps: []resource.TestStep{
			{
				Config:      `data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing URI for Bitwarden Secrets Manager IDENTITY endpoint"),
			},
		},
		CheckDestroy: checkDestroyUnsetAllEnvVars,
	})
}

func TestAccProviderExpectErrorOnMissingApiAndIdentityUrlInEnvVars(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			preCheckUnsetAllEnvVars()

			err := os.Setenv("BW_ACCESS_TOKEN", "mock_access_token")
			if err != nil {
				t.Fatal("Setting BW_ACCESS_TOKEN in acceptance tests failed")
			}
			err = os.Setenv("BW_ORGANIZATION_ID", validProjectUUID)
			if err != nil {
				t.Fatal("Setting BW_ORGANIZATION_ID in acceptance tests failed")
			}
		},
		Steps: []resource.TestStep{
			{
				Config:      `data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing URI for Bitwarden Secrets Manager API endpoint"),
			},
		},
		CheckDestroy: checkDestroyUnsetAllEnvVars,
	})
}

func TestAccProviderExpectErrorOnMissingAccessTokenInEnvVars(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			preCheckUnsetAllEnvVars()

			err := os.Setenv("BW_API_URL", "https://api.example.com")
			if err != nil {
				t.Fatal("Setting BW_API_URL in acceptance tests failed")
			}
			err = os.Setenv("BW_IDENTITY_API_URL", "https://identity.example.com")
			if err != nil {
				t.Fatal("Setting BW_IDENTITY_API_URL in acceptance tests failed")
			}
			err = os.Setenv("BW_ORGANIZATION_ID", validProjectUUID)
			if err != nil {
				t.Fatal("Setting BW_ORGANIZATION_ID in acceptance tests failed")
			}
		},
		Steps: []resource.TestStep{
			{
				Config:      `data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing Bitwarden Secrets Manager Access Token"),
			},
		},
		CheckDestroy: checkDestroyUnsetAllEnvVars,
	})
}

func TestAccProviderExpectErrorOnMissingOrganizationIdInEnvVars(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			preCheckUnsetAllEnvVars()

			err := os.Setenv("BW_API_URL", "https://api.example.com")
			if err != nil {
				t.Fatal("Setting BW_API_URL in acceptance tests failed")
			}
			err = os.Setenv("BW_IDENTITY_API_URL", "https://identity.example.com")
			if err != nil {
				t.Fatal("Setting BW_IDENTITY_API_URL in acceptance tests failed")
			}
			err = os.Setenv("BW_ACCESS_TOKEN", "mock_access_token")
			if err != nil {
				t.Fatal("Setting BW_ACCESS_TOKEN in acceptance tests failed")
			}
		},
		Steps: []resource.TestStep{
			{
				Config:      `data "bitwarden-sm_projects" "projects" {}`,
				ExpectError: regexp.MustCompile("Missing Bitwarden Secrets Manager Organization ID"),
			},
		},
		CheckDestroy: checkDestroyUnsetAllEnvVars,
	})
}

func TestAccProviderConfigurationValid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 preCheckUnsetAllEnvVars,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                        data "bitwarden-sm_projects" "projects" {}`,
			},
		},
	})
}
