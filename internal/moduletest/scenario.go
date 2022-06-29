package moduletest

import (
	"fmt"

	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/plans"
)

// Scenario is the topmost level in the organizational tree for tests,
// with each instance representing one of the directories under the "tests"
// directory.
type Scenario struct {
	// Path is the relative path from the directory containing the module
	// under test to the directory containing the test scenario, always
	// using forward slashes even on Windows.
	//
	// Due to the current directory layout convention, Path always begins
	// with the literal prefix "tests/", but we include that prefix to
	// make room for additional extension in future, such as "examples/" if
	// we add support for running examples as additional tests in future.
	Path string

	// Steps are the steps of the scenario. In today's implementation a
	// scenario always has exactly the same steps generated automatically
	// to run the fixed apply+destroy sequence, but we expect to
	// allow some amount of customization of steps in future via an explicit
	// scenario configuration file, so that e.g. authors can test the
	// handling of gradual updates to existing infrastructure over multiple
	// steps.
	Steps []*Step
}

func (s *Scenario) Addr() addrs.ModuleTestScenario {
	return addrs.ModuleTestScenario{Path: s.Path}
}

// CleanupStep returns the mandatory final cleanup step for the recieving
// Scenario.
func (s *Scenario) CleanupStep() *Step {
	// The cleanup step should always be the last element of Steps and should
	// always use the destroy planning mode.
	if len(s.Steps) < 1 {
		panic("scenario does not have a cleanup step")
	}
	step := s.Steps[len(s.Steps)-1]
	if got, want := step.PlanMode, plans.DestroyMode; got != want {
		// This should never happen, but we'll check here just to give good
		// feedback if future changes to LoadScenarios cease upholding this
		// invariant.
		panic(fmt.Sprintf("scenario's cleanup step is %s, not %s", got, want))
	}
	return step
}
