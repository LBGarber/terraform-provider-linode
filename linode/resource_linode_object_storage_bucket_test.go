package linode

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/linode/linodego"
)

const (
	objAccessKeyEnvVar = "LINODE_OBJ_ACCESS_KEY"
	objSecretKeyEnvVar = "LINODE_OBJ_SECRET_KEY"
)

func init() {
	resource.AddTestSweepers("linode_object_storage_bucket", &resource.Sweeper{
		Name: "linode_object_storage_bucket",
		F:    testSweepLinodeObjectStorageBucket,
	})
}

func generateTestCert(domain string) (certificate, privateKey string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key: %s", err)
	}
	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment

	validFrom := time.Now()
	validUntil := validFrom.Add(time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Linode"},
		},
		NotBefore:             validFrom,
		NotAfter:              validUntil,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %s", err)
	}
	certBuffer := new(bytes.Buffer)
	if err := pem.Encode(certBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", fmt.Errorf("failed to encode certificate to PEM: %s", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %s", err)
	}
	keyBuffer := new(bytes.Buffer)
	if err := pem.Encode(keyBuffer, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return "", "", fmt.Errorf("failed to encode private key to PEM: %s", err)
	}

	return string(certBuffer.Bytes()), string(keyBuffer.Bytes()), nil
}

func testSweepLinodeObjectStorageBucket(prefix string) error {
	client, err := getClientForSweepers()
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}

	listOpts := sweeperListOptions(prefix, "label")
	objectStorageBuckets, err := client.ListObjectStorageBuckets(context.Background(), listOpts)
	if err != nil {
		return fmt.Errorf("Error getting object_storage_buckets: %s", err)
	}

	accessKey, accessKeyOk := os.LookupEnv(objAccessKeyEnvVar)
	secretKey, secretKeyOk := os.LookupEnv(objSecretKeyEnvVar)
	haveBucketAccess := accessKeyOk && secretKeyOk

	for _, objectStorageBucket := range objectStorageBuckets {
		if !shouldSweepAcceptanceTestResource(prefix, objectStorageBucket.Label) {
			continue
		}
		bucket := objectStorageBucket.Label

		if haveBucketAccess {
			conn := s3.New(session.New(&aws.Config{
				Region:      aws.String("us-east-1"),
				Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
				Endpoint:    aws.String(fmt.Sprintf(linodeObjectsEndpoint, objectStorageBucket.Cluster)),
			}))
			iter := s3manager.NewDeleteListIterator(conn, &s3.ListObjectsInput{
				Bucket: aws.String(bucket),
			})
			if err := s3manager.NewBatchDeleteWithClient(conn).Delete(aws.BackgroundContext(), iter); err != nil {
				return fmt.Errorf("unable to delete objects from bucket (%s): %s", bucket, err)
			}
		}

		err := client.DeleteObjectStorageBucket(context.Background(), objectStorageBucket.Cluster, bucket)

		if err != nil {
			if apiErr, ok := err.(*linodego.Error); ok && !haveBucketAccess && strings.HasPrefix(
				apiErr.Message, fmt.Sprintf("Bucket %s is not empty", bucket)) {
				log.Printf("[WARN] will not delete Object Storage Bucket (%s) as it needs to be emptied; "+
					"specify %q and %q env variables for bucket access", bucket, objAccessKeyEnvVar, objSecretKeyEnvVar)
				continue
			}
			return fmt.Errorf("Error destroying %s during sweep: %s", bucket, err)
		}
	}

	return nil
}

func TestAccLinodeObjectStorageBucket_basic(t *testing.T) {
	t.Parallel()

	resName := "linode_object_storage_bucket.foobar"
	var objectStorageBucketName = acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigBasic(t, objectStorageBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLinodeObjectStorageBucket_access(t *testing.T) {
	t.Parallel()

	resName := "linode_object_storage_bucket.foobar"
	objectStorageBucketName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithAccess(t, objectStorageBucketName, "public-read", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
					resource.TestCheckResourceAttr(resName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resName, "cors_enabled", "true"),
				),
			},
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithAccess(t, objectStorageBucketName, "private", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
					resource.TestCheckResourceAttr(resName, "acl", "private"),
					resource.TestCheckResourceAttr(resName, "cors_enabled", "false"),
				),
			},
		},
	})
}

