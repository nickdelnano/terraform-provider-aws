package s3_test

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(s3.EndpointsID, testAccErrorCheckSkip)
}

// testAccErrorCheckSkip skips tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Number of distinct destination bucket ARNs cannot exceed",
		"destination is not allowed",
	)
}

func TestAccS3Bucket_Basic_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	hostedZoneID, _ := tfs3.HostedZoneIDForRegion(region)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", hostedZoneID),
					resource.TestCheckResourceAttr(resourceName, "region", region),
					resource.TestCheckNoResourceAttr(resourceName, "website_endpoint"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "arn", "s3", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					testAccCheckS3BucketDomainName(resourceName, "bucket_domain_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_regional_domain_name", testAccBucketRegionalDomainName(bucketName, region)),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

// Support for common Terraform 0.11 pattern
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7868
func TestAccS3Bucket_Basic_emptyString(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketBucketEmptyStringConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^terraform-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Tags_withNoSystemTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketConfig_withUpdatedTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
				),
			},
			{
				Config: testAccBucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Verify update from 0 tags.
			{
				Config: testAccBucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Tags_withSystemTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	var stackID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:  acctest.Providers,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckBucketDestroy,
			func(s *terraform.State) error {
				// Tear down CF stack.
				conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

				requestToken := resource.UniqueId()
				req := &cloudformation.DeleteStackInput{
					StackName:          aws.String(stackID),
					ClientRequestToken: aws.String(requestToken),
				}

				log.Printf("[DEBUG] Deleting CloudFormation stack: %s", req)
				if _, err := conn.DeleteStack(req); err != nil {
					return fmt.Errorf("error deleting CloudFormation stack: %w", err)
				}

				if _, err := tfcloudformation.WaitStackDeleted(conn, stackID, requestToken, 10*time.Minute); err != nil {
					return fmt.Errorf("Error waiting for CloudFormation stack deletion: %s", err)
				}

				return nil
			},
		),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckDestroyBucket(resourceName),
					testAccCheckBucketCreateViaCloudFormation(bucketName, &stackID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckBucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{
				Config: testAccBucketConfig_withUpdatedTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
					testAccCheckBucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{
				Config: testAccBucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckBucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Tags_ignoreTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketConfig_withNoTags(bucketName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketUpdateTags(resourceName, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckBucketCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
					}),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketConfig_withTags(bucketName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckBucketCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
						"Key1":       "AAA",
						"Key2":       "BBB",
						"Key3":       "CCC",
					}),
				),
			},
		},
	})
}

func TestAccS3Bucket_Tags_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket.bucket1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiBucketWithTagsConfig(rInt),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_namePrefix(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "bucket_prefix"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_generatedName(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "bucket_prefix"},
			},
		},
	})
}

