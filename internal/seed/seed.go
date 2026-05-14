package seed

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-id/internal/audit"
)

const DefaultSeedPassword = "ExampleDevPass123!"

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
}

type Config struct {
	DefaultPassword string
}

type Result struct {
	Persons       int
	Organizations int
	Capabilities  int
	Memberships   int
	Roles         int
	Credentials   int
	AuditEvents   int
}

type Runner struct {
	pool            *pgxpool.Pool
	hasher          PasswordHasher
	defaultPassword string
}

func NewRunner(pool *pgxpool.Pool, hasher PasswordHasher, cfg Config) *Runner {
	defaultPassword := strings.TrimSpace(cfg.DefaultPassword)
	if defaultPassword == "" {
		defaultPassword = DefaultSeedPassword
	}

	return &Runner{
		pool:            pool,
		hasher:          hasher,
		defaultPassword: defaultPassword,
	}
}

type personSeed struct {
	ID           string
	SyntheticNIN string
	DisplayName  string
}

type organizationSeed struct {
	ID          string
	Name        string
	PrimaryType string
}

type capabilitySeed struct {
	ID             string
	OrganizationID string
	Capability     string
}

type membershipSeed struct {
	ID             string
	PersonID       string
	OrganizationID string
	MembershipType string
}

type roleSeed struct {
	ID             string
	MembershipID   string
	PersonID       string
	OrganizationID string
	Role           string
}

type credentialSeed struct {
	ID       string
	PersonID string
	Username string
}

var persons = []personSeed{
	{
		ID:           "PER-SEED-CITIZEN-001",
		SyntheticNIN: "NIN-SEED-CIT-001",
		DisplayName:  "Amina Citizen Seed",
	},
	{
		ID:           "PER-SEED-CLINICIAN-001",
		SyntheticNIN: "NIN-SEED-CLIN-001",
		DisplayName:  "Nia Clinician Seed",
	},
	{
		ID:           "PER-SEED-CLAIMS-001",
		SyntheticNIN: "NIN-SEED-CLAIMS-001",
		DisplayName:  "Peter Claims Seed",
	},
	{
		ID:           "PER-SEED-EMPLOYER-001",
		SyntheticNIN: "NIN-SEED-EMP-001",
		DisplayName:  "Esther Employer Seed",
	},
	{
		ID:           "PER-SEED-TAX-001",
		SyntheticNIN: "NIN-SEED-TAX-001",
		DisplayName:  "Omar Tax Seed",
	},
	{
		ID:           "PER-SEED-PAY-001",
		SyntheticNIN: "NIN-SEED-PAY-001",
		DisplayName:  "Grace Pay Seed",
	},
	{
		ID:           "PER-SEED-OBSERVE-001",
		SyntheticNIN: "NIN-SEED-OBS-001",
		DisplayName:  "David Observe Seed",
	},
	{
		ID:           "PER-SEED-ADMIN-001",
		SyntheticNIN: "NIN-SEED-ADMIN-001",
		DisplayName:  "Sana Admin Seed",
	},
}

var organizations = []organizationSeed{
	{
		ID:          "ORG-GOV-CARE",
		Name:        "Taifa Care Authority",
		PrimaryType: "GOVERNMENT_AGENCY",
	},
	{
		ID:          "ORG-GOV-TAX",
		Name:        "Taifa Tax Authority",
		PrimaryType: "GOVERNMENT_AGENCY",
	},
	{
		ID:          "ORG-GOV-PAY",
		Name:        "Taifa Pay Authority",
		PrimaryType: "GOVERNMENT_AGENCY",
	},
	{
		ID:          "ORG-GOV-OBSERVE",
		Name:        "Taifa Observe Authority",
		PrimaryType: "GOVERNMENT_AGENCY",
	},
	{
		ID:          "ORG-HP-HOSP",
		Name:        "Taifa National Hospital",
		PrimaryType: "HEALTH_PROVIDER",
	},
	{
		ID:          "ORG-HP-CLINIC",
		Name:        "Taifa Community Clinic",
		PrimaryType: "HEALTH_PROVIDER",
	},
	{
		ID:          "ORG-EMP-MFG",
		Name:        "Taifa Manufacturing Employer",
		PrimaryType: "EMPLOYER",
	},
	{
		ID:          "ORG-FIN-CB",
		Name:        "Taifa Central Bank Simulation",
		PrimaryType: "FINANCIAL_INSTITUTION_SIM",
	},
	{
		ID:          "ORG-FIN-COMM",
		Name:        "Taifa Commercial Bank Simulation",
		PrimaryType: "FINANCIAL_INSTITUTION_SIM",
	},
}

