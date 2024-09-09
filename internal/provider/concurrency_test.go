package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"
)

// This acceptance test validates that the concurrency issues we experienced when reading multiple objects in parallel
// are fixed. Prior to v1.0.0, the sdk-go in had concurrency issues: https://github.com/bitwarden/sdk/pull/981
// Therefore, this test creates multiple projects and secrets and ensures that reading all of them utilizing multiple
// data sources works as intended.
func TestAccConcurrency(t *testing.T) {
	bitwardenClient, organizationId, err := newBitwardenClient()

	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}

	project1, preCheckError := bitwardenClient.Projects().Create(organizationId, "Test-Project-"+generateRandomString())
	if preCheckError != nil {
		t.Fatal("Error creating test project for provider validation.")
	}

	project2, preCheckError := bitwardenClient.Projects().Create(organizationId, "Test-Project-"+generateRandomString())
	if preCheckError != nil {
		t.Fatal("Error creating test project for provider validation.")
	}

	secret1, preCheckErr := bitwardenClient.Secrets().Create("Test-Secret-"+generateRandomString(), "secret", "", organizationId, []string{project1.ID})
	if preCheckErr != nil {
		t.Fatal("Error creating test secret for provider validation.")
	}

	secret2, preCheckErr := bitwardenClient.Secrets().Create("Test-Secret-"+generateRandomString(), "secret", "", organizationId, []string{project2.ID})
	if preCheckErr != nil {
		t.Fatal("Error creating test secret for provider validation.")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                       data "bitwarden-sm_list_secrets" "test" {}
					   data "bitwarden-sm_projects" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secret1.ID, secret1.Key)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secret2.ID, secret2.Key)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfProjectExistsInOutput(project1.ID, project1.Name)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfProjectExistsInOutput(project2.ID, project2.Name)(s)
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Secrets().Delete([]string{secret1.ID, secret2.ID})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test secret: %s", cleanUpErr)
			}
			_, cleanUpErr = bitwardenClient.Projects().Delete([]string{project1.ID, project2.ID})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr)
			}
			return nil
		},
	})
}
