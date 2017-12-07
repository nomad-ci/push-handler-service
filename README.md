HTTP service to handle push notifications, such as via a GitHub webhook.

## building

    make

## running

Assumes the local Nomad instance has a parameterized job named `clone-source`, and that there's a Vault secret that's been set up with:

    write secret/webhook-tokens/github/some-auth-token \
        secret=011746565c10e8c64df18d8724bc542da584433c

Then:

    work/push-handler-service \
        --vault-token … \
        --webhook-token-prefix secret/webhook-tokens \
        --nomad-addr http://127.0.0.1:4646 \
        --dispatch-job-id clone-source

## examples

### ping

    curl -i \
        -H 'Content-Type: application/json' \
        -H 'X-Github-Event: ping' \
        -H 'X-Hub-Signature: sha1=d9fd3f2b1dd74386ece71aec95df0b442b1a8e61' \
        -d @test/fixtures/ping.json \
        localhost:8080/notify/push/github/some-auth-token

### push

    curl -i \
        -H 'Content-Type: application/json' \
        -H 'X-Github-Event: push' \
        -H 'X-Hub-Signature: sha1=93308d96ce42201626ede454a2b420cd21b9df71' \
        -d @test/fixtures/push.json \
        localhost:8080/notify/push/github/some-auth-token
