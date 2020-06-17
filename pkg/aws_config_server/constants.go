package aws_config_server

const (
	// aws iam role tag that will make us skip over this role
	skipRolesTagKey = "aws-oidc/skip-role"

	// helps distinguish AWS "AccessDenied" errors
	ignoreAWSError = "AccessDenied"
)
