// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

// exports used for testing only.
var (
	ResourceDataCellsFilter    = newResourceDataCellsFilter
	ResourceResourceLFTag      = newResourceResourceLFTag
	ResourceLakeFormationOptIn = newResourceLakeFormationOptIn

	FindDataCellsFilterByID    = findDataCellsFilterByID
	FindResourceLFTagByID      = findResourceLFTagByID
	FindLakeFormationOptInByID = findResourceLFTagByID
	LFTagParseResourceID       = lfTagParseResourceID

	ValidPrincipal = validPrincipal
)
