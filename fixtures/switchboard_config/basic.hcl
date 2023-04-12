switchboard {
  version = "~> 1.0"

  required_provider "test" {
    version = "1.0.0"
    source = "github.com/switchboard-org/provider-test"
  }

  required_provider "test_two" {
    version = "1.0.0"
    source = "github.com/switchboard-org/provider-test-two"
  }
}