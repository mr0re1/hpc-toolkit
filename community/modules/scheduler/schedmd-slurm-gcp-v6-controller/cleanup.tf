# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Ordinary nodesets
locals {
  cleanup_nodeset_depenencies = {
    for ns in var.nodeset : ns.nodeset_name => flatten([
      ns.subnetwork_self_link,
      [for an in ns.additional_networks : an.subnetwork]
    ])
  }

}

resource "null_resource" "cleanup_nodeset_depenencies" {
  for_each = local.cleanup_nodeset_depenencies
  count    = length(each.value)
}

resource "null_resource" "cleanup_nodeset" {
  for_each = var.enable_cleanup_compute ? local.cleanup_nodeset_depenencies : {}

  triggers = {
    project_id               = var.project_id
    cluster_name             = local.slurm_cluster_name
    nodeset_name             = each.key
    universe_domain          = var.universe_domain
    compute_endpoint_version = var.endpoint_versions.compute
    gcloud_path_override     = var.gcloud_path_override
  }

  provisioner "local-exec" {
    command = "/bin/bash ${path.module}/scripts/cleanup_compute.sh ${self.triggers.project_id} ${self.triggers.cluster_name} ${self.triggers.nodeset_name} ${self.triggers.universe_domain} ${self.triggers.compute_endpoint_version} ${self.triggers.gcloud_path_override}"
    when    = destroy
  }

  # Ensure that clean up is done before attempt to delete the networks
  depends_on = [null_resource.cleanup_nodeset_depenencies[each.key]]
}


# TPU nodesets
resource "null_resource" "cleanup_nodeset_tpu" {
  for_each = var.enable_cleanup_compute ? var.nodeset_tpu : []

  triggers = {
    project_id               = var.project_id
    cluster_name             = local.slurm_cluster_name
    nodeset_name             = each.value.nodeset_name
    universe_domain          = var.universe_domain
    compute_endpoint_version = var.endpoint_versions.compute
    gcloud_path_override     = var.gcloud_path_override
  }

  provisioner "local-exec" {
    command = "/bin/bash ${path.module}/scripts/cleanup_compute.sh ${self.triggers.project_id} ${self.triggers.cluster_name} ${self.triggers.nodeset_name} ${self.triggers.universe_domain} ${self.triggers.compute_endpoint_version} ${self.triggers.gcloud_path_override}"
    when    = destroy
  }

  # Ensure that clean up is done before attempt to delete the networks
  depends_on = [each.value.subnetwork]
}