// Test TestAccS3Bucket_Basic_shouldFailNotFound is designed to fail with a "plan
// not empty" error in Terraform, to check against regressions.
// See https://github.com/hashicorp/terraform/pull/2925
func TestAccS3Bucket_Basic_shouldFailNotFound(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDestroyedConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckDestroyBucket(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Bucket_Manage_versioning(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithVersioningConfig(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketWithVersioningConfig(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_versioningDisabled(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithVersioningConfig(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_MfaDeleteDisabled(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithVersioningMfaDeleteConfig(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_versioningAndMfaDeleteDisabled(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithVersioningVersioningAndMfaDeleteConfig(bucketName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_lifecycleBasic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithLifecycleConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "false"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "30",
						"storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "60",
						"storage_class": "INTELLIGENT_TIERING",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "90",
						"storage_class": "ONEZONE_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "120",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "210",
						"storage_class": "DEEP_ARCHIVE",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.date", "2016-01-12"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.expired_object_delete_marker", "false"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.2.transition.*", map[string]string{
						"days": "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.id", "id4"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.prefix", "path4/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.terraform", "hashicorp"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.id", "id5"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.terraform", "hashicorp"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.4.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.id", "id6"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.tags.tagKey", "tagValue"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.5.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketWithVersioningLifecycleConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.noncurrent_version_expiration.0.days", "365"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.noncurrent_version_transition.*", map[string]string{
						"days":          "30",
						"storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.noncurrent_version_transition.*", map[string]string{
						"days":          "60",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.noncurrent_version_expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.2.noncurrent_version_transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
				),
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_lifecycleExpireMarkerOnly(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithLifecycleExpireMarkerConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11420
func TestAccS3Bucket_Manage_lifecycleRuleExpirationEmptyBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleRuleExpirationEmptyConfigurationBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15138
func TestAccS3Bucket_Manage_lifecycleRuleAbortIncompleteMultipartUploadDaysNoExpiration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleRuleAbortIncompleteMultipartUploadDaysConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLock(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLockEnabledNoDefaultRetention(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccObjectLockEnabledWithDefaultRetention(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.mode", "COMPLIANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.days", "3"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroy(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjects(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

// By default, the AWS Go SDK cleans up URIs by removing extra slashes
// when the service API requests use the URI as part of making a request.
// While the aws_s3_object resource automatically cleans the key
// to not contain these extra slashes, out-of-band handling and other AWS
// services may create keys with extra slashes (empty "directory" prefixes).
func TestAccS3Bucket_Basic_forceDestroyWithEmptyPrefixes(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjects(resourceName, "data.txt", "/extraleadingslash.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroyWithObjectLockEnabled(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroyWithObjectLockEnabled(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjectsWithLegalHold(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestBucketName(t *testing.T) {
	validDnsNames := []string{
		"foobar",
		"foo.bar",
		"foo.bar.baz",
		"1234",
		"foo-bar",
		strings.Repeat("x", 63),
	}

	for _, v := range validDnsNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsWest2RegionID); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidDnsNames := []string{
		"foo..bar",
		"Foo.Bar",
		"192.168.0.1",
		"127.0.0.1",
		".foo",
		"bar.",
		"foo_bar",
		strings.Repeat("x", 64),
	}

	for _, v := range invalidDnsNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsWest2RegionID); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}

	validEastNames := []string{
		"foobar",
		"foo_bar",
		"127.0.0.1",
		"foo..bar",
		"foo_bar_baz",
		"foo.bar.baz",
		"Foo.Bar",
		strings.Repeat("x", 255),
	}

	for _, v := range validEastNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsEast1RegionID); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidEastNames := []string{
		"foo;bar",
		strings.Repeat("x", 256),
	}

	for _, v := range invalidEastNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsEast1RegionID); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}
}

func TestBucketRegionalDomainName(t *testing.T) {
	const bucket = "bucket-name"

	var testCases = []struct {
		ExpectedErrCount int
		ExpectedOutput   string
		Region           string
	}{
		{
			Region:           "",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           "custom",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.custom.amazonaws.com",
		},
		{
			Region:           endpoints.UsEast1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           endpoints.UsWest2RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", endpoints.UsWest2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			Region:           endpoints.UsGovWest1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", endpoints.UsGovWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			Region:           endpoints.CnNorth1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.amazonaws.com.cn", endpoints.CnNorth1RegionID),
		},
	}

	for _, tc := range testCases {
		output, err := tfs3.BucketRegionalDomainName(bucket, tc.Region)
		if tc.ExpectedErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Region, err)
		}
		if tc.ExpectedErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Region)
		}
		if output != tc.ExpectedOutput {
			t.Fatalf("expected %q, received %q", tc.ExpectedOutput, output)
		}
	}
}

func TestWebsiteEndpoint(t *testing.T) {
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	testCases := []struct {
		TestingClient      *conns.AWSClient
		LocationConstraint string
		Expected           string
	}{
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsEast1RegionID,
			},
			LocationConstraint: "",
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsWest2RegionID,
			},
			LocationConstraint: endpoints.UsWest2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsWest2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsWest1RegionID,
			},
			LocationConstraint: endpoints.UsWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.EuWest1RegionID,
			},
			LocationConstraint: endpoints.EuWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.EuWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.EuWest3RegionID,
			},
			LocationConstraint: endpoints.EuWest3RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.EuWest3RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.EuCentral1RegionID,
			},
			LocationConstraint: endpoints.EuCentral1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.EuCentral1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApSouth1RegionID,
			},
			LocationConstraint: endpoints.ApSouth1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.ApSouth1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApSoutheast1RegionID,
			},
			LocationConstraint: endpoints.ApSoutheast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApSoutheast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApNortheast1RegionID,
			},
			LocationConstraint: endpoints.ApNortheast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApNortheast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApSoutheast2RegionID,
			},
			LocationConstraint: endpoints.ApSoutheast2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApSoutheast2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApNortheast2RegionID,
			},
			LocationConstraint: endpoints.ApNortheast2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.ApNortheast2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.SaEast1RegionID,
			},
			LocationConstraint: endpoints.SaEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.SaEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsGovEast1RegionID,
			},
			LocationConstraint: endpoints.UsGovEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.UsGovEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsGovWest1RegionID,
			},
			LocationConstraint: endpoints.UsGovWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsGovWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "c2s.ic.gov",
				Region:    endpoints.UsIsoEast1RegionID,
			},
			LocationConstraint: endpoints.UsIsoEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.c2s.ic.gov", endpoints.UsIsoEast1RegionID),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "sc2s.sgov.gov",
				Region:    endpoints.UsIsobEast1RegionID,
			},
			LocationConstraint: endpoints.UsIsobEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.sc2s.sgov.gov", endpoints.UsIsobEast1RegionID),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com.cn",
				Region:    endpoints.CnNorthwest1RegionID,
			},
			LocationConstraint: endpoints.CnNorthwest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.amazonaws.com.cn", endpoints.CnNorthwest1RegionID),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com.cn",
				Region:    endpoints.CnNorth1RegionID,
			},
			LocationConstraint: endpoints.CnNorth1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.amazonaws.com.cn", endpoints.CnNorth1RegionID),
		},
	}

	for _, testCase := range testCases {
		got := tfs3.WebsiteEndpoint(testCase.TestingClient, "bucket-name", testCase.LocationConstraint)
		if got.Endpoint != testCase.Expected {
			t.Errorf("WebsiteEndpointUrl(\"bucket-name\", %q) => %q, want %q", testCase.LocationConstraint, got.Endpoint, testCase.Expected)
		}
	}
}

