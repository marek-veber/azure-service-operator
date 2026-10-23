---
title: Temporary compatibility for legacy ARO HCP SWIFT clusters
weight: 100
---

This release temporarily serves the `2024-06-10-preview` HCP Kubernetes APIs so ASO can
read and delete clusters, node pools, and external auth resources created with that API.
The CRD storage version remains `v20260901previewstorage`.

Existing SWIFT clusters do not have a VNet integration subnet. The 2025 Kubernetes API
requires `spec.properties.platform.vnetIntegrationSubnetReference`, although the Azure
resource has no such subnet. A syntactically valid placeholder reference can satisfy
the Kubernetes schema, but it **must never be sent to Azure**. Do not create a VNet or
subnet for the placeholder.

For each legacy HCP cluster, node pool, and external auth object, set
`serviceoperator.azure.com/reconcile-policy: observe-delete` before changing its API
version. On a cluster with a placeholder subnet, also set
`serviceoperator.azure.com/legacy-swift-placeholder: "true"`. ASO then refreshes status
with ARM GET and permits ARM DELETE, but sends no ARM PUT. A successful `Ready` condition
under this policy confirms observation, not that spec changes were applied.

After installing this compatibility release, identify objects that still record the
2024 original version:

```bash
for resource in \
  hcpopenshiftclusters \
  hcpopenshiftclustersnodepools \
  hcpopenshiftclustersexternalauths; do
  oc get "${resource}.v20260901previewstorage.redhatopenshift.azure.com" -A -o json |
    jq -r --arg resource "$resource" '.items[] |
      select(.spec.originalVersion == "v1api20240610preview") |
      [$resource, .metadata.namespace, .metadata.name] | @tsv'
done
```

For each listed object, add the policy annotation, fetch it through the
`v1api20251223preview` endpoint, and review the resulting manifest. A 2024 SWIFT
cluster also needs a placeholder `vnetIntegrationSubnetReference.armId` under
`spec.properties.platform` and the placeholder annotation. Supply other fields that
the 2025 schema requires from the actual ARM resource; do not invent operational
values. Keep the existing name, namespace, UID, and resourceVersion, remove `status`
and `metadata.managedFields`, then use `oc replace --dry-run=server -f` followed by
`oc replace -f` on the reviewed 2025 manifest. Replacing through the 2025 endpoint
records `v1api20251223preview` as `spec.originalVersion` in the storage object.

```bash
# Set these from one row of the output above.
resource=hcpopenshiftclusters
namespace=<namespace>
name=<name>
oc annotate "${resource}.v20260901previewstorage.redhatopenshift.azure.com" \
  -n "$namespace" "$name" serviceoperator.azure.com/reconcile-policy=observe-delete --overwrite
# Mark the placeholder on a legacy HCP cluster:
if [ "$resource" = hcpopenshiftclusters ]; then
  oc annotate "${resource}.v20260901previewstorage.redhatopenshift.azure.com" \
    -n "$namespace" "$name" serviceoperator.azure.com/legacy-swift-placeholder=true --overwrite
fi
oc get "${resource}.v1api20251223preview.redhatopenshift.azure.com" \
  -n "$namespace" "$name" -o json |
  jq 'del(.status, .metadata.managedFields)' > legacy-2025.json
# Review and complete legacy-2025.json, then:
oc replace --dry-run=server -f legacy-2025.json
oc replace -f legacy-2025.json
```

Verify each object's `spec.originalVersion` through the storage endpoint, and verify
ASO's status reads. Only after every object formerly stored as 2024 has been rewritten
and the CRD's `status.storedVersions` is accurate can a later release remove the 2024
API. `status.storedVersions` alone does not prove which version any individual object
has in etcd; do not clear it solely because an object is readable through 2025.
