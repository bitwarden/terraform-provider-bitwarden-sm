package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"regexp"
	"testing"
)

const (
	invalidSecretUUID1 = "df636133-c709-4a5f-a3dc-da28790xxxxx"
	invalidSecretUUID2 = "df636133-c709-4a5f-a3dc-da28790657b"
)

func TestAccDatasourceSecretExpectErrorOnMissingSecretId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                       data "bitwarden-sm_secret" "test" {}`,
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found"),
			},
		},
	})
}

func TestAccDatasourceSecretExpectErrorOnInvalidSecretId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                        data "bitwarden-sm_secret" "test" {
                            id = "` + invalidSecretUUID1 + `"
                        }`,
				ExpectError: regexp.MustCompile("string attribute not a valid UUID"),
			},
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                        data "bitwarden-sm_secret" "test" {
                            id = "` + invalidSecretUUID2 + `"
                        }`,
				ExpectError: regexp.MustCompile("string attribute not a valid UUID"),
			},
		},
	})
}

func TestAccDatasourceSecretVerifySecretData(t *testing.T) {
	var secretId, projectId string
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
	projectId = project.ID

	secret, preCheckError := bitwardenClient.Secrets().Create(
		secretKey,
		secretValue,
		secretNote,
		organizationId,
		[]string{projectId},
	)
	if preCheckError != nil {
		t.Fatal("Error creating test secret for provider validation.")
	}
	secretId = secret.ID

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                            data "bitwarden-sm_secret" "secret" {
                                id ="` + secretId + `"
                            }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.bitwarden-sm_secret.secret", "id", secretId),
					resource.TestCheckResourceAttr("data.bitwarden-sm_secret.secret", "value", secretValue),
					resource.TestCheckResourceAttr("data.bitwarden-sm_secret.secret", "note", secretNote),
					resource.TestCheckResourceAttr("data.bitwarden-sm_secret.secret", "organization_id", organizationId),
					resource.TestCheckResourceAttr("data.bitwarden-sm_secret.secret", "project_id", projectId),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Secrets().Delete([]string{secretId})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test secret: %s", cleanUpErr)
			}
			_, cleanUpErr = bitwardenClient.Projects().Delete([]string{projectId})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr)
			}
			return nil
		},
	})
}