func testAccCheckBucketDestroy(s *terraform.State) error {
	return testAccCheckBucketDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckBucketDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket" {
			continue
		}

		input := &s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		// Retry for S3 eventual consistency
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.HeadBucket(input)

			if tfawserr.ErrMessageContains(err, s3.ErrCodeNoSuchBucket, "") || tfawserr.ErrMessageContains(err, "NotFound", "") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("AWS S3 Bucket still exists: %s", rs.Primary.ID))
		})

		if tfresource.TimedOut(err) {
			_, err = conn.HeadBucket(input)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckBucketExists(n string) resource.TestCheckFunc {
	return testAccCheckBucketExistsWithProvider(n, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckBucketExistsWithProvider(n string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		provider := providerF()

		conn := provider.Meta().(*conns.AWSClient).S3Conn
		_, err := conn.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("S3 bucket not found")
			}
			return err
		}
		return nil

	}
}

func testAccCheckDestroyBucket(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
		_, err := conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error destroying Bucket (%s) in testAccCheckDestroyBucket: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckBucketAddObjects(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ConnURICleaningDisabled

		for _, key := range keys {
			_, err := conn.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(rs.Primary.ID),
				Key:    aws.String(key),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %s", err)
			}
		}

		return nil
	}
}

func testAccCheckBucketAddObjectsWithLegalHold(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		for _, key := range keys {
			_, err := conn.PutObject(&s3.PutObjectInput{
				Bucket:                    aws.String(rs.Primary.ID),
				Key:                       aws.String(key),
				ObjectLockLegalHoldStatus: aws.String(s3.ObjectLockLegalHoldStatusOn),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %s", err)
			}
		}

		return nil
	}
}

