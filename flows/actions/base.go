package actions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/envs"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/utils"
	"github.com/nyaruka/goflow/utils/dates"
	"github.com/nyaruka/goflow/utils/uuids"

	"github.com/pkg/errors"
)

const (
	// common category names
	CategorySuccess = "Success"
	CategorySkipped = "Skipped"
	CategoryFailure = "Failure"
)

var webhookCategories = []string{CategorySuccess, CategoryFailure}
var webhookStatusCategories = map[flows.CallStatus]string{
	flows.CallStatusSuccess:         CategorySuccess,
	flows.CallStatusResponseError:   CategoryFailure,
	flows.CallStatusConnectionError: CategoryFailure,
	flows.CallStatusSubscriberGone:  CategoryFailure,
}

var registeredTypes = map[string](func() flows.Action){}

// registers a new type of action
func registerType(name string, initFunc func() flows.Action) {
	registeredTypes[name] = initFunc
}

// RegisteredTypes gets the registered types of action
func RegisteredTypes() map[string](func() flows.Action) {
	return registeredTypes
}

var uuidRegex = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// the base of all action types
type baseAction struct {
	Type_ string           `json:"type" validate:"required"`
	UUID_ flows.ActionUUID `json:"uuid" validate:"required,uuid4"`
}

// creates a new base action
func newBaseAction(typeName string, uuid flows.ActionUUID) baseAction {
	return baseAction{Type_: typeName, UUID_: uuid}
}

// Type returns the type of this action
func (a *baseAction) Type() string { return a.Type_ }

// UUID returns the UUID of the action
func (a *baseAction) UUID() flows.ActionUUID { return a.UUID_ }

// Validate validates our action is valid
func (a *baseAction) Validate() error { return nil }

// LocalizationUUID gets the UUID which identifies this object for localization
func (a *baseAction) LocalizationUUID() uuids.UUID { return uuids.UUID(a.UUID_) }

// helper function for actions that send a message (text + attachments) that must be localized and evalulated
func (a *baseAction) evaluateMessage(run flows.FlowRun, languages []envs.Language, actionText string, actionAttachments []string, actionQuickReplies []string, logEvent flows.EventCallback) (string, []utils.Attachment, []string) {
	// localize and evaluate the message text
	localizedText := run.GetTranslatedTextArray(uuids.UUID(a.UUID()), "text", []string{actionText}, languages)[0]
	evaluatedText, err := run.EvaluateTemplate(localizedText)
	if err != nil {
		logEvent(events.NewError(err))
	}

	// localize and evaluate the message attachments
	translatedAttachments := run.GetTranslatedTextArray(uuids.UUID(a.UUID()), "attachments", actionAttachments, languages)
	evaluatedAttachments := make([]utils.Attachment, 0, len(translatedAttachments))
	for _, a := range translatedAttachments {
		evaluatedAttachment, err := run.EvaluateTemplate(a)
		if err != nil {
			logEvent(events.NewError(err))
		}
		if evaluatedAttachment == "" {
			logEvent(events.NewErrorf("attachment text evaluated to empty string, skipping"))
			continue
		}
		evaluatedAttachments = append(evaluatedAttachments, utils.Attachment(evaluatedAttachment))
	}

	// localize and evaluate the quick replies
	translatedQuickReplies := run.GetTranslatedTextArray(uuids.UUID(a.UUID()), "quick_replies", actionQuickReplies, languages)
	evaluatedQuickReplies := make([]string, 0, len(translatedQuickReplies))
	for _, qr := range translatedQuickReplies {
		evaluatedQuickReply, err := run.EvaluateTemplate(qr)
		if err != nil {
			logEvent(events.NewError(err))
		}
		if evaluatedQuickReply == "" {
			logEvent(events.NewErrorf("quick reply text evaluated to empty string, skipping"))
			continue
		}
		evaluatedQuickReplies = append(evaluatedQuickReplies, evaluatedQuickReply)
	}

	return evaluatedText, evaluatedAttachments, evaluatedQuickReplies
}

