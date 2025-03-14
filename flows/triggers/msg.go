package triggers

import (
	"encoding/json"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/envs"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/flows/inputs"
	"github.com/nyaruka/goflow/utils"
)

func init() {
	registerType(TypeMsg, readMsgTrigger)
}

// TypeMsg is the type for message triggered sessions
const TypeMsg string = "msg"

// MsgTrigger is used when a session was triggered by a message being recieved by the caller
//
//   {
//     "type": "msg",
//     "flow": {"uuid": "50c3706e-fedb-42c0-8eab-dda3335714b7", "name": "Registration"},
//     "contact": {
//       "uuid": "9f7ede93-4b16-4692-80ad-b7dc54a1cd81",
//       "name": "Bob",
//       "created_on": "2018-01-01T12:00:00.000000Z"
//     },
//     "msg": {
//       "uuid": "2d611e17-fb22-457f-b802-b8f7ec5cda5b",
//       "channel": {"uuid": "61602f3e-f603-4c70-8a8f-c477505bf4bf", "name": "Twilio"},
//       "urn": "tel:+12065551212",
//       "text": "hi there",
//       "attachments": ["https://s3.amazon.com/mybucket/attachment.jpg"]
//     },
//     "keyword_match": {
//       "type": "first_word",
//       "keyword": "start"
//     },
//     "triggered_on": "2000-01-01T00:00:00.000000000-00:00"
//   }
//
// @trigger msg
type MsgTrigger struct {
	baseTrigger
	msg   *flows.MsgIn
	match *KeywordMatch
}

// KeywordMatchType describes how the message matched a keyword
type KeywordMatchType string

// the different types of keyword match
const (
	KeywordMatchTypeFirstWord KeywordMatchType = "first_word"
	KeywordMatchTypeOnlyWord  KeywordMatchType = "only_word"
)

// KeywordMatch describes why the message triggered a session
type KeywordMatch struct {
	Type    KeywordMatchType `json:"type" validate:"required"`
	Keyword string           `json:"keyword" validate:"required"`
}

// NewKeywordMatch creates a new keyword match
func NewKeywordMatch(typeName KeywordMatchType, keyword string) *KeywordMatch {
	return &KeywordMatch{Type: typeName, Keyword: keyword}
}

// NewMsg creates a new message trigger
func NewMsg(env envs.Environment, flow *assets.FlowReference, contact *flows.Contact, msg *flows.MsgIn, match *KeywordMatch) flows.Trigger {
	return &MsgTrigger{
		baseTrigger: newBaseTrigger(TypeMsg, env, flow, contact, nil, nil),
		msg:         msg,
		match:       match,
	}
}

// InitializeRun performs additional initialization when we visit our first node
func (t *MsgTrigger) InitializeRun(run flows.FlowRun, logEvent flows.EventCallback) error {
	// update our input
	input, err := inputs.NewMsg(run.Session().Assets(), t.msg, t.triggeredOn)
	if err != nil {
		return err
	}

	run.Session().SetInput(input)
	logEvent(events.NewMsgReceived(t.msg))

	return t.baseTrigger.InitializeRun(run, logEvent)
}

// Context for msg triggers additionally exposes the keyword match
func (t *MsgTrigger) Context(env envs.Environment) map[string]types.XValue {
	var keyword types.XValue
	if t.match != nil {
		keyword = types.NewXText(t.match.Keyword)
	}

	return map[string]types.XValue{
		"type":    types.NewXText(t.type_),
		"params":  t.params,
		"keyword": keyword,
	}
}

var _ flows.Trigger = (*MsgTrigger)(nil)

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

type msgTriggerEnvelope struct {
	baseTriggerEnvelope
	Msg   *flows.MsgIn  `json:"msg" validate:"required,dive"`
	Match *KeywordMatch `json:"keyword_match,omitempty" validate:"omitempty,dive"`
}

func readMsgTrigger(sessionAssets flows.SessionAssets, data json.RawMessage, missing assets.MissingCallback) (flows.Trigger, error) {
	e := &msgTriggerEnvelope{}
	if err := utils.UnmarshalAndValidate(data, e); err != nil {
		return nil, err
	}

	t := &MsgTrigger{
		msg:   e.Msg,
		match: e.Match,
	}

	if err := t.unmarshal(sessionAssets, &e.baseTriggerEnvelope, missing); err != nil {
		return nil, err
	}

	return t, nil
}

// MarshalJSON marshals this trigger into JSON
func (t *MsgTrigger) MarshalJSON() ([]byte, error) {
	e := &msgTriggerEnvelope{
		Msg:   t.msg,
		Match: t.match,
	}

	if err := t.marshal(&e.baseTriggerEnvelope); err != nil {
		return nil, err
	}

	return json.Marshal(e)
}
