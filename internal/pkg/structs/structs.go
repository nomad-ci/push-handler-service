package structs

type CloneDispatchPayload struct {
    CloneURL string `json:"clone_url"`
    Ref      string `json:"ref"`
    SHA      string `json:"sha"`
}
