package secretmanager_test

import (
	"testing"
	"regexp"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
)

func TestAccSecretManagerSecretVersion_writeOnlyVersioning(t *testing.T) {
	t.Parallel()

	ctx := map[string]interface{}{
		"rs": acctest.RandString(t, 8),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretManagerSecretVersion_noPayload(ctx),
				ExpectError: regexp.MustCompile(`(?i)(one of).*(secret_data).*(secret_data_wo).*(must be (set|specified))|Field \[payload\] is required`),
			},
			{
				Config: testAccSecretManagerSecretVersion_writeOnlyConfig(ctx, 1, "alpha"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("google_secret_manager_secret_version.api_key_opsgenie_secret", "secret_data"),
					resource.TestCheckNoResourceAttr("google_secret_manager_secret_version.api_key_opsgenie_secret", "secret_data_wo"),
				),
			},
			{
				ResourceName:            "google_secret_manager_secret_version.api_key_opsgenie_secret",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_data", "secret_data_wo_version"},
			},
			{
				Config: testAccSecretManagerSecretVersion_writeOnlyConfigDisabled(ctx, 1, "alpha"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("google_secret_manager_secret_version.api_key_opsgenie_secret", "secret_data"),
					resource.TestCheckNoResourceAttr("google_secret_manager_secret_version.api_key_opsgenie_secret", "secret_data_wo"),
				),
			},
			{
				ResourceName:            "google_secret_manager_secret_version.api_key_opsgenie_secret",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_data", "secret_data_wo_version"},
			},
			{
				Config: testAccSecretManagerSecretVersion_writeOnlyConfig(ctx, 2, "beta"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("google_secret_manager_secret_version.api_key_opsgenie_secret", "secret_data"),
					resource.TestCheckNoResourceAttr("google_secret_manager_secret_version.api_key_opsgenie_secret", "secret_data_wo"),
				),
			},
			{
				ResourceName:            "google_secret_manager_secret_version.api_key_opsgenie_secret",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_data", "secret_data_wo_version"},
			},
		},
	})
}

func testAccSecretManagerSecretVersion_noPayload(ctx map[string]interface{}) string {
	return fmt.Sprintf(`
resource "google_secret_manager_secret" "api_key_opsgenie" {
  secret_id = "api-key-opsgenie-%[1]s"
  replication {
    user_managed {
      replicas {
        location = "us-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "api_key_opsgenie_secret" {
  secret = google_secret_manager_secret.api_key_opsgenie.id
  # sem secret_data / secret_data_wo / secret_data_wo_version
}
`, ctx["rs"])
}

func testAccSecretManagerSecretVersion_writeOnlyConfig(ctx map[string]interface{}, bump int, value string) string {
	return fmt.Sprintf(`
variable "secret_data_wo_versioning" {
  default = %[2]d
}

variable "secret_value" {
  default = "%[3]s"
}

resource "google_secret_manager_secret" "api_key_opsgenie" {
  secret_id = "api-key-opsgenie-%[1]s"
  replication {
    user_managed {
      replicas {
        location = "us-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "api_key_opsgenie_secret" {
  secret                 = google_secret_manager_secret.api_key_opsgenie.id
  secret_data_wo         = var.secret_value
  secret_data_wo_version = var.secret_data_wo_versioning
}
`, ctx["rs"], bump, value)
}

func testAccSecretManagerSecretVersion_writeOnlyConfigDisabled(ctx map[string]interface{}, bump int, value string) string {
	return fmt.Sprintf(`
variable "secret_data_wo_versioning" {
  default = %[2]d
}

variable "secret_value" {
  default = "%[3]s"
}

resource "google_secret_manager_secret" "api_key_opsgenie" {
  secret_id = "api-key-opsgenie-%[1]s"
  replication {
    user_managed {
      replicas {
        location = "us-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "api_key_opsgenie_secret" {
  secret                 = google_secret_manager_secret.api_key_opsgenie.id
  secret_data_wo         = var.secret_value
  secret_data_wo_version = var.secret_data_wo_versioning
  enabled                = false
}
`, ctx["rs"], bump, value)
}

func TestAccSecretManagerSecretVersion_update(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckSecretManagerSecretVersionDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretManagerSecretVersion_basic(context),
			},
			{
				ResourceName:            "google_secret_manager_secret_version.secret-version-basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_data", "secret_data_wo_version"},
			},
			{
				Config: testAccSecretManagerSecretVersion_disable(context),
			},
			{
				ResourceName:      "google_secret_manager_secret_version.secret-version-basic",
				ImportState:       true,
				ImportStateVerify: true,
				// at this point the secret data is disabled and so reading the data on import will
				// give an empty string
				ImportStateVerifyIgnore: []string{"secret_data", "secret_data_wo_version"},
			},
			{
				Config: testAccSecretManagerSecretVersion_basic(context),
			},
			{
				ResourceName:            "google_secret_manager_secret_version.secret-version-basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_data", "secret_data_wo_version"},
			},
		},
	})
}

func testAccSecretManagerSecretVersion_basic(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_secret_manager_secret" "secret-basic" {
  secret_id = "tf-test-secret-version-%{random_suffix}"
  
  labels = {
    label = "my-label"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic" {
  secret = google_secret_manager_secret.secret-basic.name

  secret_data = "my-tf-test-secret%{random_suffix}"
  enabled = true
}
`, context)
}

func testAccSecretManagerSecretVersion_disable(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_secret_manager_secret" "secret-basic" {
  secret_id = "tf-test-secret-version-%{random_suffix}"

  labels = {
    label = "my-label"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret-version-basic" {
  secret = google_secret_manager_secret.secret-basic.name

  secret_data = "my-tf-test-secret%{random_suffix}"
  enabled = false
}
`, context)
}