var capabilities = []capabilitySeed{
	{
		ID:             "CAP-SEED-GOV-CARE-OPERATE",
		OrganizationID: "ORG-GOV-CARE",
		Capability:     "CAN_OPERATE_GOVERNMENT_SERVICE",
	},
	{
		ID:             "CAP-SEED-GOV-TAX-OPERATE",
		OrganizationID: "ORG-GOV-TAX",
		Capability:     "CAN_OPERATE_GOVERNMENT_SERVICE",
	},
	{
		ID:             "CAP-SEED-GOV-PAY-OPERATE",
		OrganizationID: "ORG-GOV-PAY",
		Capability:     "CAN_OPERATE_GOVERNMENT_SERVICE",
	},
	{
		ID:             "CAP-SEED-GOV-PAY-ROUTE",
		OrganizationID: "ORG-GOV-PAY",
		Capability:     "CAN_ROUTE_PAYMENTS",
	},
	{
		ID:             "CAP-SEED-GOV-PAY-RESERVE",
		OrganizationID: "ORG-GOV-PAY",
		Capability:     "CAN_HOLD_RESERVE_ACCOUNT",
	},
	{
		ID:             "CAP-SEED-GOV-OBSERVE-OPERATE",
		OrganizationID: "ORG-GOV-OBSERVE",
		Capability:     "CAN_OPERATE_GOVERNMENT_SERVICE",
	},
	{
		ID:             "CAP-SEED-GOV-OBSERVE-SECURITY",
		OrganizationID: "ORG-GOV-OBSERVE",
		Capability:     "CAN_OBSERVE_SECURITY_EVENTS",
	},
	{
		ID:             "CAP-SEED-HOSP-PROVIDE",
		OrganizationID: "ORG-HP-HOSP",
		Capability:     "CAN_PROVIDE_HEALTH_SERVICES",
	},
	{
		ID:             "CAP-SEED-HOSP-PAYOUT",
		OrganizationID: "ORG-HP-HOSP",
		Capability:     "CAN_RECEIVE_HEALTH_PAYOUTS",
	},
	{
		ID:             "CAP-SEED-CLINIC-PROVIDE",
		OrganizationID: "ORG-HP-CLINIC",
		Capability:     "CAN_PROVIDE_HEALTH_SERVICES",
	},
	{
		ID:             "CAP-SEED-CLINIC-PAYOUT",
		OrganizationID: "ORG-HP-CLINIC",
		Capability:     "CAN_RECEIVE_HEALTH_PAYOUTS",
	},
	{
		ID:             "CAP-SEED-EMP-MFG-EMPLOY",
		OrganizationID: "ORG-EMP-MFG",
		Capability:     "CAN_EMPLOY_PERSONS",
	},
	{
		ID:             "CAP-SEED-EMP-MFG-TAX",
		OrganizationID: "ORG-EMP-MFG",
		Capability:     "CAN_SUBMIT_TAX_CONTRIBUTIONS",
	},
	{
		ID:             "CAP-SEED-CB-ROUTE",
		OrganizationID: "ORG-FIN-CB",
		Capability:     "CAN_ROUTE_PAYMENTS",
	},
	{
		ID:             "CAP-SEED-CB-RESERVE",
		OrganizationID: "ORG-FIN-CB",
		Capability:     "CAN_HOLD_RESERVE_ACCOUNT",
	},
	{
		ID:             "CAP-SEED-COMM-ROUTE",
		OrganizationID: "ORG-FIN-COMM",
		Capability:     "CAN_ROUTE_PAYMENTS",
	},
}

