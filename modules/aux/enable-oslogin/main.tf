variable "enable_oslogin" {
  description = "Enable or Disable OS Login with \"ENABLE\" or \"DISABLE\". Set to \"INHERIT\" to inherit project OS Login setting. Values 'true', 'false' about to be deprecated"
  type        = string
  default     = "ENABLE"
  validation {
    condition     = var.enable_oslogin == null ? false : contains([
      "ENABLE", "DISABLE", "INHERIT", "true", "false"], var.enable_oslogin)
    error_message = "Allowed string values for var.enable_oslogin are \"ENABLE\", \"DISABLE\", or \"INHERIT\". Values 'true', 'false' about to be deprecated"
  }
}

locals {
  oslogin_api_values = {
    "DISABLE" = "FALSE"
    "ENABLE"  = "TRUE"
    "true"    = "TRUE"
    "false"   = "FALSE"
  }
  metadata = var.enable_oslogin == "INHERIT" ? {} : { enable-oslogin = lookup(local.oslogin_api_values, var.enable_oslogin, "") }
}

output "metadata" {
  value       = local.metadata
}

output "as_bool" {
  value = var.enable_oslogin == "ENABLE"

  precondition {
    condition     = var.enable_oslogin != "INHERIT"
    error_message = "One shouldn't use enable_oslogin.as_boolean with value = INHERIT"
  }
}