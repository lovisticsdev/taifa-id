package organization

import (
	"errors"
	"time"
)

type PrimaryType string

const (
	PrimaryTypeGovernmentAgency        PrimaryType = "GOVERNMENT_AGENCY"
	PrimaryTypeHealthProvider          PrimaryType = "HEALTH_PROVIDER"
	PrimaryTypeEmployer                PrimaryType = "EMPLOYER"
	PrimaryTypeFinancialInstitutionSim PrimaryType = "FINANCIAL_INSTITUTION_SIM"
	PrimaryTypeSupplier                PrimaryType = "SUPPLIER"
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusDisabled  Status = "DISABLED"
)

type Capability string

const (
	CapabilityEmployPersons          Capability = "CAN_EMPLOY_PERSONS"
	CapabilitySubmitTaxContributions Capability = "CAN_SUBMIT_TAX_CONTRIBUTIONS"
	CapabilityProvideHealthServices  Capability = "CAN_PROVIDE_HEALTH_SERVICES"
	CapabilityReceiveHealthPayouts   Capability = "CAN_RECEIVE_HEALTH_PAYOUTS"
	CapabilityRoutePayments          Capability = "CAN_ROUTE_PAYMENTS"
	CapabilityHoldReserveAccount     Capability = "CAN_HOLD_RESERVE_ACCOUNT"
	CapabilityOperateGovService      Capability = "CAN_OPERATE_GOVERNMENT_SERVICE"
	CapabilityObserveSecurityEvents  Capability = "CAN_OBSERVE_SECURITY_EVENTS"
)

type Organization struct {
	ID          string
	Name        string
	PrimaryType PrimaryType
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrganizationCapability struct {
	ID             string
	OrganizationID string
	Capability     Capability
	CreatedAt      time.Time
}

var (
	ErrValidation          = errors.New("organization validation failed")
	ErrNotFound            = errors.New("organization not found")
	ErrCapabilityNotFound  = errors.New("organization capability not found")
	ErrDuplicateCapability = errors.New("duplicate organization capability")
)

func IsValidPrimaryType(primaryType PrimaryType) bool {
	switch primaryType {
	case PrimaryTypeGovernmentAgency,
		PrimaryTypeHealthProvider,
		PrimaryTypeEmployer,
		PrimaryTypeFinancialInstitutionSim,
		PrimaryTypeSupplier:
		return true
	default:
		return false
	}
}

func IsValidStatus(status Status) bool {
	switch status {
	case StatusActive,
		StatusSuspended,
		StatusDisabled:
		return true
	default:
		return false
	}
}

func IsValidCapability(capability Capability) bool {
	switch capability {
	case CapabilityEmployPersons,
		CapabilitySubmitTaxContributions,
		CapabilityProvideHealthServices,
		CapabilityReceiveHealthPayouts,
		CapabilityRoutePayments,
		CapabilityHoldReserveAccount,
		CapabilityOperateGovService,
		CapabilityObserveSecurityEvents:
		return true
	default:
		return false
	}
}