var memberships = []membershipSeed{
	{
		ID:             "MEM-SEED-CITIZEN-CARE",
		PersonID:       "PER-SEED-CITIZEN-001",
		OrganizationID: "ORG-GOV-CARE",
		MembershipType: "AGENCY_STAFF",
	},
	{
		ID:             "MEM-SEED-CLINICIAN-CLINIC",
		PersonID:       "PER-SEED-CLINICIAN-001",
		OrganizationID: "ORG-HP-CLINIC",
		MembershipType: "PROVIDER_STAFF",
	},
	{
		ID:             "MEM-SEED-CLAIMS-HOSP",
		PersonID:       "PER-SEED-CLAIMS-001",
		OrganizationID: "ORG-HP-HOSP",
		MembershipType: "PROVIDER_STAFF",
	},
	{
		ID:             "MEM-SEED-EMPLOYER-MFG",
		PersonID:       "PER-SEED-EMPLOYER-001",
		OrganizationID: "ORG-EMP-MFG",
		MembershipType: "EMPLOYEE",
	},
	{
		ID:             "MEM-SEED-TAX-GOV",
		PersonID:       "PER-SEED-TAX-001",
		OrganizationID: "ORG-GOV-TAX",
		MembershipType: "AGENCY_STAFF",
	},
	{
		ID:             "MEM-SEED-PAY-GOV",
		PersonID:       "PER-SEED-PAY-001",
		OrganizationID: "ORG-GOV-PAY",
		MembershipType: "FINANCIAL_OPERATOR",
	},
	{
		ID:             "MEM-SEED-OBSERVE-GOV",
		PersonID:       "PER-SEED-OBSERVE-001",
		OrganizationID: "ORG-GOV-OBSERVE",
		MembershipType: "AUDITOR",
	},
	{
		ID:             "MEM-SEED-ADMIN-OBSERVE",
		PersonID:       "PER-SEED-ADMIN-001",
		OrganizationID: "ORG-GOV-OBSERVE",
		MembershipType: "SYSTEM_ADMIN",
	},
}

var roles = []roleSeed{
	{
		ID:             "ROLE-SEED-CITIZEN",
		MembershipID:   "MEM-SEED-CITIZEN-CARE",
		PersonID:       "PER-SEED-CITIZEN-001",
		OrganizationID: "ORG-GOV-CARE",
		Role:           "CITIZEN",
	},
	{
		ID:             "ROLE-SEED-CLINICIAN",
		MembershipID:   "MEM-SEED-CLINICIAN-CLINIC",
		PersonID:       "PER-SEED-CLINICIAN-001",
		OrganizationID: "ORG-HP-CLINIC",
		Role:           "PROVIDER_CLINICIAN",
	},
	{
		ID:             "ROLE-SEED-CLAIMS",
		MembershipID:   "MEM-SEED-CLAIMS-HOSP",
		PersonID:       "PER-SEED-CLAIMS-001",
		OrganizationID: "ORG-HP-HOSP",
		Role:           "PROVIDER_CLAIMS_OFFICER",
	},
	{
		ID:             "ROLE-SEED-EMPLOYER",
		MembershipID:   "MEM-SEED-EMPLOYER-MFG",
		PersonID:       "PER-SEED-EMPLOYER-001",
		OrganizationID: "ORG-EMP-MFG",
		Role:           "EMPLOYER_SUBMITTER",
	},
	{
		ID:             "ROLE-SEED-TAX",
		MembershipID:   "MEM-SEED-TAX-GOV",
		PersonID:       "PER-SEED-TAX-001",
		OrganizationID: "ORG-GOV-TAX",
		Role:           "TAX_OFFICER",
	},
	{
		ID:             "ROLE-SEED-PAY",
		MembershipID:   "MEM-SEED-PAY-GOV",
		PersonID:       "PER-SEED-PAY-001",
		OrganizationID: "ORG-GOV-PAY",
		Role:           "PAY_OPERATOR",
	},
	{
		ID:             "ROLE-SEED-OBSERVE-ANALYST",
		MembershipID:   "MEM-SEED-OBSERVE-GOV",
		PersonID:       "PER-SEED-OBSERVE-001",
		OrganizationID: "ORG-GOV-OBSERVE",
		Role:           "OBSERVE_ANALYST",
	},
	{
		ID:             "ROLE-SEED-OBSERVE-AUDITOR",
		MembershipID:   "MEM-SEED-OBSERVE-GOV",
		PersonID:       "PER-SEED-OBSERVE-001",
		OrganizationID: "ORG-GOV-OBSERVE",
		Role:           "OBSERVE_AUDITOR",
	},
	{
		ID:             "ROLE-SEED-ADMIN",
		MembershipID:   "MEM-SEED-ADMIN-OBSERVE",
		PersonID:       "PER-SEED-ADMIN-001",
		OrganizationID: "ORG-GOV-OBSERVE",
		Role:           "SYSTEM_ADMIN",
	},
}

