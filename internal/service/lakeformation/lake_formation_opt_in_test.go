// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (

	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/lakeformation/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	// "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	// "github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

func TestAccLakeFormationLakeFormationOptIn_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lake_formation_opt_in.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLakeFormationOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLakeFormationOptInConfig_basic(rName, "database"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLakeFormationOptInExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "database.0.name", "database"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", "data.aws_iam_session_context.current", "issuer_arn"),
				),
			},
			/* TODO
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccCatalogTableOptimizerStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			*/
		},
	})
}

func TestAccLakeFormationLakeFormationOptIn_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lake_formation_opt_in.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLakeFormationOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLakeFormationOptInConfig_basic(rName, "database"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLakeFormationOptInExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceLakeFormationOptIn, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLakeFormationLakeFormationOptIn_table(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lake_formation_opt_in.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLakeFormationOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLakeFormationOptInConfig_table(rName, "database", "table"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLakeFormationOptInExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "table.0.database_name", "database"),
					resource.TestCheckResourceAttr(resourceName, "table.0.name", "table"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", "data.aws_iam_session_context.current", "issuer_arn"),
				),
			},
			/* TODO
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccCatalogTableOptimizerStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			*/
		},
	})
}

func testAccCheckLakeFormationOptInDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_lake_formation_opt_in" {
				continue
			}

			input := &lakeformation.ListLakeFormationOptInsInput{
				Principal: &awstypes.DataLakePrincipal{},
				Resource:  &awstypes.Resource{},
			}

			if v, ok := rs.Primary.Attributes[names.AttrPrincipal]; ok {
				input.Principal.DataLakePrincipalIdentifier = aws.String(v)
			}

			// If Resource is a database
			if v, ok := rs.Primary.Attributes["database.0.name"]; ok {
				input.Resource.Database = &awstypes.DatabaseResource{
					Name: aws.String(v),
				}

				if v, ok := rs.Primary.Attributes["database.0.catalog_id"]; ok && len(v) > 1 {
					input.Resource.Database.CatalogId = aws.String(v)
				}
			}

			// If Resource is a table
			if v, ok := rs.Primary.Attributes["table.0.database_name"]; ok {
				input.Resource.Table = &awstypes.TableResource{
					DatabaseName: aws.String(v),
				}

				if v, ok := rs.Primary.Attributes["table.0.catalog_id"]; ok && len(v) > 1 {
					input.Resource.Table.CatalogId = aws.String(v)
				}

				if v, ok := rs.Primary.Attributes["table.0.name"]; ok {
					input.Resource.Table.Name = aws.String(v)
				}

				if v, ok := rs.Primary.Attributes["table.0.wildcard"]; ok && v == acctest.CtTrue {
					input.Resource.Table.TableWildcard = &awstypes.TableWildcard{}
				}
			}

			if _, err := tflakeformation.FindLFOptInByID(ctx, conn, input.Principal, input.Resource); err != nil {
				// Resource doesn't exist or requester doesn't have permission - the error does not distinguish
				if errs.IsA[*awstypes.AccessDeniedException](err) {
					return nil
				}
			}
		}
		return nil
	}
}

func testAccCheckLakeFormationOptInExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameLakeFormationOptIn, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameLakeFormationOptIn, name, errors.New("not set"))
		}

		input := &lakeformation.ListLakeFormationOptInsInput{
			Principal: &awstypes.DataLakePrincipal{},
			Resource:  &awstypes.Resource{},
		}

		if v, ok := rs.Primary.Attributes[names.AttrPrincipal]; ok {
			input.Principal.DataLakePrincipalIdentifier = aws.String(v)
		}

		// If Resource is a database
		if v, ok := rs.Primary.Attributes["database.0.name"]; ok {
			input.Resource.Database = &awstypes.DatabaseResource{
				Name: aws.String(v),
			}

			if v, ok := rs.Primary.Attributes["database.0.catalog_id"]; ok && len(v) > 1 {
				input.Resource.Database.CatalogId = aws.String(v)
			}
		}

		// If Resource is a table
		if v, ok := rs.Primary.Attributes["table.0.database_name"]; ok {
			input.Resource.Table = &awstypes.TableResource{
				DatabaseName: aws.String(v),
			}

			if v, ok := rs.Primary.Attributes["table.0.catalog_id"]; ok && len(v) > 1 {
				input.Resource.Table.CatalogId = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["table.0.name"]; ok {
				input.Resource.Table.Name = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["table.0.wildcard"]; ok && v == acctest.CtTrue {
				input.Resource.Table.TableWildcard = &awstypes.TableWildcard{}
			}
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		_, err := tflakeformation.FindLFOptInByID(ctx, conn, input.Principal, input.Resource)
		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameLakeFormationOptIn, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.ListLakeFormationOptInsInput{}

	_, err := conn.ListLakeFormationOptIns(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccLakeFormationOptInConfig_basic(rName, database string) string {
	/* TODO
	resource "aws_lakeformation_data_lake_settings" "test" {
		admins = [data.aws_iam_session_context.current.issuer_arn]
	}
	depends_on = [aws_lakeformation_data_lake_settings.test, aws_glue_catalog_database.test]

	*/
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_lakeformation_lake_formation_opt_in" "test" {
  principal = data.aws_iam_session_context.current.issuer_arn
 	database {
      name = "%[1]s"
  }
  depends_on = [aws_glue_catalog_database.test]
}
`, database)
}

func testAccLakeFormationOptInConfig_table(rName, database string, table string) string {
	/* TODO
	resource "aws_lakeformation_data_lake_settings" "test" {
		admins = [data.aws_iam_session_context.current.issuer_arn]
	}
	depends_on = [aws_lakeformation_data_lake_settings.test, aws_glue_catalog_database.test]

	*/
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[2]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

resource "aws_lakeformation_lake_formation_opt_in" "test" {
  principal = data.aws_iam_session_context.current.issuer_arn
 	table {
      database_name = "%[1]s"
      name = "%[2]s"
  }
  depends_on = [aws_glue_catalog_database.test]
}
`, database, table)
}