func TestAccLinodeObjectStorageBucket_versioning(t *testing.T) {
	t.Parallel()

	resName := "linode_object_storage_bucket.foobar"
	objectStorageBucketName := acctest.RandomWithPrefix("tf-test")
	objectStorageKeyName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithVersioning(t, objectStorageBucketName, objectStorageKeyName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
					resource.TestCheckResourceAttr(resName, "versioning", "true"),
				),
			},
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithVersioning(t, objectStorageBucketName, objectStorageKeyName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
					resource.TestCheckResourceAttr(resName, "versioning", "false"),
				),
			},
		},
	})
}

func TestAccLinodeObjectStorageBucket_lifecycle(t *testing.T) {
	t.Parallel()

	resName := "linode_object_storage_bucket.foobar"
	objectStorageBucketName := acctest.RandomWithPrefix("tf-test")
	objectStorageKeyName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithLifecycle(t, objectStorageBucketName, objectStorageKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
					resource.TestCheckResourceAttr(resName, "cluster", "us-east-1"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.#", "1"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.id", "test-rule"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.prefix", "tf"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.abort_incomplete_multipart_upload_days", "5"),
					resource.TestCheckResourceAttrSet(resName, "lifecycle_rule.0.expiration.0.date"),
				),
			},
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithLifecycleUpdates(t, objectStorageBucketName, objectStorageKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
					resource.TestCheckResourceAttr(resName, "cluster", "us-east-1"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.#", "1"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.id", "test-rule-update"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.prefix", "tf-update"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.enabled", "false"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.abort_incomplete_multipart_upload_days", "42"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resName, "lifecycle_rule.0.expiration.0.days", "37"),
				),
			},
		},
	})
}

func TestAccLinodeObjectStorageBucket_cert(t *testing.T) {
	t.Parallel()

	resName := "linode_object_storage_bucket.foobar"
	objectStorageBucketName := acctest.RandomWithPrefix("tf-test") + ".io"
	cert, key, err := generateTestCert(objectStorageBucketName)
	if err != nil {
		t.Fatal(err)
	}

	invalidCert, invalidKey, err := generateTestCert("bogusdomain.com")
	if err != nil {
		t.Fatal(err)
	}

	otherCert, otherKey, err := generateTestCert(objectStorageBucketName)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithCert(t, objectStorageBucketName, cert, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					testAccCheckLinodeObjectStorageBucketHasSSL(true),
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
				),
			},
			{
				Config:      testAccCheckLinodeObjectStorageBucketConfigWithCert(t, objectStorageBucketName, invalidCert, invalidKey),
				ExpectError: regexp.MustCompile("failed to upload new bucket cert"),
			},
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigWithCert(t, objectStorageBucketName, otherCert, otherKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					testAccCheckLinodeObjectStorageBucketHasSSL(true),
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
				),
			},
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigBasic(t, objectStorageBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					testAccCheckLinodeObjectStorageBucketHasSSL(false),
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
				),
			},
		},
	})
}

func TestAccLinodeObjectStorageBucket_dataSource(t *testing.T) {
	t.Parallel()

	resName := "linode_object_storage_bucket.foobar"
	var objectStorageBucketName = acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigDataSource(t, objectStorageBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLinodeObjectStorageBucket_update(t *testing.T) {
	t.Parallel()

	var objectStorageBucketName = acctest.RandomWithPrefix("tf-test")
	resName := "linode_object_storage_bucket.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigBasic(t, objectStorageBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", objectStorageBucketName),
				),
			},
			{
				Config: testAccCheckLinodeObjectStorageBucketConfigUpdates(t, objectStorageBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeObjectStorageBucketExists,
					resource.TestCheckResourceAttr(resName, "label", fmt.Sprintf("%s-renamed", objectStorageBucketName)),
				),
			},
		},
	})
}

func testAccCheckLinodeObjectStorageBucketExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_object_storage_bucket" {
			continue
		}

		cluster, label, err := decodeLinodeObjectStorageBucketID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing %s, %s", rs.Primary.ID, err)
		}

		_, err = client.GetObjectStorageBucket(context.Background(), cluster, label)
		if err != nil {
			return fmt.Errorf("Error retrieving state of ObjectStorageBucket %s: %s", rs.Primary.Attributes["label"], err)
		}
	}

	return nil
}