var credentials = []credentialSeed{
	{
		ID:       "CRD-SEED-CITIZEN",
		PersonID: "PER-SEED-CITIZEN-001",
		Username: "citizen.seed",
	},
	{
		ID:       "CRD-SEED-CLINICIAN",
		PersonID: "PER-SEED-CLINICIAN-001",
		Username: "clinician.seed",
	},
	{
		ID:       "CRD-SEED-CLAIMS",
		PersonID: "PER-SEED-CLAIMS-001",
		Username: "claims.seed",
	},
	{
		ID:       "CRD-SEED-EMPLOYER",
		PersonID: "PER-SEED-EMPLOYER-001",
		Username: "employer.seed",
	},
	{
		ID:       "CRD-SEED-TAX",
		PersonID: "PER-SEED-TAX-001",
		Username: "tax.seed",
	},
	{
		ID:       "CRD-SEED-PAY",
		PersonID: "PER-SEED-PAY-001",
		Username: "pay.seed",
	},
	{
		ID:       "CRD-SEED-OBSERVE",
		PersonID: "PER-SEED-OBSERVE-001",
		Username: "observe.seed",
	},
	{
		ID:       "CRD-SEED-ADMIN",
		PersonID: "PER-SEED-ADMIN-001",
		Username: "admin.seed",
	},
}

func (r *Runner) Run(ctx context.Context) (Result, error) {
	if r.pool == nil {
		return Result{}, fmt.Errorf("seed runner requires database pool")
	}

	if r.hasher == nil {
		return Result{}, fmt.Errorf("seed runner requires password hasher")
	}

	if !isValidPassword(r.defaultPassword) {
		return Result{}, fmt.Errorf("seed password must be 8 to 256 characters")
	}

	result := Result{}

	for _, person := range persons {
		created, err := r.insertPerson(ctx, person)
		if err != nil {
			return result, err
		}

		if created {
			result.Persons++
			if err := r.writeAudit(ctx, audit.Event{
				EventType:    audit.EventPersonCreated,
				SubjectID:    person.ID,
				ResourceType: audit.ResourcePerson,
				ResourceID:   person.ID,
				Action:       audit.ActionCreate,
				Result:       audit.ResultSuccess,
				Payload: map[string]any{
					"seed":          true,
					"synthetic_nin": person.SyntheticNIN,
					"display_name":  person.DisplayName,
					"status":        "ACTIVE",
				},
			}); err != nil {
				return result, err
			}
			result.AuditEvents++
		}
	}

	for _, organization := range organizations {
		created, err := r.insertOrganization(ctx, organization)
		if err != nil {
			return result, err
		}

		if created {
			result.Organizations++
			if err := r.writeAudit(ctx, audit.Event{
				EventType:    audit.EventOrganizationCreated,
				SubjectID:    organization.ID,
				ResourceType: audit.ResourceOrganization,
				ResourceID:   organization.ID,
				Action:       audit.ActionCreate,
				Result:       audit.ResultSuccess,
				Payload: map[string]any{
					"seed":         true,
					"name":         organization.Name,
					"primary_type": organization.PrimaryType,
					"status":       "ACTIVE",
				},
			}); err != nil {
				return result, err
			}
			result.AuditEvents++
		}
	}

	for _, capability := range capabilities {
		created, err := r.insertCapability(ctx, capability)
		if err != nil {
			return result, err
		}

		if created {
			result.Capabilities++
			if err := r.writeAudit(ctx, audit.Event{
				EventType:    audit.EventOrganizationCapabilityAdded,
				SubjectID:    capability.OrganizationID,
				ResourceType: audit.ResourceOrganizationCapability,
				ResourceID:   capability.ID,
				Action:       audit.ActionAddCapability,
				Result:       audit.ResultSuccess,
				Payload: map[string]any{
					"seed":            true,
					"organization_id": capability.OrganizationID,
					"capability":      capability.Capability,
				},
			}); err != nil {
				return result, err
			}
			result.AuditEvents++
		}
	}

	for _, membership := range memberships {
		created, err := r.insertMembership(ctx, membership)
		if err != nil {
			return result, err
		}

		if created {
			result.Memberships++
			if err := r.writeAudit(ctx, audit.Event{
				EventType:    audit.EventMembershipCreated,
				SubjectID:    membership.PersonID,
				ResourceType: audit.ResourceOrganizationMembership,
				ResourceID:   membership.ID,
				Action:       audit.ActionCreate,
				Result:       audit.ResultSuccess,
				Payload: map[string]any{
					"seed":              true,
					"person_id":         membership.PersonID,
					"organization_id":   membership.OrganizationID,
					"membership_type":   membership.MembershipType,
					"membership_status": "ACTIVE",
				},
			}); err != nil {
				return result, err
			}
			result.AuditEvents++
		}
	}

	for _, role := range roles {
		created, err := r.insertRole(ctx, role)
		if err != nil {
			return result, err
		}

		if created {
			result.Roles++
			if err := r.writeAudit(ctx, audit.Event{
				EventType:    audit.EventMembershipRoleAdded,
				SubjectID:    role.PersonID,
				ResourceType: audit.ResourceMembershipRole,
				ResourceID:   role.ID,
				Action:       audit.ActionAddRole,
				Result:       audit.ResultSuccess,
				Payload: map[string]any{
					"seed":            true,
					"membership_id":   role.MembershipID,
					"person_id":       role.PersonID,
					"organization_id": role.OrganizationID,
					"role":            role.Role,
				},
			}); err != nil {
				return result, err
			}
			result.AuditEvents++
		}
	}

	for _, credential := range credentials {
		created, err := r.insertCredential(ctx, credential)
		if err != nil {
			return result, err
		}

		if created {
			result.Credentials++
			if err := r.writeAudit(ctx, audit.Event{
				EventType:    audit.EventCredentialCreated,
				SubjectID:    credential.PersonID,
				ResourceType: audit.ResourceCredential,
				ResourceID:   credential.ID,
				Action:       audit.ActionCreate,
				Result:       audit.ResultSuccess,
				Payload: map[string]any{
					"seed":      true,
					"person_id": credential.PersonID,
					"username":  credential.Username,
					"status":    "ACTIVE",
				},
			}); err != nil {
				return result, err
			}
			result.AuditEvents++
		}
	}

	return result, nil
}

