# Agent workspaces EFS

Scratch Terraform for the EFS filesystem that agent threads mount at `/workspace` and `/shared`. It lives here so it can be applied by hand against `dev-central` and proved to work before it is adopted by a repository that applies Terraform through CI. It is not applied by any pipeline and it is not the long-term home for this code.

State is local. Nothing here is shared, so `terraform.tfstate` is gitignored and the filesystem this creates is disposable. Destroy and recreate it rather than trying to preserve it.

## What it creates

One `aws_efs_file_system` with elastic throughput, encrypted at rest. A security group allowing NFS on port 2049 from the cluster's nodes. One mount target per availability zone. The `efs-agent-workspaces` StorageClass. A CloudWatch alarm on `StorageBytes`.

The StorageClass uses `provisioningMode: efs-ap`, so the EFS CSI driver creates a fresh access point for every PVC. The operator creates one ReadWriteMany PVC per agent, so each agent gets its own access point and cannot reach another agent's files. Access point roots are owned by uid and gid 1000 with mode 750, matching the `runAsUser`, `runAsGroup` and `fsGroup` in the thread pod spec.

## Facts about dev-central this relies on

The EFS CSI driver is already installed on `dev-central`. It is a Helm release in `kube-system`, not an EKS addon, so it does not show up in `aws eks list-addons`. Both `efs-csi-controller-sa` and `efs-csi-node-sa` are bound to the IRSA role `aws-efs-csi-driver-20250826192017736900000018`. This Terraform therefore installs no driver.

That role can call `CreateAccessPoint` on any filesystem in the account, on the condition that the request carries the tag `efs.csi.aws.com/cluster=true`. The CSI driver sets that tag on the access points it creates, so dynamic provisioning works against a brand new filesystem with no IAM change. Do not attach an EFS filesystem policy without rechecking this, because the role's `ClientMount` and `ClientWrite` grants are conditioned on `AccessedViaMountTarget`.

The cluster's four subnets are all private and each sits in a different availability zone. EFS allows only one mount target per zone, so the configuration picks one subnet per zone rather than iterating the subnet list directly. That keeps the apply working if a zone ever gains a second subnet.

EC2 nodes carry two security groups, `dev-central-node-sg` and the EKS-managed `eks-cluster-sg-dev-central-*`. Fargate pod network interfaces carry only the cluster security group. The ingress rule is created for both so a thread pod can mount regardless of where it is scheduled.

## Applying

```bash
cd docs/terraform/agent-workspaces-efs
terraform init
terraform plan
terraform apply
```

The AWS profile, account, region and cluster all default to `dev-central` in `core-platform-nonprod`. The provider pins `allowed_account_ids`, so an apply with the wrong credentials fails rather than creating a filesystem somewhere unexpected.

## Verifying a pod can mount it

Confirm the StorageClass and the driver agree, then let the operator create a real claim by giving an agent a runtime and a thread:

```bash
kubectl get storageclass efs-agent-workspaces
kubectl get pvc -n <namespace>
kubectl get pv
```

A claim stuck in `Pending` is the interesting case. Read the events on it and the controller log:

```bash
kubectl describe pvc -n <namespace> <agent>-workspace
kubectl -n kube-system logs deploy/efs-csi-controller -c csi-provisioner
```

Once a pod is running, check that both mounts landed and that uid 1000 can write to them:

```bash
kubectl -n <namespace> exec <pod> -- sh -c 'id; mount | grep efs; touch /workspace/ok /shared/ok && ls -la /workspace /shared'
```

A write that fails with permission denied means the access point POSIX user and the pod's security context disagree. Compare `workspace_uid` here with the `RunAsUser` in `internal/agentpod/statefulset.go`.

## Where this should end up

The natural home is the `eks` workspace in core-platform-infra, which already owns the cluster and can read the VPC and worker security group from its own module outputs instead of the data source lookups used here. See [core-platform-infra#574](https://github.com/chanzuckerberg/core-platform-infra/pull/574) for that version. Whichever way it lands, this directory should be deleted and the filesystem created here destroyed.
