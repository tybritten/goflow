package runs

import (
	"fmt"

	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/flows"
)

// RunContextTopLevels are the allowed top-level variables for expression evaluations
var RunContextTopLevels = []string{"contact", "child", "parent", "run", "trigger"}

type runContext struct {
	run flows.FlowRun
}

// creates a new evaluation context for the passed in run
func newRunContext(run flows.FlowRun) types.XResolvable {
	return &runContext{run: run}
}

// Resolve resolves the given top-level key in an expression
func (c *runContext) Resolve(key string) types.XValue {
	switch key {
	case "contact":
		return c.run.Contact()
	case "child":
		return newRelatedRunContext(c.run.Session().GetCurrentChild(c.run))
	case "parent":
		return newRelatedRunContext(c.run.Parent())
	case "run":
		return c.run
	case "trigger":
		return c.run.Session().Trigger()
	}

	return fmt.Errorf("no field '%s' on context", key)
}

// wraps parent/child runs and provides a reduced set of keys in the context
type relatedRunContext struct {
	run flows.RunSummary
}

// creates a new context for related runs
func newRelatedRunContext(run flows.RunSummary) *relatedRunContext {
	if run != nil {
		return &relatedRunContext{run: run}
	}
	return nil
}

// Resolve resolves the given key when this related run is referenced in an expression
func (c *relatedRunContext) Resolve(key string) types.XValue {
	switch key {
	case "uuid":
		return string(c.run.UUID())
	case "contact":
		return c.run.Contact()
	case "flow":
		return c.run.Flow()
	case "status":
		return string(c.run.Status())
	case "results":
		return c.run.Results()
	}

	return fmt.Errorf("no field '%s' on related run", key)
}

func (c *relatedRunContext) String() string {
	return c.run.UUID().String()
}

var _ types.XResolvable = (*runContext)(nil)
var _ types.XResolvable = (*relatedRunContext)(nil)