func (r *Runner) insertPerson(ctx context.Context, person personSeed) (bool, error) {
	const query = `
		INSERT INTO persons (
			id,
			synthetic_nin,
			display_name,
			status
		)
		VALUES ($1, $2, $3, 'ACTIVE')
		ON CONFLICT DO NOTHING
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		person.ID,
		person.SyntheticNIN,
		person.DisplayName,
	)
	if err != nil {
		return false, fmt.Errorf("seed person %s: %w", person.ID, err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *Runner) insertOrganization(ctx context.Context, organization organizationSeed) (bool, error) {
	const query = `
		INSERT INTO organizations (
			id,
			name,
			primary_type,
			status
		)
		VALUES ($1, $2, $3, 'ACTIVE')
		ON CONFLICT DO NOTHING
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		organization.ID,
		organization.Name,
		organization.PrimaryType,
	)
	if err != nil {
		return false, fmt.Errorf("seed organization %s: %w", organization.ID, err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *Runner) insertCapability(ctx context.Context, capability capabilitySeed) (bool, error) {
	const query = `
		INSERT INTO organization_capabilities (
			id,
			organization_id,
			capability
		)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		capability.ID,
		capability.OrganizationID,
		capability.Capability,
	)
	if err != nil {
		return false, fmt.Errorf("seed capability %s: %w", capability.ID, err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *Runner) insertMembership(ctx context.Context, membership membershipSeed) (bool, error) {
	const query = `
		INSERT INTO organization_memberships (
			id,
			person_id,
			organization_id,
			membership_type,
			status
		)
		VALUES ($1, $2, $3, $4, 'ACTIVE')
		ON CONFLICT DO NOTHING
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		membership.ID,
		membership.PersonID,
		membership.OrganizationID,
		membership.MembershipType,
	)
	if err != nil {
		return false, fmt.Errorf("seed membership %s: %w", membership.ID, err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *Runner) insertRole(ctx context.Context, role roleSeed) (bool, error) {
	const query = `
		INSERT INTO membership_roles (
			id,
			membership_id,
			role
		)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		role.ID,
		role.MembershipID,
		role.Role,
	)
	if err != nil {
		return false, fmt.Errorf("seed role %s: %w", role.ID, err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *Runner) insertCredential(ctx context.Context, credential credentialSeed) (bool, error) {
	passwordHash, err := r.hasher.Hash(r.defaultPassword)
	if err != nil {
		return false, fmt.Errorf("hash seed credential %s: %w", credential.ID, err)
	}

	const query = `
		INSERT INTO credentials (
			id,
			person_id,
			username,
			password_hash,
			status
		)
		VALUES ($1, $2, $3, $4, 'ACTIVE')
		ON CONFLICT DO NOTHING
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		credential.ID,
		credential.PersonID,
		credential.Username,
		passwordHash,
	)
	if err != nil {
		return false, fmt.Errorf("seed credential %s: %w", credential.ID, err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *Runner) writeAudit(ctx context.Context, event audit.Event) error {
	event.CorrelationID = "seed"
	event.ActorID = "system:seed"
	event.CreatedAt = time.Now().UTC()

	if err := audit.Insert(ctx, r.pool, event); err != nil {
		return fmt.Errorf("seed audit event %s: %w", event.EventType, err)
	}

	return nil
}

func isValidPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= 8 && length <= 256
}
