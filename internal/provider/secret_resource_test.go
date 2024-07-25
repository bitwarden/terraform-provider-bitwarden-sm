package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"regexp"
	"strings"
	"testing"
)

func TestAccResourceSecretExpectErrorOnMissingKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                       resource "bitwarden-sm_secret" "test" {
                                value          = "mock-value"
                            }`,
				ExpectError: regexp.MustCompile("The argument \"key\" is required, but no definition was found."),
			},
		},
	})
}

func TestAccResourceSecretExpectErrorOnMissingValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                       resource "bitwarden-sm_secret" "test" {
                                key          = "mock-key"
                            }`,
				ExpectError: regexp.MustCompile("The argument \"value\" is required, but no definition was found."),
			},
		},
	})
}

func TestAccResourceSecretCreateSecret(t *testing.T) {
	secretKey := "Test-Secret-" + generateRandomString()
	secretValue := generateRandomString()
	secretNote := generateRandomString()
	projectName := "Test-Project-" + generateRandomString()

	bitwardenClient, organizationId, err := newBitwardenClient()

	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}

	project, preCheckError := bitwardenClient.Projects().Create(organizationId, projectName)
	if preCheckError != nil {
		t.Fatal("Error creating test project for provider validation.")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(secretKey, secretValue, secretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", secretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", secretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Projects().Delete([]string{project.ID})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr.Error())
			}
			return nil
		},
	})
}

func TestAccResourceSecretUpdateSecret(t *testing.T) {
	secretKey := "Test-Secret-" + generateRandomString()
	updatedSecretKey := "Test-Secret-" + generateRandomString()
	secretValue := generateRandomString()
	updatedSecretValue := generateRandomString()
	secretNote := generateRandomString()
	UpdatedSecretNote := generateRandomString()
	projectName := "Test-Project-" + generateRandomString()
	updatedProjectName := "Test-Project-" + generateRandomString()

	bitwardenClient, organizationId, err := newBitwardenClient()

	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}

	project, preCheckError := bitwardenClient.Projects().Create(organizationId, projectName)
	if preCheckError != nil {
		t.Fatal("Error creating test project for provider validation.")
	}

	updatedProject, preCheckError := bitwardenClient.Projects().Create(organizationId, updatedProjectName)
	if preCheckError != nil {
		t.Fatal("Error creating test project for provider validation.")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(secretKey, secretValue, secretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", secretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", secretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
				),
			},
			{
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(secretKey, updatedSecretValue, secretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", updatedSecretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", secretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
				),
			},
			{
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(secretKey, updatedSecretValue, UpdatedSecretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", updatedSecretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", UpdatedSecretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
				),
			},
			{
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(updatedSecretKey, updatedSecretValue, UpdatedSecretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", updatedSecretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", updatedSecretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", UpdatedSecretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
				),
			},
			{
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(updatedSecretKey, updatedSecretValue, UpdatedSecretNote, updatedProject.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", updatedSecretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", updatedSecretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", UpdatedSecretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", updatedProject.ID),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Projects().Delete([]string{project.ID})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr.Error())
			}
			_, cleanUpErr = bitwardenClient.Projects().Delete([]string{updatedProject.ID})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr.Error())
			}
			return nil
		},
	})
}

func TestAccResourceSecretDeleteSecret(t *testing.T) {
	secretKey := "Test-Secret-" + generateRandomString()
	secretValue := generateRandomString()
	secretNote := generateRandomString()
	projectName := "Test-Project-" + generateRandomString()

	bitwardenClient, organizationId, err := newBitwardenClient()
	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}

	project, preCheckError := bitwardenClient.Projects().Create(organizationId, projectName)
	if preCheckError != nil {
		t.Fatal("Error creating test project for provider validation.")
	}

	var secretID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(secretKey, secretValue, secretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", secretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", secretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["bitwarden-sm_secret.test"]
						if !ok {
							return fmt.Errorf("not found: %s", "bitwarden-sm_secret.test")
						}
						secretID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: buildProviderConfigFromEnvFile(t),
				Check: func(s *terraform.State) error {
					// Verify the secret no longer exists
					_, checkErr := bitwardenClient.Secrets().Get(secretID)
					if checkErr == nil {
						return fmt.Errorf("secret still exists: %s", secretID)
					}

					if !strings.Contains(checkErr.Error(), "404 Not Found") {
						return fmt.Errorf("expected api error message to contain: '404 Not Found', got: %s", checkErr.Error())
					}

					return nil
				},
			},
			{
				Config: buildProviderConfigFromEnvFile(t),
				Check: func(s *terraform.State) error {
					// Clean up test project
					// Needs to run here because CheckDestroy is not executed after previous Delete step
					_, cleanUpErr := bitwardenClient.Projects().Delete([]string{project.ID})
					if cleanUpErr != nil {
						t.Fatalf("Error cleaning up test project: %s", cleanUpErr.Error())
					}

					return nil
				},
			},
		},
	})
}

func TestAccResourceSecretImportSecret(t *testing.T) {
	secretKey := "Test-Secret-" + generateRandomString()
	secretValue := generateRandomString()
	secretNote := generateRandomString()
	projectName := "Test-Project-" + generateRandomString()

	bitwardenClient, organizationId, err := newBitwardenClient()
	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}

	project, preCheckError := bitwardenClient.Projects().Create(organizationId, projectName)
	if preCheckError != nil {
		t.Fatal("Error creating test project for provider validation.")
	}

	secret, preCheckError := bitwardenClient.Secrets().Create(
		secretKey,
		secretValue,
		secretNote,
		organizationId,
		[]string{project.ID},
	)
	if preCheckError != nil {
		t.Fatal("Error creating test secret for provider validation.")
	}

	secretId := secret.ID

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ResourceName:  "bitwarden-sm_secret.test",
				ImportState:   true,
				ImportStateId: secretId,
				Config: buildProviderConfigFromEnvFile(t) + `
                                    resource "bitwarden-sm_secret" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", secretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", secretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
				),
			},
			{
				Config: buildProviderConfigFromEnvFile(t),
				Check: func(s *terraform.State) error {
					// Clean up test project and secret
					// Needs to run here because CheckDestroy is not executed after previous Delete step
					_, cleanUpErr := bitwardenClient.Secrets().Delete([]string{secretId})
					if cleanUpErr != nil {
						t.Fatalf("Error cleaning up test secret: %s", cleanUpErr)
					}
					_, cleanUpErr = bitwardenClient.Projects().Delete([]string{project.ID})
					if cleanUpErr != nil {
						t.Fatalf("Error cleaning up test project: %s", cleanUpErr)
					}
					return nil
				},
			},
		},
	})
}
