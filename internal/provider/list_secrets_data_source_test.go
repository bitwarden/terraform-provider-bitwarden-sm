package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"strconv"
	"testing"
)

func TestAccListZeroSecretsMachineAccountWithNoAccess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile("../../.env.local.no.access") + `
                       data "bitwarden-sm_secrets" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.bitwarden-sm_secrets.test", "secrets.#", "0"),
				),
			},
		},
	})
}

func TestAccListOneSecret(t *testing.T) {
	var secretId, projectId string
	secretKey := "Test-Secret-" + generateRandomString()
	projectName := "Test-Project-" + generateRandomString()
	bitwardenClient, organizationId, err := newBitwardenClient()
	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			project, preCheckErr := bitwardenClient.Projects().Create(organizationId, projectName)
			if preCheckErr != nil {
				t.Fatal("Error creating test project for provider validation.")
			}
			projectId = project.ID

			secret, preCheckErr := bitwardenClient.Secrets().Create(secretKey, "secret", "", organizationId, []string{projectId})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret for provider validation.")
			}
			secretId = secret.ID
		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile() + `
                       data "bitwarden-sm_secrets" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secretId, secretKey)(s)
					},
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

func TestAccListTwoSecrets(t *testing.T) {
	var secretId1, secretId2, projectId string
	secretKey1 := "Test-Secret-" + generateRandomString()
	secretKey2 := "Test-Secret-" + generateRandomString()
	projectName := "Test-Project-" + generateRandomString()
	bitwardenClient, organizationId, err := newBitwardenClient()
	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			project, err := bitwardenClient.Projects().Create(organizationId, projectName)
			if err != nil {
				t.Fatal("Error creating test project for provider validation.")
			}
			projectId = project.ID

			secret, err := bitwardenClient.Secrets().Create(secretKey1, "secret", "", organizationId, []string{projectId})
			if err != nil {
				t.Fatal("Error creating test secret for provider validation.")
			}
			secretId1 = secret.ID

			secret, err = bitwardenClient.Secrets().Create(secretKey2, "secret", "", organizationId, []string{projectId})
			if err != nil {
				t.Fatal("Error creating test secret for provider validation.")
			}
			secretId2 = secret.ID
		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile() + `
                       data "bitwarden-sm_secrets" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secretId1, secretKey1)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secretId2, secretKey2)(s)
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Secrets().Delete([]string{secretId1, secretId2})
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

func testAccCheckIfSecretExistsInOutput(secretId, secretKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources["data.bitwarden-sm_secrets.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "data.bitwarden-sm_secrets.test")
		}
		attributes := rs.Primary.Attributes
		numberOfProjects, err := strconv.Atoi(attributes["secrets.#"])
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}

		for i := range numberOfProjects {
			key := "secrets." + strconv.Itoa(i) + ".id"
			if attributes[key] == secretId {
				key = "secrets." + strconv.Itoa(i) + ".key"
				if attributes[key] == secretKey {
					return nil
				} else {
					return fmt.Errorf("secret with ID %s found but key did not match, expected: %s, got: %s\n", secretId, secretKey, attributes[key])
				}
			}

		}

		return fmt.Errorf("secret with the ID: %s does not exist\n", secretId)
	}
}
