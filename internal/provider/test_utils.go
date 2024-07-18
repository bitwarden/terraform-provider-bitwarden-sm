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
	"testing"
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

	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

const (
	charset           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	envFileAccTests   = "../../.env.local.test"
	apiUrlKey         = "BW_API_URL"
	identityUrlKey    = "BW_IDENTITY_API_URL"
	accessTokenKey    = "BW_ACCESS_TOKEN"
	organizationIDKey = "BW_ORGANIZATION_ID"
	stateFileKey      = "BW_STATE_FILE"
)

func generateRandomString() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

func resolveFilePath(filePath []string) string {
	if len(filePath) == 0 {
		return envFileAccTests
	}
	if len(filePath) > 1 {
		log.Println("The calling test function passed more than 1 filePath to .env files. Only the first was used.")
	}
	return filePath[0]
}

func readEnvFile(envFile string) (map[string]string, error) {
	envMap, err := godotenv.Read(envFile)
	if err != nil {
		return nil, fmt.Errorf("error loading %s file: %w", envFile, err)
	}

	mandatoryKeys := map[string]string{
		apiUrlKey:         "apiUrl is missing or empty",
		identityUrlKey:    "identityUrl is missing or empty",
		accessTokenKey:    "accessToken is missing or empty",
		organizationIDKey: "organizationId is missing or empty",
		stateFileKey:      "stateFile is missing or empty",
	}

	for key, errMsg := range mandatoryKeys {
		if value, exists := envMap[key]; !exists || value == "" {
			return nil, fmt.Errorf("%s. Please verify .env file: %s", errMsg, envFile)
		}
	}

	return envMap, nil
}

func newBitwardenClient(filePath ...string) (sdk.BitwardenClientInterface, string, error) {
	envFilePath := resolveFilePath(filePath)
	envMap, err := readEnvFile(envFilePath)
	if err != nil {
		return nil, "", err
	}

	apiUrl := envMap[apiUrlKey]
	identityUrl := envMap[identityUrlKey]
	accessToken := envMap[accessTokenKey]
	organizationId := envMap[organizationIDKey]
	stateFile := envMap[stateFileKey]

	bitwardenClient, err := sdk.NewBitwardenClient(&apiUrl, &identityUrl)
	if err != nil {
		return nil, "", err
	}

	err = bitwardenClient.AccessTokenLogin(accessToken, &stateFile)
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
	keys := []string{apiUrlKey,
		identityUrlKey,
		accessTokenKey,
		organizationIDKey,
		stateFileKey}

	for _, key := range keys {
		err := os.Unsetenv(key)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildProviderConfigFromEnvFile(t *testing.T, filePath ...string) string {
	envFilePath := resolveFilePath(filePath)
	envMap, err := readEnvFile(envFilePath)
	if err != nil {
		t.Fatalf("Error during provider configuration build: %s", err.Error())
	}

	apiUrl := envMap[apiUrlKey]
	identityUrl := envMap[identityUrlKey]
	accessToken := envMap[accessTokenKey]
	organizationId := envMap[organizationIDKey]

	providerConfig := fmt.Sprintf(`
        provider "bitwarden-sm" {
            api_url = "%s"
            identity_url = "%s"
            access_token = "%s"
            organization_id = "%s"
        }`, apiUrl, identityUrl, accessToken, organizationId)

	return providerConfig
}

func buildSecretResourceConfig(key, value, note, projectId string) string {
	return fmt.Sprintf(`

        resource "bitwarden-sm_secret" "test" {
            key = "%s"
            value = "%s"
            note = "%s"
            project_id = "%s"
        }
        `, key, value, note, projectId)
}
