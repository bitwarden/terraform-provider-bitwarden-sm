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

// This acceptance test validates that our provider implementation is compatible with Dynamic Secrets.
// Dynamic Secrets are secrets that support updated secret values and automated secret value rotation.
// To support this, our provider imports updated secret values into its own state even if terraform owns
// or manages the secret and its value was updated outside terraform. However, this feature only works
// for secret resources were no explicit value is provided in the terraform configuration.
func TestAccResourceSecretDynamicSecret(t *testing.T) {
	secretKey := "Test-Secret-" + generateRandomString()
	updatedSecretValue := generateRandomString()
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
				// This step creates a secret resource, meaning, terraform owns this resource
				// IMPORTANT: the empty value passed to buildSecretResourceConfig() creates a terraform plan
				// without an explicitly provided secret value. The secret value gets generates by the provider.
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(secretKey, "", secretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					// The following checks validate that the secret was creates successfully
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttrSet("bitwarden-sm_secret.test", "value"),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", secretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),

					// In this "check" is used to update the secret value outside terraform.
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["bitwarden-sm_secret.test"]
						if !ok {
							return fmt.Errorf("not found: %s", "bitwarden-sm_secret.test")
						}
						_, updateErr := bitwardenClient.Secrets().Update(
							rs.Primary.ID,
							secretKey,
							updatedSecretValue,
							secretNote,
							organizationId,
							[]string{project.ID},
						)
						if updateErr != nil {
							return fmt.Errorf("unable to Update Secret: %s", updateErr.Error())
						}
						return nil
					},
				),
			},
			{
				// The generated config in this step is the same as before. However, the secret value changed outside
				// terraform. But, since the provider supports Dynamic Secrets, the updated secret value does not
				// create a new plan: `ExpectNonEmptyPlan: false` and the value inside the terraform state has
				// the expected value: `TestCheckResourceAttr("bitwarden-sm_secret.test", "value", updatedSecretValue)`
				Config: buildProviderConfigFromEnvFile(t) +
					buildSecretResourceConfig(secretKey, "", secretNote, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "key", secretKey),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "value", updatedSecretValue),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "organization_id", organizationId),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "note", secretNote),
					resource.TestCheckResourceAttr("bitwarden-sm_secret.test", "project_id", project.ID),
				),
				ExpectNonEmptyPlan: false,
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
