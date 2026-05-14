package membership

import (
	"errors"
	"time"
)

type Type string

const (
	TypeEmployee          Type = "EMPLOYEE"
	TypeProviderStaff     Type = "PROVIDER_STAFF"
	TypeAgencyStaff       Type = "AGENCY_STAFF"
	TypeFinancialOperator Type = "FINANCIAL_OPERATOR"
	TypeAuditor           Type = "AUDITOR"
	TypeSystemAdmin       Type = "SYSTEM_ADMIN"
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusEnded     Status = "ENDED"
	StatusPending   Status = "PENDING"
)

type Role string

const (
	RoleCitizen               Role = "CITIZEN"
	RoleProviderClinician     Role = "PROVIDER_CLINICIAN"
	RoleProviderClaimsOfficer Role = "PROVIDER_CLAIMS_OFFICER"
	RoleCareAdjudicator       Role = "CARE_ADJUDICATOR"
	RoleTaxOfficer            Role = "TAX_OFFICER"
	RoleEmployerSubmitter     Role = "EMPLOYER_SUBMITTER"
	RolePayOperator           Role = "PAY_OPERATOR"
	RoleObserveAnalyst        Role = "OBSERVE_ANALYST"
	RoleObserveAuditor        Role = "OBSERVE_AUDITOR"
	RoleSystemAdmin           Role = "SYSTEM_ADMIN"
)

type Membership struct {
	ID             string
	PersonID       string
	OrganizationID string
	MembershipType Type
	Status         Status
	StartsAt       time.Time
	EndsAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type MembershipRole struct {
	ID           string
	MembershipID string
	Role         Role
	CreatedAt    time.Time
}

var (
	ErrValidation                = errors.New("membership validation failed")
	ErrNotFound                  = errors.New("membership not found")
	ErrRoleNotFound              = errors.New("membership role not found")
	ErrReferenceNotFound         = errors.New("referenced person or organization not found")
	ErrPersonNotActive           = errors.New("person is not active")
	ErrOrganizationNotActive     = errors.New("organization is not active")
	ErrDuplicateActiveMembership = errors.New("duplicate active membership")
	ErrDuplicateRole             = errors.New("duplicate membership role")
	ErrMembershipNotActive       = errors.New("membership is not active")
)

func IsValidType(membershipType Type) bool {
	switch membershipType {
	case TypeEmployee,
		TypeProviderStaff,
		TypeAgencyStaff,
		TypeFinancialOperator,
		TypeAuditor,
		TypeSystemAdmin:
		return true
	default:
		return false
	}
}

func IsValidStatus(status Status) bool {
	switch status {
	case StatusActive,
		StatusSuspended,
		StatusEnded,
		StatusPending:
		return true
	default:
		return false
	}
}

func IsValidRole(role Role) bool {
	switch role {
	case RoleCitizen,
		RoleProviderClinician,
		RoleProviderClaimsOfficer,
		RoleCareAdjudicator,
		RoleTaxOfficer,
		RoleEmployerSubmitter,
		RolePayOperator,
		RoleObserveAnalyst,
		RoleObserveAuditor,
		RoleSystemAdmin:
		return true
	default:
		return false
	}
}
