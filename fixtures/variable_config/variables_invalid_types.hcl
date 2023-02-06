variable "service_address" {
  type = "string"
  default = "https://my-server.com"
}

variable "service_user" {
  type = invalid
  default = "ec2_user"
}
