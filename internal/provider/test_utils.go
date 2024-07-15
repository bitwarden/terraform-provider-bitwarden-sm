package provider

import (
	"fmt"
	"github.com/bitwarden/sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"bitwarden-sm": providerserver.NewProtocol6WithError(New("test")()),
	}

	envFileAccTests = "../../.env.local.test"
)

func newBitwardenClient() (sdk.BitwardenClientInterface, string, error) {
	apiUrl, identityUrl, accessToken, organizationId := readEnvFile(envFileAccTests)

	bitwardenClient, err := sdk.NewBitwardenClient(&apiUrl, &identityUrl)
	if err != nil {
		return nil, "", err
	}

	testStatePath := "./"
	err = bitwardenClient.AccessTokenLogin(accessToken, &testStatePath)
	if err != nil {
		return nil, "", err
	}

	return bitwardenClient, organizationId, nil
}

func preCheckUnsetAllEnvVars() {
	err := unsetAllEnvVars()
	if err != nil {
		log.Fatalf("Error unsetting environment variables: %v\n", err)
	}
}

func checkDestroyUnsetAllEnvVars(_ *terraform.State) error {
	return unsetAllEnvVars()
}

func unsetAllEnvVars() error {
	vars := []string{"BW_API_URL", "BW_IDENTITY_API_URL", "BW_ACCESS_TOKEN", "BW_ORGANIZATION_ID"}
	for _, v := range vars {
		if err := os.Unsetenv(v); err != nil {
			return err
		}
	}
	return nil
}

func buildProviderConfigFromEnvFile(filePath ...string) string {
	if len(filePath) > 0 {
		envFileAccTests = filePath[0]
	}

	if len(filePath) > 1 {
		log.Println("The calling test function passed more than 1 filePath to .env files. Only the first was used.")
	}

	apiUrl, identityUrl, accessToken, organizationId := readEnvFile(envFileAccTests)

	providerConfig := fmt.Sprintf(`
        provider "bitwarden-sm" {
            api_url = "%s"
            identity_url = "%s"
            access_token = "%s"
            organization_id = "%s"
        }`, apiUrl, identityUrl, accessToken, organizationId)

	return providerConfig
}

func readEnvFile(envFile string) (string, string, string, string) {
	envMap, err := godotenv.Read(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file during provider configuration\n", envFile)
	}

	apiUrl := envMap["BW_API_URL"]
	identityUrl := envMap["BW_IDENTITY_API_URL"]
	accessToken := envMap["BW_ACCESS_TOKEN"]
	organizationId := envMap["BW_ORGANIZATION_ID"]

	if apiUrl == "" {
		log.Fatalf("Provider configuration value apiUrl either missing or empty. Please verify .env file: %s", envFile)
	}

	if identityUrl == "" {
		log.Fatalf("Provider configuration value identityUrl either missing or empty. Please verify .env file: %s", envFile)
	}

	if accessToken == "" {
		log.Fatalf("Provider configuration value accessToken either missing or empty. Please verify .env file: %s", envFile)
	}

	if organizationId == "" {
		log.Fatalf("Provider configuration value organizationId either missing or empty. Please verify .env file: %s", envFile)
	}

	return apiUrl, identityUrl, accessToken, organizationId
}

func generateRandomString() string {
	charset := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(charset))]
	}
	return string(b)
}