// Create an S3 bucket via a CF stack so that it has system tags.
func testAccCheckBucketCreateViaCloudFormation(n string, stackID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn
		stackName := sdkacctest.RandomWithPrefix("tf-acc-test-s3tags")
		templateBody := fmt.Sprintf(`{
  "Resources": {
    "TfTestBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "%s"
      }
    }
  }
}`, n)

		requestToken := resource.UniqueId()
		req := &cloudformation.CreateStackInput{
			StackName:          aws.String(stackName),
			TemplateBody:       aws.String(templateBody),
			ClientRequestToken: aws.String(requestToken),
		}

		log.Printf("[DEBUG] Creating CloudFormation stack: %s", req)
		resp, err := conn.CreateStack(req)
		if err != nil {
			return fmt.Errorf("error creating CloudFormation stack: %w", err)
		}

		stack, err := tfcloudformation.WaitStackCreated(conn, aws.StringValue(resp.StackId), requestToken, 10*time.Minute)
		if err != nil {
			return fmt.Errorf("Error waiting for CloudFormation stack creation: %w", err)
		}
		status := aws.StringValue(stack.StackStatus)
		if status != cloudformation.StackStatusCreateComplete {
			return fmt.Errorf("Invalid CloudFormation stack creation status: %s", status)
		}

		*stackID = aws.StringValue(resp.StackId)
		return nil
	}
}

func testAccCheckBucketTagKeys(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		got, err := tfs3.BucketListTags(conn, rs.Primary.Attributes["bucket"])
		if err != nil {
			return err
		}

		for _, want := range keys {
			ok := false
			for _, key := range got.Keys() {
				if want == key {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("Key %s not found in bucket's tag set", want)
			}
		}

		return nil
	}
}

func testAccCheckS3BucketDomainName(resourceName string, attributeName string, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue := acctest.Provider.Meta().(*conns.AWSClient).PartitionHostname(fmt.Sprintf("%s.s3", bucketName))

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccBucketRegionalDomainName(bucket, region string) string {
	regionalEndpoint, err := tfs3.BucketRegionalDomainName(bucket, region)
	if err != nil {
		return fmt.Sprintf("Regional endpoint not found for bucket %s", bucket)
	}
	return regionalEndpoint
}

func testAccCheckBucketUpdateTags(n string, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		return tfs3.BucketUpdateTags(conn, rs.Primary.Attributes["bucket"], oldTags, newTags)
	}
}

func testAccCheckBucketCheckTags(n string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		got, err := tfs3.BucketListTags(conn, rs.Primary.Attributes["bucket"])
		if err != nil {
			return err
		}

		want := tftags.New(expectedTags)
		if !reflect.DeepEqual(want, got) {
			return fmt.Errorf("Incorrect tags, want: %v got: %v", want, got)
		}

		return nil
	}
}

func testAccBucketConfig_Basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccBucketConfig_withNoTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  force_destroy = false

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
`, bucketName)
}

func testAccBucketConfig_withTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  force_destroy = false

  tags = {
    Key1 = "AAA"
    Key2 = "BBB"
    Key3 = "CCC"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
`, bucketName)
}

