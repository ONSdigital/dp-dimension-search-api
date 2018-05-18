package models

const (
	AuditActionGetSearch   = "getSearch"
	AuditActionCreateIndex = "createSearchIndex"
	AuditActionDeleteIndex = "deleteSearchIndex"

	AuditResultAttempted    = "attempted"
	AuditResultSuccessful   = "successful"
	AuditResultUnsuccessful = "unsuccessful"

	Scenario_attemptOnly       = "mock audit a user attempted an action"
	Scenario_attemptAndSucceed = "mock audit a user attempted an action that will be successful"
	Scenario_attemptAndFail    = "mock audit a user attempted an action that will be unsuccessful"
)