// helper to save a run result and log it as an event
func (a *baseAction) saveResult(run flows.FlowRun, step flows.Step, name, value, category, categoryLocalized string, input string, extra json.RawMessage, logEvent flows.EventCallback) {
	result := flows.NewResult(name, value, category, categoryLocalized, step.NodeUUID(), input, extra, dates.Now())
	run.SaveResult(result)
	logEvent(events.NewRunResultChanged(result))
}

// helper to save a run result based on a webhook call and log it as an event
func (a *baseAction) saveWebhookResult(run flows.FlowRun, step flows.Step, name string, webhook *flows.WebhookCall, logEvent flows.EventCallback) {
	input := fmt.Sprintf("%s %s", webhook.Method, webhook.URL)
	value := strconv.Itoa(webhook.StatusCode)
	category := webhookStatusCategories[webhook.Status]
	extra := utils.ExtractResponseJSON(webhook.Response)

	a.saveResult(run, step, name, value, category, "", input, extra, logEvent)
}

// helper to apply a contact modifier
func (a *baseAction) applyModifier(run flows.FlowRun, mod flows.Modifier, logModifier flows.ModifierCallback, logEvent flows.EventCallback) {
	mod.Apply(run.Session().Environment(), run.Session().Assets(), run.Contact(), logEvent)
	logModifier(mod)
}

// helper to log a failure
func (a *baseAction) fail(run flows.FlowRun, err error, logEvent flows.EventCallback) {
	run.Exit(flows.RunStatusFailed)
	logEvent(events.NewFailure(err))
}

// utility struct which sets the allowed flow types to any
type universalAction struct{}

// AllowedFlowTypes returns the flow types which this action is allowed to occur in
func (a *universalAction) AllowedFlowTypes() []flows.FlowType {
	return []flows.FlowType{flows.FlowTypeMessaging, flows.FlowTypeMessagingOffline, flows.FlowTypeVoice}
}

// utility struct which sets the allowed flow types to any which run online
type onlineAction struct{}

// AllowedFlowTypes returns the flow types which this action is allowed to occur in
func (a *onlineAction) AllowedFlowTypes() []flows.FlowType {
	return []flows.FlowType{flows.FlowTypeMessaging, flows.FlowTypeVoice}
}

// utility struct which sets the allowed flow types to just voice
type voiceAction struct{}

// AllowedFlowTypes returns the flow types which this action is allowed to occur in
func (a *voiceAction) AllowedFlowTypes() []flows.FlowType {
	return []flows.FlowType{flows.FlowTypeVoice}
}

// utility struct for actions which operate on other contacts
type otherContactsAction struct {
	URNs         []urns.URN                `json:"urns,omitempty"`
	Groups       []*assets.GroupReference  `json:"groups,omitempty" validate:"dive"`
	Contacts     []*flows.ContactReference `json:"contacts,omitempty" validate:"dive"`
	ContactQuery string                    `json:"contact_query,omitempty" engine:"evaluated"`
	LegacyVars   []string                  `json:"legacy_vars,omitempty" engine:"evaluated"`
}

func (a *otherContactsAction) resolveRecipients(run flows.FlowRun, logEvent flows.EventCallback) ([]*assets.GroupReference, []*flows.ContactReference, string, []urns.URN, error) {
	groupSet := run.Session().Assets().Groups()

	// copy URNs
	urnList := make([]urns.URN, 0, len(a.URNs))
	for _, urn := range a.URNs {
		urnList = append(urnList, urn)
	}

	// copy contact references
	contactRefs := make([]*flows.ContactReference, 0, len(a.Contacts))
	for _, contactRef := range a.Contacts {
		contactRefs = append(contactRefs, contactRef)
	}

	// resolve group references
	groups, err := resolveGroups(run, a.Groups, false, logEvent)
	if err != nil {
		return nil, nil, "", nil, err
	}
	groupRefs := make([]*assets.GroupReference, 0, len(groups))
	for _, group := range groups {
		groupRefs = append(groupRefs, group.Reference())
	}

	// evaluate the legacy variables
	for _, legacyVar := range a.LegacyVars {
		evaluatedLegacyVar, err := run.EvaluateTemplate(legacyVar)
		if err != nil {
			logEvent(events.NewError(err))
		}

		if uuidRegex.MatchString(evaluatedLegacyVar) {
			// if variable evaluates to a UUID, we assume it's a contact UUID
			contactRefs = append(contactRefs, flows.NewContactReference(flows.ContactUUID(evaluatedLegacyVar), ""))

		} else if groupByName := groupSet.FindByName(evaluatedLegacyVar); groupByName != nil {
			// next up we look for a group with a matching name
			groupRefs = append(groupRefs, groupByName.Reference())
		} else {
			// next up try it as a URN
			urn := urns.URN(evaluatedLegacyVar)
			if urn.Validate() == nil {
				urn = urn.Normalize(string(run.Environment().DefaultCountry()))
				urnList = append(urnList, urn)
			} else {
				// if that fails, assume this is a phone number, and let the caller worry about validation
				urn, err := urns.NewURNFromParts(urns.TelScheme, evaluatedLegacyVar, "", "")
				if err != nil {
					logEvent(events.NewError(err))
				} else {
					urn = urn.Normalize(string(run.Environment().DefaultCountry()))
					urnList = append(urnList, urn)
				}
			}
		}
	}

	// evalaute contact query
	contactQuery, err := run.EvaluateTemplateWithEscaping(a.ContactQuery, flows.ContactQueryEscaping)

	return groupRefs, contactRefs, contactQuery, urnList, nil
}

