output "file_system_id" {
  description = "EFS filesystem backing every agent workspace claim."
  value       = aws_efs_file_system.agent_workspaces.id
}

output "file_system_dns_name" {
  description = "DNS name to mount by hand when debugging a workspace pod."
  value       = aws_efs_file_system.agent_workspaces.dns_name
}

output "security_group_id" {
  description = "Security group on the mount targets."
  value       = aws_security_group.agent_workspaces.id
}

output "mount_target_subnets" {
  description = "Subnet chosen for the mount target in each availability zone."
  value       = local.mount_target_subnets
}

output "storage_class_name" {
  description = "Value to pass to the operator's --agent-storage-class flag."
  value       = kubernetes_storage_class_v1.agent_workspaces.metadata[0].name
}