func testAccCheckLinodeObjectStorageBucketHasSSL(expected bool) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ProviderMeta).Client
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "linode_object_storage_bucket" {
				continue
			}

			cluster, label, err := decodeLinodeObjectStorageBucketID(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("could not parse bucket ID %s: %s", rs.Primary.ID, err)
			}

			cert, err := client.GetObjectStorageBucketCert(context.TODO(), cluster, label)
			if err != nil {
				return fmt.Errorf("failed to get bucket cert: %s", err)
			}

			if cert.SSL != expected {
				return fmt.Errorf("expected cert.SSL to be %v; got %v", expected, cert.SSL)
			}
		}
		return nil
	}
}

func testAccCheckLinodeObjectStorageBucketDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_object_storage_bucket" {
			continue
		}

		id := rs.Primary.ID
		cluster, label, err := decodeLinodeObjectStorageBucketID(id)
		if err != nil {
			return fmt.Errorf("Error parsing %s", id)
		}
		if label == "" {
			return fmt.Errorf("Would have considered %s as %s", id, label)

		}

		_, err = client.GetObjectStorageBucket(context.Background(), cluster, label)

		if err == nil {
			return fmt.Errorf("Linode ObjectStorageBucket with id %s still exists", id)
		}

		if apiErr, ok := err.(*linodego.Error); ok && apiErr.Code != 404 {
			return fmt.Errorf("Error requesting Linode ObjectStorageBucket with id %s", id)
		}
	}

	return nil
}

type ObjectStorageBucketTemplateData struct {
	Label       string
	ACL         string
	CorsEnabled bool
	Versioning  bool
	SSLCert     string
	SSLKey      string
	Key         ObjectStorageKeyTemplateData
}

func testAccCheckLinodeObjectStorageBucketConfigBasic(t *testing.T, object_storage_bucket string) string {
	return testAccExecuteTemplate(t, "object_storage_bucket_basic",
		ObjectStorageBucketTemplateData{
			Label: object_storage_bucket,
		})
}

func testAccCheckLinodeObjectStorageBucketConfigWithAccess(t *testing.T, object_storage_bucket, acl string, cors bool) string {
	return testAccExecuteTemplate(t, "object_storage_bucket_access",
		ObjectStorageBucketTemplateData{
			Label:       object_storage_bucket,
			ACL:         acl,
			CorsEnabled: cors,
		})
}

func testAccCheckLinodeObjectStorageBucketConfigWithCert(t *testing.T, object_storage_bucket, cert, key string) string {
	return testAccExecuteTemplate(t, "object_storage_bucket_cert",
		ObjectStorageBucketTemplateData{
			Label:   object_storage_bucket,
			SSLCert: cert,
			SSLKey:  key,
		})
}

func testAccCheckLinodeObjectStorageBucketConfigWithVersioning(t *testing.T, bucketName, keyName string, versioning bool) string {
	return testAccExecuteTemplate(t, "object_storage_bucket_versioning",
		ObjectStorageBucketTemplateData{
			Label:      bucketName,
			Versioning: versioning,
			Key: ObjectStorageKeyTemplateData{
				Label: keyName,
			},
		})
}

func testAccCheckLinodeObjectStorageBucketConfigWithLifecycle(t *testing.T, bucketName, keyName string) string {
	return testAccExecuteTemplate(t, "object_storage_bucket_lifecycle",
		ObjectStorageBucketTemplateData{
			Label: bucketName,
			Key: ObjectStorageKeyTemplateData{
				Label: keyName,
			},
		})
}

func testAccCheckLinodeObjectStorageBucketConfigWithLifecycleUpdates(t *testing.T, bucketName, keyName string) string {
	return testAccExecuteTemplate(t, "object_storage_bucket_lifecycle_updates",
		ObjectStorageBucketTemplateData{
			Label: bucketName,
			Key: ObjectStorageKeyTemplateData{
				Label: keyName,
			},
		})
}

func testAccCheckLinodeObjectStorageBucketConfigUpdates(t *testing.T, object_storage_bucket string) string {
	return testAccExecuteTemplate(t, "object_storage_bucket_updates",
		ObjectStorageBucketTemplateData{
			Label: object_storage_bucket,
		})
}

func testAccCheckLinodeObjectStorageBucketConfigDataSource(t *testing.T, object_storage_bucket string) string {
	return testAccExecuteTemplate(t, "data_object_storage_bucket_basic",
		ObjectStorageBucketTemplateData{
			Label: object_storage_bucket,
		})
}
