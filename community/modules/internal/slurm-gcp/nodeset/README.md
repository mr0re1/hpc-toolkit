<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
Copyright 2025 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.3 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 4.84 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | >= 4.84 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_template"></a> [template](#module\_template) | ../instance_template | n/a |

## Resources

| Name | Type |
|------|------|
| [google_storage_bucket_object.config](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/storage_bucket_object) | resource |
| [google_storage_bucket_object.startup_scripts](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/storage_bucket_object) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_nodeset"></a> [nodeset](#input\_nodeset) | Define nodeset. | <pre>object({<br/>    node_count_static      = optional(number, 0)<br/>    node_count_dynamic_max = optional(number, 1)<br/>    node_conf              = optional(map(string), {})<br/>    nodeset_name           = string<br/>    additional_disks = optional(list(object({<br/>      disk_name                  = optional(string)<br/>      device_name                = optional(string)<br/>      disk_size_gb               = optional(number)<br/>      disk_type                  = optional(string)<br/>      disk_labels                = optional(map(string), {})<br/>      auto_delete                = optional(bool, true)<br/>      boot                       = optional(bool, false)<br/>      disk_resource_manager_tags = optional(map(string), {})<br/>    })), [])<br/>    bandwidth_tier                   = optional(string, "platform_default")<br/>    can_ip_forward                   = optional(bool, false)<br/>    disk_auto_delete                 = optional(bool, true)<br/>    disk_labels                      = optional(map(string), {})<br/>    disk_resource_manager_tags       = optional(map(string), {})<br/>    disk_size_gb                     = optional(number)<br/>    disk_type                        = optional(string)<br/>    enable_confidential_vm           = optional(bool, false)<br/>    enable_placement                 = optional(bool, false)<br/>    placement_max_distance           = optional(number, null)<br/>    enable_oslogin                   = optional(bool, true)<br/>    enable_shielded_vm               = optional(bool, false)<br/>    enable_maintenance_reservation   = optional(bool, false)<br/>    enable_opportunistic_maintenance = optional(bool, false)<br/>    gpu = optional(object({<br/>      count = number<br/>      type  = string<br/>    }))<br/>    dws_flex = object({<br/>      enabled          = bool<br/>      max_run_duration = number<br/>      use_job_duration = bool<br/>    })<br/>    labels       = optional(map(string), {})<br/>    machine_type = optional(string)<br/>    advanced_machine_features = object({<br/>      enable_nested_virtualization = optional(bool)<br/>      threads_per_core             = optional(number)<br/>      turbo_mode                   = optional(string)<br/>      visible_core_count           = optional(number)<br/>      performance_monitoring_unit  = optional(string)<br/>      enable_uefi_networking       = optional(bool)<br/>    })<br/>    maintenance_interval     = optional(string)<br/>    instance_properties_json = string<br/>    metadata                 = optional(map(string), {})<br/>    min_cpu_platform         = optional(string)<br/>    network_tier             = optional(string, "STANDARD")<br/>    network_storage = optional(list(object({<br/>      server_ip             = string<br/>      remote_mount          = string<br/>      local_mount           = string<br/>      fs_type               = string<br/>      mount_options         = string<br/>      client_install_runner = optional(map(string))<br/>      mount_runner          = optional(map(string))<br/>    })), [])<br/>    on_host_maintenance   = optional(string)<br/>    preemptible           = optional(bool, false)<br/>    region                = optional(string)<br/>    resource_manager_tags = optional(map(string), {})<br/>    service_account = optional(object({<br/>      email  = optional(string)<br/>      scopes = optional(list(string), ["https://www.googleapis.com/auth/cloud-platform"])<br/>    }))<br/>    shielded_instance_config = optional(object({<br/>      enable_integrity_monitoring = optional(bool, true)<br/>      enable_secure_boot          = optional(bool, true)<br/>      enable_vtpm                 = optional(bool, true)<br/>    }))<br/>    source_image_family  = optional(string)<br/>    source_image_project = optional(string)<br/>    source_image         = optional(string)<br/>    subnetwork_self_link = string<br/>    additional_networks = optional(list(object({<br/>      network            = string<br/>      subnetwork         = string<br/>      subnetwork_project = string<br/>      network_ip         = string<br/>      nic_type           = string<br/>      stack_type         = string<br/>      queue_count        = number<br/>      access_config = list(object({<br/>        nat_ip       = string<br/>        network_tier = string<br/>      }))<br/>      ipv6_access_config = list(object({<br/>        network_tier = string<br/>      }))<br/>      alias_ip_range = list(object({<br/>        ip_cidr_range         = string<br/>        subnetwork_range_name = string<br/>      }))<br/>    })))<br/>    access_config = optional(list(object({<br/>      nat_ip       = string<br/>      network_tier = string<br/>    })))<br/>    spot               = optional(bool, false)<br/>    tags               = optional(list(string), [])<br/>    termination_action = optional(string)<br/>    reservation_name   = optional(string)<br/>    future_reservation = string<br/><br/>    zone_target_shape = string<br/>    zone_policy_allow = set(string)<br/>    zone_policy_deny  = set(string)<br/>  })</pre> | n/a | yes |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | Project ID to create resources in. | `string` | n/a | yes |
| <a name="input_slurm_bucket_dir"></a> [slurm\_bucket\_dir](#input\_slurm\_bucket\_dir) | Cluster name | `string` | n/a | yes |
| <a name="input_slurm_bucket_name"></a> [slurm\_bucket\_name](#input\_slurm\_bucket\_name) | Cluster name | `string` | n/a | yes |
| <a name="input_slurm_bucket_path"></a> [slurm\_bucket\_path](#input\_slurm\_bucket\_path) | Cluster name | `string` | n/a | yes |
| <a name="input_slurm_cluster_name"></a> [slurm\_cluster\_name](#input\_slurm\_cluster\_name) | Cluster name | `string` | n/a | yes |
| <a name="input_startup_scripts"></a> [startup\_scripts](#input\_startup\_scripts) | List of scripts to be ran on nodeset VMs startup. | <pre>list(object({<br/>    filename = string<br/>    content  = string<br/>  }))</pre> | `[]` | no |
| <a name="input_universe_domain"></a> [universe\_domain](#input\_universe\_domain) | Domain address for alternate API universe | `string` | `"googleapis.com"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_service_account"></a> [service\_account](#output\_service\_account) | !!! |
<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
