package push_handler

// the push handlers consume requests like /notify/push/github/<token>.  the
// <token> is looked up in Vault, and the payload is used to validate the
// request.  in the case of github, the hmac secret is contained in the Vault
// secret and shared with github, which uses it to sign the payload.

import (
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
)

type PushHandler struct {
    vault              interfaces.VaultLogical
    webhookTokenPrefix string
}

func NewPushHandler(vaultLogical interfaces.VaultLogical, tokenPrefix string) *PushHandler {
    return &PushHandler{
        vault: vaultLogical,
        webhookTokenPrefix: tokenPrefix,
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

func (self *PushHandler) checkGitHubMac(body []byte, secret, messageMAC string) bool {
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

// https://developer.github.com/v3/activity/events/types/#pushevent
func (self *PushHandler) GitHubPushEvent(resp http.ResponseWriter, req *http.Request) {
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

    logEntry := log.WithField("remote_ip", remoteAddr)

    secret, _ := self.vault.Read(path.Join(self.webhookTokenPrefix, "github", vars["auth_token"]))
    if secret == nil {
        logEntry.Warnf("unauthorized webhook %s", vars["auth_token"])
        resp.WriteHeader(http.StatusNotFound)
        return
    }

    hmacSecret := secret.Data["secret"].(string)

    var hubSignature string
    if xhs, ok := req.Header["X-Hub-Signature"]; ok {
        hubSignature = xhs[0]
    } else {
        logEntry.Error("no X-Hub-Signature header")
        resp.WriteHeader(http.StatusBadRequest)
        return
    }

    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        logEntry.Errorf("unable to read body: %s", err)
        http.Error(resp, "unable to read body", http.StatusBadRequest)
        return
    }

    if ! self.checkGitHubMac(body, hmacSecret, hubSignature) {
        resp.WriteHeader(http.StatusForbidden)
        return
    }

    var payload github.PushEvent
    err = json.Unmarshal(body, &payload)
    if err != nil {
        logEntry.Errorf("unable to unmarshal body: %s", err)
        resp.WriteHeader(http.StatusBadRequest)
        return
    }

    resp.WriteHeader(http.StatusAccepted)
}

// https://developer.github.com/v3/activity/events/types/#pushevent
func (self *PushHandler) GitHubPingEvent(resp http.ResponseWriter, req *http.Request) {
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

    logEntry := log.WithField("remote_ip", remoteAddr)

    secret, _ := self.vault.Read(path.Join(self.webhookTokenPrefix, "github", vars["auth_token"]))
    hmacSecret := secret.Data["secret"].(string)

    var hubSignature string
    if xhs, ok := req.Header["X-Hub-Signature"]; ok {
        hubSignature = xhs[0]
    } else {
        logEntry.Error("no X-Hub-Signature header")
        resp.WriteHeader(http.StatusBadRequest)
        return
    }

    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        logEntry.Errorf("unable to read body: %s", err)
        http.Error(resp, "unable to read body", http.StatusBadRequest)
        return
    }

    if ! self.checkGitHubMac(body, hmacSecret, hubSignature) {
        resp.WriteHeader(http.StatusForbidden)
        return
    }

    var payload github.PingEvent
    err = json.Unmarshal(body, &payload)
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
