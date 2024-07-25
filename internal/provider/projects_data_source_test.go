package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"strconv"
	"testing"
)

func TestAccZeroProjectsMachineAccountWithNoAccess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile("../../.env.local.no.access") + `
                       data "bitwarden-sm_projects" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.bitwarden-sm_projects.test", "projects.#", "0"),
				),
			},
		},
	})
}

func TestAccListOneProject(t *testing.T) {
	var projectId string
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
		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile() + `
                       data "bitwarden-sm_projects" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return verifyIfProjectExistsInOutput(projectId, projectName)(s)
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Projects().Delete([]string{projectId})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr)
			}
			return nil
		},
	})
}

func TestAccListTwoProjects(t *testing.T) {
	var projectId1, projectId2 string
	projectName1 := "Test-Project-" + generateRandomString()
	projectName2 := "Test-Project-" + generateRandomString()
	bitwardenClient, organizationId, err := newBitwardenClient()
	if err != nil {
		t.Fatalf("Error creating bitwardenClient: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			project, err := bitwardenClient.Projects().Create(organizationId, projectName1)
			if err != nil {
				t.Fatal("Error creating test project for provider validation.")
			}
			projectId1 = project.ID

			project, err = bitwardenClient.Projects().Create(organizationId, projectName2)
			if err != nil {
				t.Fatal("Error creating test project for provider validation.")
			}
			projectId2 = project.ID

		},
		Steps: []resource.TestStep{
			{
				Config: buildProviderConfigFromEnvFile() + `
                       data "bitwarden-sm_projects" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return verifyIfProjectExistsInOutput(projectId1, projectName1)(s)
					},
					func(s *terraform.State) error {
						return verifyIfProjectExistsInOutput(projectId2, projectName2)(s)
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			_, cleanUpErr := bitwardenClient.Projects().Delete([]string{projectId1, projectId2})
			if cleanUpErr != nil {
				t.Fatalf("Error cleaning up test project: %s", cleanUpErr)
			}
			return nil
		},
	})
}

func verifyIfProjectExistsInOutput(projectId, projectName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources["data.bitwarden-sm_projects.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "data.bitwarden-sm_projects.test")
		}
		attributes := rs.Primary.Attributes
		numberOfProjects, err := strconv.Atoi(attributes["projects.#"])
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}

		for i := range numberOfProjects {
			key := "projects." + strconv.Itoa(i) + ".id"
			if attributes[key] == projectId {
				key = "projects." + strconv.Itoa(i) + ".name"
				if attributes[key] == projectName {
					return nil
				} else {
					return fmt.Errorf("project with ID %s found but name did not match, expected: %s, got: %s\n", projectId, projectName, attributes[key])
				}
			}

		}

		return fmt.Errorf("project with the ID: %s does not exist\n", projectId)
	}
}