// utility struct for actions which create a message
type createMsgAction struct {
	Text         string   `json:"text" validate:"required" engine:"localized,evaluated"`
	Attachments  []string `json:"attachments,omitempty" engine:"localized,evaluated"`
	QuickReplies []string `json:"quick_replies,omitempty" engine:"localized,evaluated"`
}

// helper function for actions that have a set of group references that must be resolved to actual groups
func resolveGroups(run flows.FlowRun, references []*assets.GroupReference, staticOnly bool, logEvent flows.EventCallback) ([]*flows.Group, error) {
	groupSet := run.Session().Assets().Groups()
	groups := make([]*flows.Group, 0, len(references))

	for _, ref := range references {
		var group *flows.Group

		if ref.UUID != "" {
			// group is a fixed group with a UUID
			group = groupSet.Get(ref.UUID)
		} else {
			// group is an expression that evaluates to an existing group's name
			evaluatedGroupName, err := run.EvaluateTemplate(ref.NameMatch)
			if err != nil {
				logEvent(events.NewError(err))
			} else {
				// look up the set of all groups to see if such a group exists
				group = groupSet.FindByName(evaluatedGroupName)
				if group == nil {
					logEvent(events.NewErrorf("no such group with name '%s'", evaluatedGroupName))
				}
			}
		}

		if group != nil {
			if staticOnly && group.IsDynamic() {
				logEvent(events.NewErrorf("can't add or remove contacts from a dynamic group '%s'", group.Name()))
			} else {
				groups = append(groups, group)
			}
		}
	}

	return groups, nil
}

// helper function for actions that have a set of label references that must be resolved to actual labels
func resolveLabels(run flows.FlowRun, references []*assets.LabelReference, logEvent flows.EventCallback) ([]*flows.Label, error) {
	labelSet := run.Session().Assets().Labels()
	labels := make([]*flows.Label, 0, len(references))

	for _, ref := range references {
		var label *flows.Label

		if ref.UUID != "" {
			// label is a fixed label with a UUID
			label = labelSet.Get(ref.UUID)
		} else {
			// label is an expression that evaluates to an existing label's name
			evaluatedLabelName, err := run.EvaluateTemplate(ref.NameMatch)
			if err != nil {
				logEvent(events.NewError(err))
			} else {
				// look up the set of all labels to see if such a label exists
				label = labelSet.FindByName(evaluatedLabelName)
				if label == nil {
					logEvent(events.NewErrorf("no such label with name '%s'", evaluatedLabelName))
				}
			}
		}

		if label != nil {
			labels = append(labels, label)
		}
	}

	return labels, nil
}

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

// ReadAction reads an action from the given JSON
func ReadAction(data json.RawMessage) (flows.Action, error) {
	typeName, err := utils.ReadTypeFromJSON(data)
	if err != nil {
		return nil, err
	}

	f := registeredTypes[typeName]
	if f == nil {
		return nil, errors.Errorf("unknown type: '%s'", typeName)
	}

	action := f()
	return action, utils.UnmarshalAndValidate(data, action)
}
