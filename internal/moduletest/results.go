package moduletest

import (
	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/checks"
	"github.com/hashicorp/terraform/internal/states"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

type Results addrs.Map[addrs.ModuleTestCase, *TestCaseResult]

type TestCaseResult struct {
	// AggregateStatus is the status representing the outcome of the entire
	// test case.
	//
	// If AggregateStatus is checks.StatusError then there will always be at
	// least one error in field Diagnostics, and conversely for any other
	// status we guarantee that there are no error diagnostics.
	AggregateStatus checks.Status

	// ObjectResults gives the statuses of each of the individual objects
	// belonging to this test case, if any.
	//
	// If this map is empty then the meaning depends on AggregateStatus:
	//   - If Passed, this test step didn't include any objects for this test case.
	//   - If Unknown or Error, this test step didn't get a chance to expand at
	//     all because it was blocked by an error.
	ObjectResults addrs.Map[addrs.Checkable, *states.CheckResultObject]

	// Diagnostics is the subset of diagnostics that were explicitly associated
	// with this test case by Terraform Core.
	//
	// Not all diagnostics are annotated with information that allows us to
	// determine which test case they relate to (if any), so any diagnostics
	// unaccounted for by a test result should be returned separately by the
	// test runner as its own diagnostics.
	//
	// Diagnostics does not include any diagnostics that would be redundant
	// with failure messages included in the ObjectResults values, so we can
	// always return actual failures against the specific dynamic object that
	// failed.
	Diagnostics tfdiags.Diagnostics
}
