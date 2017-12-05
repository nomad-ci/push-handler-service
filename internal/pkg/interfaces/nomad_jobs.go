package interfaces

import (
    "github.com/hashicorp/nomad/api"
)

type NomadJobs interface {
    // dispatches a parameterized job
    Dispatch(jobID string, meta map[string]string, payload []byte, q *api.WriteOptions) (*api.JobDispatchResponse, *api.WriteMeta, error)
}
