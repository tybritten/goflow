package flows

import (
	"time"

	"github.com/nyaruka/goflow/utils"
)

type UUID string
type NodeUUID UUID

func (u NodeUUID) String() string { return string(u) }

type ExitUUID UUID

func (u ExitUUID) String() string { return string(u) }

type FlowUUID UUID

func (u FlowUUID) String() string { return string(u) }

type ActionUUID UUID

func (u ActionUUID) String() string { return string(u) }

type ContactUUID UUID

func (u ContactUUID) String() string { return string(u) }

type FieldUUID UUID

func (u FieldUUID) String() string { return string(u) }

type ChannelUUID UUID

func (u ChannelUUID) String() string { return string(u) }

type RunUUID UUID

func (u RunUUID) String() string { return string(u) }

type StepUUID UUID

func (u StepUUID) String() string { return string(u) }

type LabelUUID UUID

func (u LabelUUID) String() string { return string(u) }

type GroupUUID UUID

func (u GroupUUID) String() string { return string(u) }

type Flow interface {
	UUID() FlowUUID
	Name() string
	Language() utils.Language
	ExpireAfterMinutes() *int
	Translations() FlowTranslations

	Nodes() []Node
	GetNode(uuid NodeUUID) Node

	Validate(Assets) error
}

// RunStatus represents the current status of the flow run
type RunStatus string

const (
	// StatusActive represents an active flow run that is awaiting input
	StatusActive RunStatus = "active"

	// StatusCompleted represents a flow run that has run to completion
	StatusCompleted RunStatus = "completed"

	// StatusErrored represents a flow run that encountered an error
	StatusErrored RunStatus = "errored"

	// StatusExpired represents a flow run that expired due to inactivity
	StatusExpired RunStatus = "expired"

	// StatusInterrupted represents a flow run that was interrupted by another flow
	StatusInterrupted RunStatus = "interrupted"
)

func (r RunStatus) String() string { return string(r) }

type Assets interface {
	Validate() error

	GetChannel(ChannelUUID) (Channel, error)
	GetFlow(FlowUUID) (Flow, error)
}

type Node interface {
	UUID() NodeUUID

	Router() Router
	Actions() []Action
	Exits() []Exit
	Wait() Wait
}

type Action interface {
	UUID() ActionUUID

	Execute(FlowRun, Step) error
	Validate(Assets) error
	utils.Typed
}

type Router interface {
	PickRoute(FlowRun, []Exit, Step) (Route, error)
	Validate([]Exit) error
	ResultName() string
	utils.Typed
}

type Route struct {
	exit  ExitUUID
	match string
}

func (r Route) Exit() ExitUUID { return r.exit }
func (r Route) Match() string  { return r.match }

var NoRoute = Route{}

func NewRoute(exit ExitUUID, match string) Route {
	return Route{exit, match}
}

type Exit interface {
	UUID() ExitUUID
	DestinationNodeUUID() NodeUUID
	Name() string
}

type Wait interface {
	Begin(FlowRun, Step) error
	GetEndEvent(FlowRun, Step) (Event, error)
	End(FlowRun, Step, Event) error
	utils.Typed
	utils.VariableResolver
}

// FlowTranslations provide a way to get the Translations for a flow for a specific language
type FlowTranslations interface {
	GetLanguageTranslations(utils.Language) Translations
}

// Translations provide a way to get the translation for a specific language for a uuid/key pair
type Translations interface {
	GetTextArray(uuid UUID, key string) []string
}

type Event interface {
	CreatedOn() time.Time
	SetCreatedOn(time.Time)

	FromCaller() bool
	SetFromCaller(bool)

	Apply(FlowRun)

	utils.Typed
}

type Input interface {
	utils.VariableResolver
	utils.Typed

	CreatedOn() time.Time
	Channel() Channel
}

type Step interface {
	utils.VariableResolver

	UUID() StepUUID
	NodeUUID() NodeUUID
	ExitUUID() ExitUUID

	ArrivedOn() time.Time
	LeftOn() *time.Time

	Leave(ExitUUID)

	Events() []Event
}

// LogEntry is a container for a new event generated by the engine, i.e. not from the caller
type LogEntry interface {
	Step() Step
	Action() Action
	Event() Event
}

// Session represents the session of a flow run which may contain many runs
type Session interface {
	Assets() Assets

	Environment() utils.Environment
	SetEnvironment(utils.Environment)

	Contact() *Contact
	SetContact(*Contact)

	StartFlow(FlowUUID, FlowRun, []Event) error
	Resume([]Event) error

	CreateRun(Flow, FlowRun) FlowRun
	Runs() []FlowRun
	GetRun(RunUUID) (FlowRun, error)
	ActiveRun() FlowRun

	Log() []LogEntry
	LogEvent(Step, Action, Event)
	ClearLog()
}

// FlowRun represents a single run on a flow by a single contact
type FlowRun interface {
	UUID() RunUUID

	Environment() utils.Environment
	Session() Session
	Context() utils.VariableResolver

	Flow() Flow
	Results() *Results

	Contact() *Contact
	SetContact(*Contact)

	SetExtra(utils.JSONFragment)
	Extra() utils.JSONFragment

	Status() RunStatus
	Exit(RunStatus)
	IsComplete() bool

	Wait() Wait
	SetWait(Wait)

	Input() Input
	SetInput(Input)

	ApplyEvent(Step, Action, Event)
	AddError(Step, error)

	CreateStep(Node) Step
	Path() []Step

	GetText(uuid UUID, key string, native string) string
	GetTextArray(uuid UUID, key string, native []string) []string

	Webhook() *utils.RequestResponse
	SetWebhook(*utils.RequestResponse)

	Child() FlowRunReference
	Parent() FlowRunReference

	ExpiresOn() *time.Time
	ResetExpiration(*time.Time)

	CreatedOn() time.Time
	ModifiedOn() time.Time
	TimesOutOn() *time.Time
	ExitedOn() *time.Time
}

// FlowRunReference represents a flow run reference within a flow
type FlowRunReference interface {
	UUID() RunUUID
	Flow() Flow
	Contact() *Contact

	Results() *Results
	Status() RunStatus

	ExpiresOn() *time.Time
	ResetExpiration(*time.Time)

	CreatedOn() time.Time
	ModifiedOn() time.Time
	TimesOutOn() *time.Time
	ExitedOn() *time.Time
}

// ChannelType represents the type of a Channel
type ChannelType string

func (ct ChannelType) String() string { return string(ct) }

// Channel represents a channel for sending and receiving messages
type Channel interface {
	UUID() ChannelUUID
	Name() string
	Address() string
	Type() ChannelType
}

// MsgDirection is the direction of a Msg (either in or out)
type MsgDirection string

const (
	// MsgOut represents an outgoing message
	MsgOut MsgDirection = "O"

	// MsgIn represents an incoming message
	MsgIn MsgDirection = "I"
)
