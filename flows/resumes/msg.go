package resumes

import (
	"encoding/json"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/flows/inputs"
	"github.com/nyaruka/goflow/utils"
)

func init() {
	RegisterType(TypeMsg, ReadMsgResume)
}

// TypeMsg is the type for resuming a session with a message
const TypeMsg string = "msg"

// MsgResume is used when a session is resumed with a new message from the contact
//
//   {
//     "type": "msg",
//     "contact": {
//       "uuid": "9f7ede93-4b16-4692-80ad-b7dc54a1cd81",
//       "name": "Bob",
//       "created_on": "2018-01-01T12:00:00.000000Z",
//       "language": "fra",
//       "fields": {"gender": {"text": "Male"}},
//       "groups": []
//     },
//     "msg": {
//       "uuid": "2d611e17-fb22-457f-b802-b8f7ec5cda5b",
//       "channel": {"uuid": "61602f3e-f603-4c70-8a8f-c477505bf4bf", "name": "Twilio"},
//       "urn": "tel:+12065551212",
//       "text": "hi there",
//       "attachments": ["https://s3.amazon.com/mybucket/attachment.jpg"]
//     },
//     "resumed_on": "2000-01-01T00:00:00.000000000-00:00"
//   }
//
// @resume msg
type MsgResume struct {
	baseResume
	msg *flows.MsgIn
}

// NewMsgResume creates a new message resume with the passed in values
func NewMsgResume(env utils.Environment, contact *flows.Contact, msg *flows.MsgIn) *MsgResume {
	return &MsgResume{
		baseResume: newBaseResume(TypeMsg, env, contact),
		msg:        msg,
	}
}

// Apply applies our state changes and saves any events to the run
func (r *MsgResume) Apply(run flows.FlowRun, step flows.Step) error {
	// update the run's input
	input, err := inputs.NewMsgInput(run.Session().Assets(), r.msg, r.ResumedOn())
	if err != nil {
		return err
	}

	run.SetInput(input)
	run.LogEvent(step, events.NewMsgReceivedEvent(r.msg))

	return r.baseResume.Apply(run, step)
}

var _ flows.Resume = (*MsgResume)(nil)

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

type msgResumeEnvelope struct {
	baseResumeEnvelope
	Msg *flows.MsgIn `json:"msg" validate:"required,dive"`
}

// ReadMsgResume reads a message resume
func ReadMsgResume(session flows.Session, data json.RawMessage) (flows.Resume, error) {
	e := &msgResumeEnvelope{}
	if err := utils.UnmarshalAndValidate(data, e); err != nil {
		return nil, err
	}

	r := &MsgResume{
		msg: e.Msg,
	}

	if err := r.unmarshal(session, &e.baseResumeEnvelope); err != nil {
		return nil, err
	}

	return r, nil
}

// MarshalJSON marshals this resume into JSON
func (r *MsgResume) MarshalJSON() ([]byte, error) {
	e := &msgResumeEnvelope{
		Msg: r.msg,
	}

	if err := r.marshal(&e.baseResumeEnvelope); err != nil {
		return nil, err
	}

	return json.Marshal(e)
}
