data "aws_eks_cluster" "this" {
  name = var.cluster_name
}

data "aws_eks_cluster_auth" "this" {
  name = var.cluster_name
}

data "aws_subnet" "cluster" {
  for_each = toset(data.aws_eks_cluster.this.vpc_config[0].subnet_ids)

  id = each.value
}

data "aws_security_group" "node" {
  vpc_id = data.aws_eks_cluster.this.vpc_config[0].vpc_id

  filter {
    name   = "group-name"
    values = [coalesce(var.node_security_group_name, "${var.cluster_name}-node-sg")]
  }
}

locals {
  subnets_by_az = {
    for id, subnet in data.aws_subnet.cluster : subnet.availability_zone => id...
  }

  mount_target_subnets = {
    for az, ids in local.subnets_by_az : az => sort(ids)[0]
  }

  client_security_groups = toset([
    data.aws_security_group.node.id,
    data.aws_eks_cluster.this.vpc_config[0].cluster_security_group_id,
  ])

  name = "${var.cluster_name}-agent-workspaces"
}

resource "aws_efs_file_system" "agent_workspaces" {
  creation_token   = local.name
  encrypted        = true
  performance_mode = "generalPurpose"
  throughput_mode  = "elastic"

  tags = {
    Name = local.name
  }
}

resource "aws_security_group" "agent_workspaces" {
  name        = "${local.name}-efs"
  description = "NFS from the ${var.cluster_name} nodes to the agent workspaces EFS filesystem"
  vpc_id      = data.aws_eks_cluster.this.vpc_config[0].vpc_id

  tags = {
    Name = "${local.name}-efs"
  }
}

resource "aws_vpc_security_group_ingress_rule" "nfs" {
  for_each = local.client_security_groups

  security_group_id            = aws_security_group.agent_workspaces.id
  description                  = "NFS from ${each.value}"
  from_port                    = 2049
  to_port                      = 2049
  ip_protocol                  = "tcp"
  referenced_security_group_id = each.value
}

resource "aws_efs_mount_target" "agent_workspaces" {
  for_each = local.mount_target_subnets

  file_system_id  = aws_efs_file_system.agent_workspaces.id
  subnet_id       = each.value
  security_groups = [aws_security_group.agent_workspaces.id]
}

resource "kubernetes_storage_class_v1" "agent_workspaces" {
  metadata {
    name = var.storage_class_name
  }

  storage_provisioner    = "efs.csi.aws.com"
  reclaim_policy         = "Delete"
  volume_binding_mode    = "Immediate"
  allow_volume_expansion = false

  parameters = {
    provisioningMode = "efs-ap"
    fileSystemId     = aws_efs_file_system.agent_workspaces.id
    directoryPerms   = "750"
    uid              = tostring(var.workspace_uid)
    gid              = tostring(var.workspace_uid)
  }

  depends_on = [aws_efs_mount_target.agent_workspaces]
}

resource "aws_cloudwatch_metric_alarm" "agent_workspaces_size" {
  alarm_name          = "${local.name}-size"
  alarm_description   = "Agent workspace EFS filesystem has grown past the expected size. No hard quota exists, so cleanup is manual."
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "StorageBytes"
  namespace           = "AWS/EFS"
  period              = 86400
  statistic           = "Maximum"
  threshold           = var.size_alarm_bytes
  treat_missing_data  = "notBreaching"

  dimensions = {
    FileSystemId = aws_efs_file_system.agent_workspaces.id
    StorageClass = "Total"
  }
}
