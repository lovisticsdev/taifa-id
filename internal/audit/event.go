package audit

import (
	"time"

	"taifa-id/internal/platform/ids"
)

const SourceTaifaID = "taifa-id"

const (
	ResultSuccess = "SUCCESS"
	ResultFailure = "FAILURE"
	ResultAllowed = "ALLOWED"
	ResultDenied  = "DENIED"
)

const (
	ResourcePerson                 = "PERSON"
	ResourceOrganization           = "ORGANIZATION"
	ResourceOrganizationCapability = "ORGANIZATION_CAPABILITY"
	ResourceOrganizationMembership = "ORGANIZATION_MEMBERSHIP"
	ResourceMembershipRole         = "MEMBERSHIP_ROLE"
	ResourceCredential             = "CREDENTIAL"
	ResourceActorContext           = "ACTOR_CONTEXT"
	ResourceAuthSession            = "AUTH_SESSION"
)

const (
	ActionCreate              = "CREATE"
	ActionUpdateStatus        = "UPDATE_STATUS"
	ActionAddCapability       = "ADD_CAPABILITY"
	ActionRemoveCapability    = "REMOVE_CAPABILITY"
	ActionAddRole             = "ADD_ROLE"
	ActionRemoveRole          = "REMOVE_ROLE"
	ActionAuthenticate        = "AUTHENTICATE"
	ActionIntrospectToken     = "INTROSPECT_TOKEN"
	ActionResolveActorContext = "RESOLVE_ACTOR_CONTEXT"
)

const (
	EventPersonCreated       = "identity.person.created"
	EventPersonStatusChanged = "identity.person.status_changed"

	EventOrganizationCreated       = "identity.organization.created"
	EventOrganizationStatusChanged = "identity.organization.status_changed"

	EventOrganizationCapabilityAdded   = "identity.organization_capability.added"
	EventOrganizationCapabilityRemoved = "identity.organization_capability.removed"

	EventMembershipCreated       = "identity.membership.created"
	EventMembershipStatusChanged = "identity.membership.status_changed"
	EventMembershipRoleAdded     = "identity.membership_role.added"
	EventMembershipRoleRemoved   = "identity.membership_role.removed"

	EventCredentialCreated = "identity.credential.created"

	EventAuthSucceeded       = "identity.auth.succeeded"
	EventAuthFailed          = "identity.auth.failed"
	EventTokenIntrospected   = "identity.auth.token_introspected"
	EventActorContextAllowed = "identity.actor_context.allowed"
	EventActorContextDenied  = "identity.actor_context.denied"
)

type Event struct {
	ID            string
	EventType     string
	SourceSystem  string
	SubjectID     string
	ActorID       string
	ResourceType  string
	ResourceID    string
	Action        string
	Result        string
	CorrelationID string
	Payload       map[string]any
	CreatedAt     time.Time
}

func (e Event) WithDefaults() Event {
	if e.ID == "" {
		e.ID = ids.NewEventID()
	}

	if e.SourceSystem == "" {
		e.SourceSystem = SourceTaifaID
	}

	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}

	if e.Payload == nil {
		e.Payload = map[string]any{}
	}

	return e
}
