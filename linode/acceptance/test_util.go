package acceptance

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode"
	"github.com/linode/terraform-provider-linode/linode/helper"
	"golang.org/x/crypto/ssh"
)

const optInTestsEnvVar = "ACC_OPT_IN_TESTS"
const SkipInstanceReadyPollKey = "skip_instance_ready_poll"

var (
	optInTests         map[string]struct{}
	privateKeyMaterial string
	PublicKeyMaterial  string
	TestAccProviders   map[string]*schema.Provider
	TestAccProvider    *schema.Provider
	ConfigTemplates    *template.Template
)

func init() {
	var err error
	PublicKeyMaterial, privateKeyMaterial, err = acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		log.Fatalf("Failed to generate random SSH key pair for testing: %s", err)
	}
	// optInTests = make(map[string]struct{})
	// optInTestsValue, ok := os.LookupEnv(optInTestsEnvVar)
	// if !ok {
	// 	return
	// }

	// for _, testName := range strings.Split(optInTestsValue, ",") {
	// 	optInTests[testName] = struct{}{}
	// }
	TestAccProvider = linode.Provider()
	TestAccProviders = map[string]*schema.Provider{
		"linode": TestAccProvider,
	}

	var templateFiles []string

	err = filepath.Walk("../", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".gotf" {
			templateFiles = append(templateFiles, path)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("failed to load template files: %v", err)
	}

	ConfigTemplates = template.New("tf-test")
	if _, err := ConfigTemplates.ParseFiles(templateFiles...); err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}
}

func TestProvider(t *testing.T) {
	if err := linode.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("LINODE_TOKEN"); v == "" {
		t.Fatal("LINODE_TOKEN must be set for acceptance tests")
	}
}

func AccTestWithProvider(config string, options map[string]interface{}) string {
	sb := strings.Builder{}
	sb.WriteString("provider \"linode\" {\n")
	for key, value := range options {
		sb.WriteString(fmt.Sprintf("\t%s = %#v\n", key, value))
	}
	sb.WriteString("}\n")
	return sb.String() + config
}

func OptInTest(t *testing.T) {
	t.Helper()

	if _, ok := optInTests[t.Name()]; !ok {
		t.Skipf("skipping opt-in test; specify test in environment variable %q to run", optInTestsEnvVar)
	}
}

func GetSSHClient(t *testing.T, user, addr string) (client *ssh.Client) {
	t.Helper()

	signer, err := ssh.ParsePrivateKey([]byte(privateKeyMaterial))
	if err != nil {
		t.Fatalf("failed to parse private key: %s", err)
	}
	config := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Minute,
	}

	attempts := 3

	for attempts != 0 {
		client, err = ssh.Dial("tcp", addr+":22", config)
		if err == nil {
			break
		}

		t.Logf("ssh dial failed: %s", err)
		attempts--

		time.Sleep(5 * time.Second)
	}

	if client == nil {
		t.Fatal("failed to get ssh client")
	}
	return
}

func TestAccCheckResourceAttrNotEqual(resName string, path, notValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resName]
		if !ok {
			return fmt.Errorf("Not found: %s", resName)
		}

		if value, ok := rs.Primary.Attributes[path]; !ok {
			return fmt.Errorf("attribute %s does not exist", path)
		} else if value == notValue {
			return fmt.Errorf("attribute was equal")
		}

		return nil
	}
}

func TestAccCheckResourceNonEmptyList(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		instCount, err := strconv.Atoi(rs.Primary.Attributes[fmt.Sprintf("%s.#", attrName)])
		if err != nil {
			return fmt.Errorf("failed to parse: %s", err)
		}

		if instCount < 1 {
			return fmt.Errorf("expected at least 1 element in %s", attrName)
		}

		return nil
	}
}

func CheckLKEClusterDestroy(s *terraform.State) error {
	client := TestAccProvider.Meta().(*helper.ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_lke_cluster" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse LKE Cluster ID: %s", err)
		}

		if id == 0 {
			return fmt.Errorf("should not have LKE Cluster ID of 0")
		}

		if _, err = client.GetLKECluster(context.Background(), id); err == nil {
			return fmt.Errorf("should not find Linode ID %d existing after delete", id)
		} else if apiErr, ok := err.(*linodego.Error); !ok {
			return fmt.Errorf("expected API Error but got %#v", err)
		} else if apiErr.Code != 404 {
			return fmt.Errorf("expected an error 404 but got %#v", apiErr)
		}
	}

	return nil
}

func ExecuteTemplate(t *testing.T, templateName string, data interface{}) string {
	var b bytes.Buffer

	err := ConfigTemplates.ExecuteTemplate(&b, templateName, data)
	if err != nil {
		t.Fatalf("failed to execute template %s: %v", templateName, err)
	}

	return b.String()
}
