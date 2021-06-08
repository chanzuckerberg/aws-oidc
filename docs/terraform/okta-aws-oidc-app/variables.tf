variable app_name {
  type        = string
  description = "The name to give this okta app."
}

variable authorized_groups {
  type        = list(string)
  description = "A list of authorized Okta group ids"
  default     = []
}