func testAccBucketConfig_withUpdatedTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  force_destroy = false

  tags = {
    Key2 = "BBB"
    Key3 = "XXX"
    Key4 = "DDD"
    Key5 = "EEE"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
`, bucketName)
}

func testAccMultiBucketWithTagsConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket1" {
  bucket        = "tf-test-bucket-1-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-1-%[1]d"
    Environment = "%[1]d"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test1" {
  bucket = aws_s3_bucket.bucket1.id
  acl    = "private"
}

resource "aws_s3_bucket" "bucket2" {
  bucket        = "tf-test-bucket-2-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-2-%[1]d"
    Environment = "%[1]d"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test2" {
  bucket = aws_s3_bucket.bucket2.id
  acl    = "private"
}

resource "aws_s3_bucket" "bucket3" {
  bucket        = "tf-test-bucket-3-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-3-%[1]d"
    Environment = "%[1]d"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test3" {
  bucket = aws_s3_bucket.bucket3.id
  acl    = "private"
}

resource "aws_s3_bucket" "bucket4" {
  bucket        = "tf-test-bucket-4-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-4-%[1]d"
    Environment = "%[1]d"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test4" {
  bucket = aws_s3_bucket.bucket4.id
  acl    = "private"
}

resource "aws_s3_bucket" "bucket5" {
  bucket        = "tf-test-bucket-5-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-5-%[1]d"
    Environment = "%[1]d"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test5" {
  bucket = aws_s3_bucket.bucket5.id
  acl    = "private"
}

resource "aws_s3_bucket" "bucket6" {
  bucket        = "tf-test-bucket-6-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-6-%[1]d"
    Environment = "%[1]d"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test6" {
  bucket = aws_s3_bucket.bucket6.id
  acl    = "private"
}
`, randInt)
}

func testAccBucketDestroyedConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "public-read"
}
`, bucketName)
}

func testAccBucketWithVersioningConfig(bucketName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    enabled = %[2]t
  }
}
`, bucketName, enabled)
}

func testAccBucketWithVersioningMfaDeleteConfig(bucketName string, mfaDelete bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    mfa_delete = %[2]t
  }
}
`, bucketName, mfaDelete)
}

func testAccBucketWithVersioningVersioningAndMfaDeleteConfig(bucketName string, enabled, mfaDelete bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    enabled    = %[2]t
    mfa_delete = %[3]t
  }
}
`, bucketName, enabled, mfaDelete)
}

func testAccBucketWithLifecycleConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      days = 365
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "INTELLIGENT_TIERING"
    }

    transition {
      days          = 90
      storage_class = "ONEZONE_IA"
    }

    transition {
      days          = 120
      storage_class = "GLACIER"
    }

    transition {
      days          = 210
      storage_class = "DEEP_ARCHIVE"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = true

    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id4"
    prefix  = "path4/"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id5"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id6"
    enabled = true

    tags = {
      "tagKey" = "tagValue"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
`, bucketName)
}

func testAccBucketWithLifecycleExpireMarkerConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      expired_object_delete_marker = "true"
    }
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
`, bucketName)
}

func testAccBucketWithVersioningLifecycleConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    enabled = false
  }

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    noncurrent_version_expiration {
      days = 365
    }

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = false

    noncurrent_version_expiration {
      days = 365
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true

    noncurrent_version_transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
`, bucketName)
}

func testAccBucketLifecycleRuleExpirationEmptyConfigurationBlockConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    enabled = true
    id      = "id1"

    expiration {}
  }
}
`, rName)
}

func testAccBucketLifecycleRuleAbortIncompleteMultipartUploadDaysConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    abort_incomplete_multipart_upload_days = 7
    enabled                                = true
    id                                     = "id1"
  }
}
`, rName)
}

func testAccObjectLockEnabledNoDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}
`, bucketName)
}

func testAccObjectLockEnabledWithDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 3
      }
    }
  }
}
`, bucketName)
}

func testAccBucketConfig_forceDestroy(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  force_destroy = true

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "public-read"
}
`, bucketName)
}

func testAccBucketConfig_forceDestroyWithObjectLockEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  force_destroy = true

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }

  lifecycle {
    ignore_changes = [
      grant,
    ]
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
`, bucketName)
}

const testAccBucketBucketEmptyStringConfig = `
resource "aws_s3_bucket" "test" {
  bucket = ""
}
`

const testAccBucketConfig_namePrefix = `
resource "aws_s3_bucket" "test" {
  bucket_prefix = "tf-test-"
}
`

const testAccBucketConfig_generatedName = `
resource "aws_s3_bucket" "test" {
  bucket_prefix = "tf-test-"
}
`
