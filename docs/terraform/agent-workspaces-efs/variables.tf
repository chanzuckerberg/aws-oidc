variable "cluster_name" {
  description = "EKS cluster whose nodes mount the agent workspaces filesystem."
  type        = string
  default     = "dev-central"
}

variable "region" {
  description = "Region holding the cluster and the filesystem."
  type        = string
  default     = "us-west-2"
}

variable "account_id" {
  description = "Account the filesystem is created in. Guards against applying to the wrong account."
  type        = string
  default     = "471112759938"
}

variable "aws_profile" {
  description = "Local AWS profile used for the manual apply."
  type        = string
  default     = "core-platform-nonprod"
}

variable "node_security_group_name" {
  description = "Security group carried by the cluster's EC2 nodes. Defaults to <cluster_name>-node-sg."
  type        = string
  default     = null
}

variable "storage_class_name" {
  description = "StorageClass the aws-oidc operator provisions agent workspace claims from."
  type        = string
  default     = "efs-agent-workspaces"
}

variable "workspace_uid" {
  description = "POSIX uid and gid the access point root is owned by. Must match the thread pod's securityContext."
  type        = number
  default     = 1000
}

variable "size_alarm_bytes" {
  description = "StorageBytes level the growth alarm fires at. EFS has no hard quota, so this is the only backstop."
  type        = number
  default     = 107374182400
}

variable "tags" {
  description = "Tags applied to every AWS resource."
  type        = map(string)
  default = {
    project   = "core-platform"
    env       = "dev"
    service   = "aws-oidc"
    owner     = "core-platform"
    managedBy = "terraform"
  }
}
