package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatasourceListSecretsZeroSecretsMachineAccountWithNoAccess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t, "../../.env.local.no.access") + `
                       data "bitwarden-secrets_list_secrets" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.bitwarden-secrets_list_secrets.test", "secrets.#", "0"),
				),
			},
		},
	})
}

func TestAccDatasourceListSecretsOneSecret(t *testing.T) {
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
				Config: buildProviderConfigFromEnvFile(t) + `
                       data "bitwarden-secrets_list_secrets" "test" {}`,
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

func TestAccDatasourceListSecretsTwoSecrets(t *testing.T) {
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
			project, preCheckErr := bitwardenClient.Projects().Create(organizationId, projectName)
			if preCheckErr != nil {
				t.Fatal("Error creating test project for provider validation.")
			}
			projectId = project.ID

			secret, preCheckErr := bitwardenClient.Secrets().Create(secretKey1, "secret", "", organizationId, []string{projectId})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret for provider validation.")
			}
			secretId1 = secret.ID

			secret, preCheckErr = bitwardenClient.Secrets().Create(secretKey2, "secret", "", organizationId, []string{projectId})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret for provider validation.")
			}
			secretId2 = secret.ID
		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                       data "bitwarden-secrets_list_secrets" "test" {}`,
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

func TestAccDatasourceListSecretsFilterByProjectId(t *testing.T) {
	var secretId1, secretId2, projectId1, projectId2 string
	secretKey1 := "Test-Secret-" + generateRandomString()
	secretKey2 := "Test-Secret-" + generateRandomString()
	projectName1 := "Test-Project-" + generateRandomString()
	projectName2 := "Test-Project-" + generateRandomString()
	bitwardenClient, organizationId, err := newBitwardenClient()
	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			project1, preCheckErr := bitwardenClient.Projects().Create(organizationId, projectName1)
			if preCheckErr != nil {
				t.Fatal("Error creating test project 1 for provider validation.")
			}
			projectId1 = project1.ID

			project2, preCheckErr := bitwardenClient.Projects().Create(organizationId, projectName2)
			if preCheckErr != nil {
				t.Fatal("Error creating test project 2 for provider validation.")
			}
			projectId2 = project2.ID

			secret1, preCheckErr := bitwardenClient.Secrets().Create(secretKey1, "secret1", "", organizationId, []string{projectId1})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret 1 for provider validation.")
			}
			secretId1 = secret1.ID

			secret2, preCheckErr := bitwardenClient.Secrets().Create(secretKey2, "secret2", "", organizationId, []string{projectId2})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret 2 for provider validation.")
			}
			secretId2 = secret2.ID
		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + fmt.Sprintf(`
                       data "bitwarden-secrets_list_secrets" "test" {
                         project_id = "%s"
                       }`, projectId1),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secretId1, secretKey1)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfSecretDoesNotExistInOutput(secretId2)(s)
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Secrets().Delete([]string{secretId1, secretId2})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test secrets: %s", cleanUpErr)
			}
			_, cleanUpErr = bitwardenClient.Projects().Delete([]string{projectId1, projectId2})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test projects: %s", cleanUpErr)
			}
			return nil
		},
	})
}

func TestAccDatasourceListSecretsFilterByKey(t *testing.T) {
	var secretId1, secretId2, projectId string
	uniqueToken := generateRandomString()
	secretKey1 := "Alpha-Secret-" + uniqueToken
	secretKey2 := "Bravo-Secret-" + generateRandomString()
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

			secret1, preCheckErr := bitwardenClient.Secrets().Create(secretKey1, "secret1", "", organizationId, []string{projectId})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret 1 for provider validation.")
			}
			secretId1 = secret1.ID

			secret2, preCheckErr := bitwardenClient.Secrets().Create(secretKey2, "secret2", "", organizationId, []string{projectId})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret 2 for provider validation.")
			}
			secretId2 = secret2.ID
		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + fmt.Sprintf(`
                       data "bitwarden-secrets_list_secrets" "test" {
                         filter = "Alpha-Secret-%s"
                       }`, uniqueToken),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secretId1, secretKey1)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfSecretDoesNotExistInOutput(secretId2)(s)
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Secrets().Delete([]string{secretId1, secretId2})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test secrets: %s", cleanUpErr)
			}
			_, cleanUpErr = bitwardenClient.Projects().Delete([]string{projectId})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr)
			}
			return nil
		},
	})
}

