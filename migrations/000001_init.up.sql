CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE persons (
    id              TEXT PRIMARY KEY,
    synthetic_nin   TEXT UNIQUE NOT NULL,
    display_name    TEXT NOT NULL,
    status          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT persons_status_check CHECK (
        status IN (
            'ACTIVE',
            'SUSPENDED',
            'DECEASED',
            'DUPLICATE_REVIEW',
            'DISABLED'
        )
    )
);

CREATE TRIGGER persons_set_updated_at
BEFORE UPDATE ON persons
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE organizations (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    primary_type  TEXT NOT NULL,
    status        TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT organizations_primary_type_check CHECK (
        primary_type IN (
            'GOVERNMENT_AGENCY',
            'HEALTH_PROVIDER',
            'EMPLOYER',
            'FINANCIAL_INSTITUTION_SIM',
            'SUPPLIER'
        )
    ),

    CONSTRAINT organizations_status_check CHECK (
        status IN (
            'ACTIVE',
            'SUSPENDED',
            'DISABLED'
        )
    )
);

CREATE TRIGGER organizations_set_updated_at
BEFORE UPDATE ON organizations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE organization_capabilities (
    id               TEXT PRIMARY KEY,
    organization_id  TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    capability       TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT organization_capabilities_capability_check CHECK (
        capability IN (
            'CAN_EMPLOY_PERSONS',
            'CAN_SUBMIT_TAX_CONTRIBUTIONS',
            'CAN_PROVIDE_HEALTH_SERVICES',
            'CAN_RECEIVE_HEALTH_PAYOUTS',
            'CAN_ROUTE_PAYMENTS',
            'CAN_HOLD_RESERVE_ACCOUNT',
            'CAN_OPERATE_GOVERNMENT_SERVICE',
            'CAN_OBSERVE_SECURITY_EVENTS'
        )
    ),

    CONSTRAINT organization_capabilities_unique UNIQUE (organization_id, capability)
);

CREATE TABLE organization_memberships (
    id               TEXT PRIMARY KEY,
    person_id        TEXT NOT NULL REFERENCES persons(id) ON DELETE RESTRICT,
    organization_id  TEXT NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    membership_type  TEXT NOT NULL,
    status           TEXT NOT NULL,
    starts_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at          TIMESTAMPTZ NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT organization_memberships_membership_type_check CHECK (
        membership_type IN (
            'EMPLOYEE',
            'PROVIDER_STAFF',
            'AGENCY_STAFF',
            'FINANCIAL_OPERATOR',
            'AUDITOR',
            'SYSTEM_ADMIN'
        )
    ),

    CONSTRAINT organization_memberships_status_check CHECK (
        status IN (
            'ACTIVE',
            'SUSPENDED',
            'ENDED',
            'PENDING'
        )
    ),

    CONSTRAINT organization_memberships_time_check CHECK (
        ends_at IS NULL OR ends_at >= starts_at
    )
);

CREATE TRIGGER organization_memberships_set_updated_at
BEFORE UPDATE ON organization_memberships
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE UNIQUE INDEX organization_memberships_one_active_idx
ON organization_memberships (person_id, organization_id, membership_type)
WHERE status IN ('ACTIVE', 'PENDING');

CREATE TABLE membership_roles (
    id             TEXT PRIMARY KEY,
    membership_id  TEXT NOT NULL REFERENCES organization_memberships(id) ON DELETE CASCADE,
    role           TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT membership_roles_role_check CHECK (
        role IN (
            'CITIZEN',
            'PROVIDER_CLINICIAN',
            'PROVIDER_CLAIMS_OFFICER',
            'CARE_ADJUDICATOR',
            'TAX_OFFICER',
            'EMPLOYER_SUBMITTER',
            'PAY_OPERATOR',
            'OBSERVE_ANALYST',
            'OBSERVE_AUDITOR',
            'SYSTEM_ADMIN'
        )
    ),

    CONSTRAINT membership_roles_unique UNIQUE (membership_id, role)
);

CREATE TABLE credentials (
    id             TEXT PRIMARY KEY,
    person_id      TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    username       TEXT UNIQUE NOT NULL,
    password_hash  TEXT NOT NULL,
    status         TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT credentials_status_check CHECK (
        status IN (
            'ACTIVE',
            'SUSPENDED',
            'DISABLED'
        )
    )
);

CREATE TRIGGER credentials_set_updated_at
BEFORE UPDATE ON credentials
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE audit_outbox (
    id              TEXT PRIMARY KEY,
    event_type      TEXT NOT NULL,
    source_system   TEXT NOT NULL,
    actor_id        TEXT NULL,
    subject_id      TEXT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT NOT NULL,
    action          TEXT NOT NULL,
    result          TEXT NOT NULL,
    correlation_id  TEXT NULL,
    payload_json    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ NULL
);

CREATE INDEX persons_synthetic_nin_idx
ON persons (synthetic_nin);

CREATE INDEX organizations_primary_type_idx
ON organizations (primary_type);

CREATE INDEX organization_capabilities_organization_id_idx
ON organization_capabilities (organization_id);

CREATE INDEX organization_memberships_person_id_idx
ON organization_memberships (person_id);

CREATE INDEX organization_memberships_organization_id_idx
ON organization_memberships (organization_id);

CREATE INDEX membership_roles_membership_id_idx
ON membership_roles (membership_id);

CREATE INDEX credentials_person_id_idx
ON credentials (person_id);

CREATE INDEX credentials_username_idx
ON credentials (username);

CREATE INDEX audit_outbox_created_at_idx
ON audit_outbox (created_at);

CREATE INDEX audit_outbox_published_at_idx
ON audit_outbox (published_at);

CREATE INDEX audit_outbox_event_type_idx
ON audit_outbox (event_type);

CREATE INDEX audit_outbox_correlation_id_idx
ON audit_outbox (correlation_id);