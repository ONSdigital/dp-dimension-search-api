package models

const (
	AuditTaskGetSearch   = "getSearch"
	AuditTaskCreateIndex = "createSearchIndex"
	AuditTaskDeleteIndex = "deleteSearchIndex"

	AuditActionAttempted    = "attempted"
	AuditActionSuccessful   = "successful"
	AuditActionUnsuccessful = "unsuccessful"

	AuditActionAttemptedErr    = "failed to audit action attempted event, returning internal server error"
	AuditActionUnsuccessfulErr = "failed to audit action unsuccessful event"
	AuditActionSuccessfulErr   = "failed to audit action successful event"

	Scenario_attemptOnly       = "mock audit a user attempted an action"
	Scenario_attemptAndSucceed = "mock audit a user attempted an action that will be successful"
	Scenario_attemptAndFail    = "mock audit a user attempted an action that will be unsuccessful"
)
