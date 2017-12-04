package push_handler

import (
    "io/ioutil"

    "net"
    "net/http"

    "crypto/hmac"
    "crypto/sha1"
    "encoding/hex"

    log "github.com/Sirupsen/logrus"

    "github.com/gorilla/mux"
)

type PushHandler struct {
    secret []byte
}

func NewPushHandler(secret []byte) *PushHandler {
    return &PushHandler{secret}
}

func (self *PushHandler) InstallHandlers(router *mux.Router) {
    router.
        Methods("POST").
        Path("/github/{auth_token}").
        HandlerFunc(self.GitHub)
}

func (self *PushHandler) checkGitHubMac(body []byte, messageMAC string) bool {
    if messageMAC[:5] != "sha1=" {
        return false
    }

    // strip off "sha1="
    realMessageMac, err := hex.DecodeString(messageMAC[5:len(messageMAC)])
    if err != nil {
        return false
    }

    mac := hmac.New(sha1.New, self.secret)
    mac.Write(body)
    expectedMAC := mac.Sum(nil)


    return hmac.Equal(realMessageMac, expectedMAC)
}

// https://developer.github.com/v3/activity/events/types/#pushevent
func (self *PushHandler) GitHub(resp http.ResponseWriter, req *http.Request) {
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

    if ! self.checkGitHubMac(body, hubSignature) {
        resp.WriteHeader(http.StatusForbidden)
        return
    }

    resp.WriteHeader(http.StatusAccepted)
}
