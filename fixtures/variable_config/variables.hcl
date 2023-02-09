variable "service_address" {
  type = string
  default = "https://my-server.com"
}

variable "service_active" {
  type = boolean
}

variable "service_password" {
  type = number
  default = 1
}

variable "service_user" {
  type = string
  default = "mary"
}

variable "service_other" {
  type = number
}