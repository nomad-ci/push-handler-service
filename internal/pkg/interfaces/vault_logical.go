package interfaces

import (
    "github.com/hashicorp/vault/api"
)

type VaultLogical interface {
    Read(path string) (*api.Secret, error)
}
