package push_handler

// the push handlers consume requests like /notify/push/github/<token>.  the
// <token> is looked up in Vault, and the payload is used to validate the
// request.  in the case of github, the hmac secret is contained in the Vault
// secret and shared with github, which uses it to sign the payload.

import (
    "fmt"
    "io/ioutil"

    "net"
    "net/http"
    "path"

    "crypto/hmac"
    "crypto/sha1"
    "encoding/hex"
    "encoding/json"

    log "github.com/Sirupsen/logrus"

    "github.com/gorilla/mux"
    "github.com/google/go-github/github"

    "github.com/nomad-ci/push-handler-service/internal/pkg/interfaces"
    "github.com/nomad-ci/push-handler-service/internal/pkg/structs"
)

// error returned when request preflight fails
type preflightError struct {
    msg string
    statusCode int
}

func newPreflightError(msg string, statusCode int) *preflightError {
    return &preflightError{msg, statusCode}
}

func (self *preflightError) Error() string {
    return self.msg
}

type PushHandler struct {
    vault              interfaces.VaultLogical
    webhookTokenPrefix string
    nomad              interfaces.NomadJobs
    dispatchId         string
}

func NewPushHandler(
    vault interfaces.VaultLogical,
    tokenPrefix string,
    nomad interfaces.NomadJobs,
    dispatchId string,
) *PushHandler {
    return &PushHandler{
        vault:              vault,
        webhookTokenPrefix: tokenPrefix,
        nomad:              nomad,
        dispatchId:         dispatchId,
    }
}

func (self *PushHandler) InstallHandlers(router *mux.Router) {
    router.
        Methods("POST").
        Path("/github/{auth_token}").
        Headers(
            "Content-Type", "application/json",
            "X-Github-Event", "push",
        ).
        HandlerFunc(self.GitHubPushEvent)

    router.
        Methods("POST").
        Path("/github/{auth_token}").
        Headers(
            "Content-Type", "application/json",
            "X-Github-Event", "ping",
        ).
        HandlerFunc(self.GitHubPingEvent)
}

func checkGitHubMac(body []byte, secret, messageMAC string) bool {
    if messageMAC[:5] != "sha1=" {
        return false
    }

    // strip off "sha1="
    realMessageMac, err := hex.DecodeString(messageMAC[5:len(messageMAC)])
    if err != nil {
        return false
    }

    mac := hmac.New(sha1.New, []byte(secret))
    mac.Write(body)
    expectedMAC := mac.Sum(nil)


    return hmac.Equal(realMessageMac, expectedMAC)
}

func (self *PushHandler) preflightGitHubEvent(resp http.ResponseWriter, req *http.Request) ([]byte, *log.Entry, *preflightError) {
    vars := mux.Vars(req)

    var err error
    var remoteAddr string

    if xff, ok := req.Header["X-Forwarded-For"]; ok {
        remoteAddr = xff[0]
    } else {
        remoteAddr, _, err = net.SplitHostPort(req.RemoteAddr)
        if err != nil {
            log.Warnf("unable to parse RemoteAddr '%s': %s", remoteAddr, err)
            remoteAddr = req.RemoteAddr
        }
    }

    logEntry := log.
        WithField("remote_ip", remoteAddr).
        WithField("provider", "github").
        WithField("auth_token", vars["auth_token"])

    // https://developer.github.com/webhooks/securing/
    secret, _ := self.vault.Read(path.Join(self.webhookTokenPrefix, "github", vars["auth_token"]))
    if secret == nil {
        return nil, logEntry, newPreflightError(fmt.Sprintf("unauthorized webhook %s", vars["auth_token"]), http.StatusNotFound)
    }

    hmacSecret := secret.Data["secret"].(string)

    var hubSignature string
    if xhs, ok := req.Header["X-Hub-Signature"]; ok {
        hubSignature = xhs[0]
    } else {
        return nil, logEntry, newPreflightError("no X-Hub-Signature header", http.StatusBadRequest)
    }

    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        return nil, logEntry, newPreflightError(fmt.Sprintf("unable to read body: %s", err), http.StatusBadRequest)
    }

    if ! checkGitHubMac(body, hmacSecret, hubSignature) {
        return nil, logEntry, newPreflightError("bad payload signature", http.StatusForbidden)
    }

    return body, logEntry, nil
}

// https://developer.github.com/v3/activity/events/types/#pushevent
func (self *PushHandler) GitHubPushEvent(resp http.ResponseWriter, req *http.Request) {
    body, logEntry, preflightErr := self.preflightGitHubEvent(resp, req)
    if preflightErr != nil {
        logEntry.Error(preflightErr.msg)
        resp.WriteHeader(preflightErr.statusCode)
        return
    }

    var payload github.PushEvent
    err := json.Unmarshal(body, &payload)
    if err != nil {
        logEntry.Errorf("unable to unmarshal body: %s", err)
        resp.WriteHeader(http.StatusBadRequest)
        return
    }

    // create payload for dispatch
    dispatchBytes, err := json.Marshal(structs.CloneDispatchPayload{
        CloneURL: *payload.Repo.CloneURL,
        Ref:      *payload.Ref,
        SHA:      *payload.After,
    })

    if err != nil {
        logEntry.Errorf("unable to marshal dispatch payload: %s", err)
        resp.WriteHeader(http.StatusInternalServerError)
        return
    }

    // actually dispatch the job to nomad
    dispatchResp, _, err := self.nomad.Dispatch(
        self.dispatchId,
        map[string]string{},
        dispatchBytes,
        nil,
    )

    logEntry.Infof("dispatched %s with eval %s", dispatchResp.DispatchedJobID, dispatchResp.EvalID)

    resp.WriteHeader(http.StatusAccepted)
}

// https://developer.github.com/webhooks/#ping-event
func (self *PushHandler) GitHubPingEvent(resp http.ResponseWriter, req *http.Request) {
    body, logEntry, preflightErr := self.preflightGitHubEvent(resp, req)
    if preflightErr != nil {
        logEntry.Error(preflightErr.msg)
        resp.WriteHeader(preflightErr.statusCode)
        return
    }

    var payload github.PingEvent
    err := json.Unmarshal(body, &payload)
    if err != nil {
        logEntry.Errorf("unable to unmarshal body: %s", err)
        resp.WriteHeader(http.StatusBadRequest)
        return
    }

    logEntry.Infof(
        "ping received for hook %d, %s, %s",
        payload.Hook.ID,
        *payload.Hook.Name,
        *payload.Hook.URL,
    )

    resp.WriteHeader(http.StatusNoContent)
}
