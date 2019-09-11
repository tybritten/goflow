package actions

import (
	"net/http"
	"strings"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"

	"github.com/pkg/errors"
	"golang.org/x/net/http/httpguts"
)

func init() {
	registerType(TypeCallWebhook, func() flows.Action { return &CallWebhookAction{} })
}

// TypeCallWebhook is the type for the call webhook action
const TypeCallWebhook string = "call_webhook"

// CallWebhookAction can be used to call an external service. The body, header and url fields may be
// templates and will be evaluated at runtime. A [event:webhook_called] event will be created based on
// the results of the HTTP call. If this action has a `result_name`, then addtionally it will create
// a new result with that name. If the webhook returned valid JSON, that will be accessible
// through `extra` on the result.
//
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "call_webhook",
//     "method": "GET",
//     "url": "http://localhost/?cmd=success",
//     "headers": {
//       "Authorization": "Token AAFFZZHH"
//     },
//     "result_name": "webhook"
//   }
//
// @action call_webhook
type CallWebhookAction struct {
	baseAction
	onlineAction

	Method     string            `json:"method" validate:"required,http_method"`
	URL        string            `json:"url" validate:"required" engine:"evaluated"`
	Headers    map[string]string `json:"headers,omitempty" engine:"evaluated"`
	Body       string            `json:"body,omitempty" engine:"evaluated"`
	ResultName string            `json:"result_name,omitempty"`
}

// NewCallWebhook creates a new call webhook action
func NewCallWebhook(uuid flows.ActionUUID, method string, url string, headers map[string]string, body string, resultName string) *CallWebhookAction {
	return &CallWebhookAction{
		baseAction: newBaseAction(TypeCallWebhook, uuid),
		Method:     method,
		URL:        url,
		Headers:    headers,
		Body:       body,
		ResultName: resultName,
	}
}

// Validate validates our action is valid
func (a *CallWebhookAction) Validate() error {
	if a.Body != "" && a.Method == "GET" {
		return errors.Errorf("can't specify body if method is GET")
	}

	for key := range a.Headers {
		if !httpguts.ValidHeaderFieldName(key) {
			return errors.Errorf("header '%s' is not a valid HTTP header", key)
		}
	}

	return nil
}

// Execute runs this action
func (a *CallWebhookAction) Execute(run flows.FlowRun, step flows.Step, logModifier flows.ModifierCallback, logEvent flows.EventCallback) error {

	// substitute any variables in our url
	url, err := run.EvaluateTemplate(a.URL)
	if err != nil {
		logEvent(events.NewError(err))
	}
	if url == "" {
		logEvent(events.NewErrorf("call_webhook URL evaluated to empty string, skipping"))
		return nil
	}

	method := strings.ToUpper(a.Method)
	body := a.Body

	// substitute any body variables
	if body != "" {
		body, err = run.EvaluateTemplate(body)
		if err != nil {
			logEvent(events.NewError(err))
		}
	}

	// build our request
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return err
	}

	// add the custom headers, substituting any template vars
	for key, value := range a.Headers {
		headerValue, err := run.EvaluateTemplate(value)
		if err != nil {
			logEvent(events.NewError(err))
		}

		req.Header.Add(key, headerValue)
	}

	webhookSvc := run.Session().Engine().Services().Webhook(run.Session())

	call, err := webhookSvc.Call(req, "")

	if err != nil {
		logEvent(events.NewError(err))
	} else {
		logEvent(events.NewWebhookCalled(call))
		if a.ResultName != "" {
			a.saveWebhookResult(run, step, a.ResultName, call, logEvent)
		}
	}

	return nil
}

// Results enumerates any results generated by this flow object
func (a *CallWebhookAction) Results(node flows.Node, include func(*flows.ResultInfo)) {
	if a.ResultName != "" {
		include(flows.NewResultInfo(a.ResultName, webhookCategories, node))
	}
}
