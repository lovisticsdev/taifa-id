ALTER TABLE organization_capabilities
	DROP CONSTRAINT IF EXISTS organization_capabilities_capability_check;

ALTER TABLE organization_capabilities
	ADD CONSTRAINT organization_capabilities_capability_check
	CHECK (
		capability IN (
			'CAN_EMPLOY_PERSONS',
			'CAN_SUBMIT_TAX_CONTRIBUTIONS',
			'CAN_PROVIDE_HEALTH_SERVICES',
			'CAN_RECEIVE_HEALTH_PAYOUTS',
			'CAN_ROUTE_PAYMENTS',
			'CAN_HOLD_RESERVE_ACCOUNT',
			'CAN_OPERATE_GOVERNMENT_SERVICE'
		)
	);