func TestAccDatasourceListSecretsFilterByProjectIdAndKey(t *testing.T) {
	var secretId1, secretId2, secretId3, projectId1, projectId2 string
	uniqueToken := generateRandomString()
	secretKey1 := "Gamma-Secret-" + uniqueToken
	secretKey2 := "Delta-Secret-" + generateRandomString()
	secretKey3 := "Gamma-Secret-" + generateRandomString()
	projectName1 := "Test-Project-" + generateRandomString()
	projectName2 := "Test-Project-" + generateRandomString()
	bitwardenClient, organizationId, err := newBitwardenClient()
	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			project1, preCheckErr := bitwardenClient.Projects().Create(organizationId, projectName1)
			if preCheckErr != nil {
				t.Fatal("Error creating test project 1 for provider validation.")
			}
			projectId1 = project1.ID

			project2, preCheckErr := bitwardenClient.Projects().Create(organizationId, projectName2)
			if preCheckErr != nil {
				t.Fatal("Error creating test project 2 for provider validation.")
			}
			projectId2 = project2.ID

			secret1, preCheckErr := bitwardenClient.Secrets().Create(secretKey1, "secret1", "", organizationId, []string{projectId1})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret 1 for provider validation.")
			}
			secretId1 = secret1.ID

			secret2, preCheckErr := bitwardenClient.Secrets().Create(secretKey2, "secret2", "", organizationId, []string{projectId1})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret 2 for provider validation.")
			}
			secretId2 = secret2.ID

			secret3, preCheckErr := bitwardenClient.Secrets().Create(secretKey3, "secret3", "", organizationId, []string{projectId2})
			if preCheckErr != nil {
				t.Fatal("Error creating test secret 3 for provider validation.")
			}
			secretId3 = secret3.ID
		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + fmt.Sprintf(`
                       data "bitwarden-secrets_list_secrets" "test" {
                         project_id = "%s"
                         filter = "Gamma-Secret-%s"
                       }`, projectId1, uniqueToken),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckIfSecretExistsInOutput(secretId1, secretKey1)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfSecretDoesNotExistInOutput(secretId2)(s)
					},
					func(s *terraform.State) error {
						return testAccCheckIfSecretDoesNotExistInOutput(secretId3)(s)
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Secrets().Delete([]string{secretId1, secretId2, secretId3})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test secrets: %s", cleanUpErr)
			}
			_, cleanUpErr = bitwardenClient.Projects().Delete([]string{projectId1, projectId2})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test projects: %s", cleanUpErr)
			}
			return nil
		},
	})
}

func TestAccDatasourceListSecretsFilterByProjectIdInvalidUUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile(t) + `
                       data "bitwarden-secrets_list_secrets" "test" {
                         project_id = "not-a-valid-uuid"
                       }`,
				ExpectError: regexp.MustCompile("not a valid UUID"),
			},
		},
	})
}

func testAccCheckIfSecretExistsInOutput(secretId, secretKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources["data.bitwarden-secrets_list_secrets.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "data.bitwarden-secrets_secrets.test")
		}
		attributes := rs.Primary.Attributes
		numberOfSecrets, err := strconv.Atoi(attributes["secrets.#"])
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}

		for i := range numberOfSecrets {
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

func testAccCheckIfSecretDoesNotExistInOutput(secretId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["data.bitwarden-secrets_list_secrets.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "data.bitwarden-secrets_list_secrets.test")
		}
		attributes := rs.Primary.Attributes
		numberOfSecrets, err := strconv.Atoi(attributes["secrets.#"])
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}

		for i := range numberOfSecrets {
			key := "secrets." + strconv.Itoa(i) + ".id"
			if attributes[key] == secretId {
				return fmt.Errorf("secret with ID %s should not exist in output but was found\n", secretId)
			}
		}

		return nil
	}
}
