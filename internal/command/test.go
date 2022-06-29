package command

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform/internal/command/arguments"
	"github.com/hashicorp/terraform/internal/command/views"
	"github.com/hashicorp/terraform/internal/configs"
	"github.com/hashicorp/terraform/internal/depsfile"
	"github.com/hashicorp/terraform/internal/moduletest"
	"github.com/hashicorp/terraform/internal/providercache"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

// TestCommand is the implementation of "terraform test".
type TestCommand struct {
	Meta
}

func (c *TestCommand) Run(rawArgs []string) int {
	// Parse and apply global view arguments
	common, rawArgs := arguments.ParseView(rawArgs)
	c.View.Configure(common)

	args, diags := arguments.ParseTest(rawArgs)
	view := views.NewTest(c.View, args.Output)
	if diags.HasErrors() {
		view.Diagnostics(diags)
		return 1
	}

	diags = diags.Append(tfdiags.Sourceless(
		tfdiags.Warning,
		`The "terraform test" command is experimental`,
		"We'd like to invite adventurous module authors to write integration tests for their modules using this command, but all of the behaviors of this command are currently experimental and may change based on feedback.\n\nFor more information on the testing experiment, including ongoing research goals and avenues for feedback, see:\n    https://www.terraform.io/docs/language/modules/testing-experiment.html",
	))

	ctx, cancel := c.InterruptibleContext()
	defer cancel()

	_, moreDiags := c.run(ctx, args)
	diags = diags.Append(moreDiags)

	initFailed := diags.HasErrors()
	view.Diagnostics(diags)
	diags = nil // view.Results(results)
	resultsFailed := diags.HasErrors()
	view.Diagnostics(diags) // possible additional errors from saving the results

	var testsFailed bool

	// TODO: Actually analyze the results and report out to the user.

	// Lots of things can possibly have failed
	if initFailed || resultsFailed || testsFailed {
		return 1
	}
	return 0
}

// TODO: Decide what type the results return value ought to have.
func (c *TestCommand) run(ctx context.Context, args arguments.Test) (results []struct{}, diags tfdiags.Diagnostics) {
	scenarios, diags := moduletest.LoadScenarios(".")
	if diags.HasErrors() {
		return nil, diags
	}

	for _, scenario := range scenarios {
		if ctx.Err() != nil {
			// If the context has already failed in some way then we'll
			// halt early and report whatever's already happened.
			// (Something downstream should have already emitted the error
			// into our diagnostics.)
			break
		}

		result, moreDiags := c.runScenario(ctx, scenario)
		diags = diags.Append(moreDiags)
		results = append(results, result)
	}

	return nil, diags
}

// TODO: Decide what type the result return value ought to have.
func (c *TestCommand) runScenario(ctx context.Context, scenario *moduletest.Scenario) (struct{}, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	log.Printf("[TRACE] TODO: Run %s", scenario.Addr().InitPseudoStep())

	// TODO: Actually implement this whole thing.
	for _, step := range scenario.Steps {
		log.Printf("[TRACE] TODO: Run %s", step.Addr())
	}

	return struct{}{}, diags
}

func (c *TestCommand) Help() string {
	helpText := `
Usage: terraform test [options]

  This is an experimental command to help with automated integration
  testing of shared modules. The usage and behavior of this command is
  likely to change in breaking ways in subsequent releases, as we
  are currently using this command primarily for research purposes.

  In its current experimental form, "test" will look under the current
  working directory for a subdirectory called "tests", and then within
  that directory search for one or more subdirectories that contain
  ".tf" or ".tf.json" files. For any that it finds, it will perform
  Terraform operations similar to the following sequence of commands
  in each of those directories:
      terraform validate
      terraform apply
      terraform destroy

  The test configurations should not declare any input variables and
  should at least contain a call to the module being tested, which
  will always be available at the path ../.. due to the expected
  filesystem layout.

  The tests are considered to be successful if all of the above steps
  succeed.

  Test configurations may optionally include uses of the special
  built-in test provider terraform.io/builtin/test, which allows
  writing explicit test assertions which must also all pass in order
  for the test run to be considered successful.

  This initial implementation is intended as a minimally-viable
  product to use for further research and experimentation, and in
  particular it currently lacks the following capabilities that we
  expect to consider in later iterations, based on feedback:
    - Testing of subsequent updates to existing infrastructure,
      where currently it only supports initial creation and
      then destruction.
    - Testing top-level modules that are intended to be used for
      "real" environments, which typically have hard-coded values
      that don't permit creating a separate "copy" for testing.
    - Some sort of support for unit test runs that don't interact
      with remote systems at all, e.g. for use in checking pull
      requests from untrusted contributors.

  In the meantime, we'd like to hear feedback from module authors
  who have tried writing some experimental tests for their modules
  about what sorts of tests you were able to write, what sorts of
  tests you weren't able to write, and any tests that you were
  able to write but that were difficult to model in some way.

Options:

  -compact-warnings  Use a more compact representation for warnings, if
                     this command produces only warnings and no errors.

  -junit-xml=FILE    In addition to the usual output, also write test
                     results to the given file path in JUnit XML format.
                     This format is commonly supported by CI systems, and
                     they typically expect to be given a filename to search
                     for in the test workspace after the test run finishes.

  -no-color          Don't include virtual terminal formatting sequences in
                     the output.
`
	return strings.TrimSpace(helpText)
}

func (c *TestCommand) Synopsis() string {
	return "Experimental support for module integration testing"
}

type testCommandSuiteDirs struct {
	SuiteName string

	ConfigDir    string
	ModulesDir   string
	ProvidersDir string

	Config        *configs.Config
	ProviderCache *providercache.Dir
	ProviderLocks *depsfile.Locks
}
