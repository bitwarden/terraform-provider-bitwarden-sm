package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"regexp"
	"testing"
)

// TODO Add Secret Data Source Prefix to Test name to make it easier to attribute tests to resources

func TestAccExpectErrorOnMissingSecretId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                       data "bitwarden-sm_secret" "test" {}`,
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found."),
			},
		},
	})
}

func TestAccVerifySecretData(t *testing.T) {
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
