package push_handler_test

import (
	. "github.com/nomad-ci/push-handler-service/internal/app/push_handler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

    "github.com/stretchr/testify/mock"

    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"

    "github.com/gorilla/mux"

    vaultapi "github.com/hashicorp/vault/api"
    nomadapi "github.com/hashicorp/nomad/api"
    "github.com/nomad-ci/push-handler-service/internal/pkg/interfaces"
    "github.com/nomad-ci/push-handler-service/internal/pkg/structs"
)

// actual payloads captured with requestb.in
var githubPushEventExamplePayload string = `{"ref":"refs/heads/master","before":"0000000000000000000000000000000000000000","after":"024acfdef6b2f11d8b9b2d1e49b9dc401e64ffd7","created":true,"deleted":false,"forced":false,"base_ref":null,"compare":"https://github.com/nomad-ci/push-handler-service/compare/0b85e8064939^...024acfdef6b2","commits":[{"id":"0b85e806493942b8e30ee58b5b14de63c908cdd7","tree_id":"4b825dc642cb6eb9a060e54bf8d69288fbee4904","distinct":true,"message":"repo create","timestamp":"2017-12-03T07:41:37-05:00","url":"https://github.com/nomad-ci/push-handler-service/commit/0b85e806493942b8e30ee58b5b14de63c908cdd7","author":{"name":"Brian Lalor","email":"blalor@bravo5.org","username":"blalor"},"committer":{"name":"Brian Lalor","email":"blalor@bravo5.org","username":"blalor"},"added":[],"removed":[],"modified":[]},{"id":"024acfdef6b2f11d8b9b2d1e49b9dc401e64ffd7","tree_id":"345b3baf64a2c3c059ff65c87236a1fd364ca7e6","distinct":true,"message":"dep init","timestamp":"2017-12-04T06:08:08-05:00","url":"https://github.com/nomad-ci/push-handler-service/commit/024acfdef6b2f11d8b9b2d1e49b9dc401e64ffd7","author":{"name":"Brian Lalor","email":"blalor@bravo5.org","username":"blalor"},"committer":{"name":"Brian Lalor","email":"blalor@bravo5.org","username":"blalor"},"added":["Gopkg.lock","Gopkg.toml"],"removed":[],"modified":[]}],"head_commit":{"id":"024acfdef6b2f11d8b9b2d1e49b9dc401e64ffd7","tree_id":"345b3baf64a2c3c059ff65c87236a1fd364ca7e6","distinct":true,"message":"dep init","timestamp":"2017-12-04T06:08:08-05:00","url":"https://github.com/nomad-ci/push-handler-service/commit/024acfdef6b2f11d8b9b2d1e49b9dc401e64ffd7","author":{"name":"Brian Lalor","email":"blalor@bravo5.org","username":"blalor"},"committer":{"name":"Brian Lalor","email":"blalor@bravo5.org","username":"blalor"},"added":["Gopkg.lock","Gopkg.toml"],"removed":[],"modified":[]},"repository":{"id":113032935,"name":"push-handler-service","full_name":"nomad-ci/push-handler-service","owner":{"name":"nomad-ci","email":null,"login":"nomad-ci","id":34209530,"avatar_url":"https://avatars1.githubusercontent.com/u/34209530?v=4","gravatar_id":"","url":"https://api.github.com/users/nomad-ci","html_url":"https://github.com/nomad-ci","followers_url":"https://api.github.com/users/nomad-ci/followers","following_url":"https://api.github.com/users/nomad-ci/following{/other_user}","gists_url":"https://api.github.com/users/nomad-ci/gists{/gist_id}","starred_url":"https://api.github.com/users/nomad-ci/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/nomad-ci/subscriptions","organizations_url":"https://api.github.com/users/nomad-ci/orgs","repos_url":"https://api.github.com/users/nomad-ci/repos","events_url":"https://api.github.com/users/nomad-ci/events{/privacy}","received_events_url":"https://api.github.com/users/nomad-ci/received_events","type":"Organization","site_admin":false},"private":false,"html_url":"https://github.com/nomad-ci/push-handler-service","description":null,"fork":false,"url":"https://github.com/nomad-ci/push-handler-service","forks_url":"https://api.github.com/repos/nomad-ci/push-handler-service/forks","keys_url":"https://api.github.com/repos/nomad-ci/push-handler-service/keys{/key_id}","collaborators_url":"https://api.github.com/repos/nomad-ci/push-handler-service/collaborators{/collaborator}","teams_url":"https://api.github.com/repos/nomad-ci/push-handler-service/teams","hooks_url":"https://api.github.com/repos/nomad-ci/push-handler-service/hooks","issue_events_url":"https://api.github.com/repos/nomad-ci/push-handler-service/issues/events{/number}","events_url":"https://api.github.com/repos/nomad-ci/push-handler-service/events","assignees_url":"https://api.github.com/repos/nomad-ci/push-handler-service/assignees{/user}","branches_url":"https://api.github.com/repos/nomad-ci/push-handler-service/branches{/branch}","tags_url":"https://api.github.com/repos/nomad-ci/push-handler-service/tags","blobs_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/blobs{/sha}","git_tags_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/tags{/sha}","git_refs_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/refs{/sha}","trees_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/trees{/sha}","statuses_url":"https://api.github.com/repos/nomad-ci/push-handler-service/statuses/{sha}","languages_url":"https://api.github.com/repos/nomad-ci/push-handler-service/languages","stargazers_url":"https://api.github.com/repos/nomad-ci/push-handler-service/stargazers","contributors_url":"https://api.github.com/repos/nomad-ci/push-handler-service/contributors","subscribers_url":"https://api.github.com/repos/nomad-ci/push-handler-service/subscribers","subscription_url":"https://api.github.com/repos/nomad-ci/push-handler-service/subscription","commits_url":"https://api.github.com/repos/nomad-ci/push-handler-service/commits{/sha}","git_commits_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/commits{/sha}","comments_url":"https://api.github.com/repos/nomad-ci/push-handler-service/comments{/number}","issue_comment_url":"https://api.github.com/repos/nomad-ci/push-handler-service/issues/comments{/number}","contents_url":"https://api.github.com/repos/nomad-ci/push-handler-service/contents/{+path}","compare_url":"https://api.github.com/repos/nomad-ci/push-handler-service/compare/{base}...{head}","merges_url":"https://api.github.com/repos/nomad-ci/push-handler-service/merges","archive_url":"https://api.github.com/repos/nomad-ci/push-handler-service/{archive_format}{/ref}","downloads_url":"https://api.github.com/repos/nomad-ci/push-handler-service/downloads","issues_url":"https://api.github.com/repos/nomad-ci/push-handler-service/issues{/number}","pulls_url":"https://api.github.com/repos/nomad-ci/push-handler-service/pulls{/number}","milestones_url":"https://api.github.com/repos/nomad-ci/push-handler-service/milestones{/number}","notifications_url":"https://api.github.com/repos/nomad-ci/push-handler-service/notifications{?since,all,participating}","labels_url":"https://api.github.com/repos/nomad-ci/push-handler-service/labels{/name}","releases_url":"https://api.github.com/repos/nomad-ci/push-handler-service/releases{/id}","deployments_url":"https://api.github.com/repos/nomad-ci/push-handler-service/deployments","created_at":1512386008,"updated_at":"2017-12-04T11:13:28Z","pushed_at":1512388275,"git_url":"git://github.com/nomad-ci/push-handler-service.git","ssh_url":"git@github.com:nomad-ci/push-handler-service.git","clone_url":"https://github.com/nomad-ci/push-handler-service.git","svn_url":"https://github.com/nomad-ci/push-handler-service","homepage":null,"size":0,"stargazers_count":0,"watchers_count":0,"language":null,"has_issues":true,"has_projects":true,"has_downloads":true,"has_wiki":true,"has_pages":false,"forks_count":0,"mirror_url":null,"archived":false,"open_issues_count":0,"license":null,"forks":0,"open_issues":0,"watchers":0,"default_branch":"master","stargazers":0,"master_branch":"master","organization":"nomad-ci"},"pusher":{"name":"blalor","email":"blalor@bravo5.org"},"organization":{"login":"nomad-ci","id":34209530,"url":"https://api.github.com/orgs/nomad-ci","repos_url":"https://api.github.com/orgs/nomad-ci/repos","events_url":"https://api.github.com/orgs/nomad-ci/events","hooks_url":"https://api.github.com/orgs/nomad-ci/hooks","issues_url":"https://api.github.com/orgs/nomad-ci/issues","members_url":"https://api.github.com/orgs/nomad-ci/members{/member}","public_members_url":"https://api.github.com/orgs/nomad-ci/public_members{/member}","avatar_url":"https://avatars1.githubusercontent.com/u/34209530?v=4","description":null},"sender":{"login":"blalor","id":109915,"avatar_url":"https://avatars0.githubusercontent.com/u/109915?v=4","gravatar_id":"","url":"https://api.github.com/users/blalor","html_url":"https://github.com/blalor","followers_url":"https://api.github.com/users/blalor/followers","following_url":"https://api.github.com/users/blalor/following{/other_user}","gists_url":"https://api.github.com/users/blalor/gists{/gist_id}","starred_url":"https://api.github.com/users/blalor/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/blalor/subscriptions","organizations_url":"https://api.github.com/users/blalor/orgs","repos_url":"https://api.github.com/users/blalor/repos","events_url":"https://api.github.com/users/blalor/events{/privacy}","received_events_url":"https://api.github.com/users/blalor/received_events","type":"User","site_admin":false}}`
var githubWebhookPingExamplePayload string = `{"zen":"Favor focus over features.","hook_id":18642661,"hook":{"type":"Repository","id":18642661,"name":"web","active":true,"events":["push"],"config":{"content_type":"json","insecure_ssl":"0","secret":"********","url":"https://requestb.in/zedrkcze"},"updated_at":"2017-12-04T11:43:17Z","created_at":"2017-12-04T11:43:17Z","url":"https://api.github.com/repos/nomad-ci/push-handler-service/hooks/18642661","test_url":"https://api.github.com/repos/nomad-ci/push-handler-service/hooks/18642661/test","ping_url":"https://api.github.com/repos/nomad-ci/push-handler-service/hooks/18642661/pings","last_response":{"code":null,"status":"unused","message":null}},"repository":{"id":113032935,"name":"push-handler-service","full_name":"nomad-ci/push-handler-service","owner":{"login":"nomad-ci","id":34209530,"avatar_url":"https://avatars1.githubusercontent.com/u/34209530?v=4","gravatar_id":"","url":"https://api.github.com/users/nomad-ci","html_url":"https://github.com/nomad-ci","followers_url":"https://api.github.com/users/nomad-ci/followers","following_url":"https://api.github.com/users/nomad-ci/following{/other_user}","gists_url":"https://api.github.com/users/nomad-ci/gists{/gist_id}","starred_url":"https://api.github.com/users/nomad-ci/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/nomad-ci/subscriptions","organizations_url":"https://api.github.com/users/nomad-ci/orgs","repos_url":"https://api.github.com/users/nomad-ci/repos","events_url":"https://api.github.com/users/nomad-ci/events{/privacy}","received_events_url":"https://api.github.com/users/nomad-ci/received_events","type":"Organization","site_admin":false},"private":false,"html_url":"https://github.com/nomad-ci/push-handler-service","description":null,"fork":false,"url":"https://api.github.com/repos/nomad-ci/push-handler-service","forks_url":"https://api.github.com/repos/nomad-ci/push-handler-service/forks","keys_url":"https://api.github.com/repos/nomad-ci/push-handler-service/keys{/key_id}","collaborators_url":"https://api.github.com/repos/nomad-ci/push-handler-service/collaborators{/collaborator}","teams_url":"https://api.github.com/repos/nomad-ci/push-handler-service/teams","hooks_url":"https://api.github.com/repos/nomad-ci/push-handler-service/hooks","issue_events_url":"https://api.github.com/repos/nomad-ci/push-handler-service/issues/events{/number}","events_url":"https://api.github.com/repos/nomad-ci/push-handler-service/events","assignees_url":"https://api.github.com/repos/nomad-ci/push-handler-service/assignees{/user}","branches_url":"https://api.github.com/repos/nomad-ci/push-handler-service/branches{/branch}","tags_url":"https://api.github.com/repos/nomad-ci/push-handler-service/tags","blobs_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/blobs{/sha}","git_tags_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/tags{/sha}","git_refs_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/refs{/sha}","trees_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/trees{/sha}","statuses_url":"https://api.github.com/repos/nomad-ci/push-handler-service/statuses/{sha}","languages_url":"https://api.github.com/repos/nomad-ci/push-handler-service/languages","stargazers_url":"https://api.github.com/repos/nomad-ci/push-handler-service/stargazers","contributors_url":"https://api.github.com/repos/nomad-ci/push-handler-service/contributors","subscribers_url":"https://api.github.com/repos/nomad-ci/push-handler-service/subscribers","subscription_url":"https://api.github.com/repos/nomad-ci/push-handler-service/subscription","commits_url":"https://api.github.com/repos/nomad-ci/push-handler-service/commits{/sha}","git_commits_url":"https://api.github.com/repos/nomad-ci/push-handler-service/git/commits{/sha}","comments_url":"https://api.github.com/repos/nomad-ci/push-handler-service/comments{/number}","issue_comment_url":"https://api.github.com/repos/nomad-ci/push-handler-service/issues/comments{/number}","contents_url":"https://api.github.com/repos/nomad-ci/push-handler-service/contents/{+path}","compare_url":"https://api.github.com/repos/nomad-ci/push-handler-service/compare/{base}...{head}","merges_url":"https://api.github.com/repos/nomad-ci/push-handler-service/merges","archive_url":"https://api.github.com/repos/nomad-ci/push-handler-service/{archive_format}{/ref}","downloads_url":"https://api.github.com/repos/nomad-ci/push-handler-service/downloads","issues_url":"https://api.github.com/repos/nomad-ci/push-handler-service/issues{/number}","pulls_url":"https://api.github.com/repos/nomad-ci/push-handler-service/pulls{/number}","milestones_url":"https://api.github.com/repos/nomad-ci/push-handler-service/milestones{/number}","notifications_url":"https://api.github.com/repos/nomad-ci/push-handler-service/notifications{?since,all,participating}","labels_url":"https://api.github.com/repos/nomad-ci/push-handler-service/labels{/name}","releases_url":"https://api.github.com/repos/nomad-ci/push-handler-service/releases{/id}","deployments_url":"https://api.github.com/repos/nomad-ci/push-handler-service/deployments","created_at":"2017-12-04T11:13:28Z","updated_at":"2017-12-04T11:13:28Z","pushed_at":"2017-12-04T11:13:29Z","git_url":"git://github.com/nomad-ci/push-handler-service.git","ssh_url":"git@github.com:nomad-ci/push-handler-service.git","clone_url":"https://github.com/nomad-ci/push-handler-service.git","svn_url":"https://github.com/nomad-ci/push-handler-service","homepage":null,"size":0,"stargazers_count":0,"watchers_count":0,"language":null,"has_issues":true,"has_projects":true,"has_downloads":true,"has_wiki":true,"has_pages":false,"forks_count":0,"mirror_url":null,"archived":false,"open_issues_count":0,"license":null,"forks":0,"open_issues":0,"watchers":0,"default_branch":"master"},"sender":{"login":"blalor","id":109915,"avatar_url":"https://avatars0.githubusercontent.com/u/109915?v=4","gravatar_id":"","url":"https://api.github.com/users/blalor","html_url":"https://github.com/blalor","followers_url":"https://api.github.com/users/blalor/followers","following_url":"https://api.github.com/users/blalor/following{/other_user}","gists_url":"https://api.github.com/users/blalor/gists{/gist_id}","starred_url":"https://api.github.com/users/blalor/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/blalor/subscriptions","organizations_url":"https://api.github.com/users/blalor/orgs","repos_url":"https://api.github.com/users/blalor/repos","events_url":"https://api.github.com/users/blalor/events{/privacy}","received_events_url":"https://api.github.com/users/blalor/received_events","type":"User","site_admin":false}}`

var _ = Describe("PushHandler", func() {
    var ph *PushHandler
    var router *mux.Router
    var resp *httptest.ResponseRecorder

    dispatchJobId := "clone-some-repo"

    var mockVaultLogical interfaces.MockVaultLogical
    var mockNomadJobs interfaces.MockNomadJobs

    BeforeEach(func() {
        router = mux.NewRouter()
        resp = httptest.NewRecorder()

        mockVaultLogical = interfaces.MockVaultLogical{}
        mockNomadJobs = interfaces.MockNomadJobs{}

        ph = NewPushHandler(
            &mockVaultLogical,
            "webhook-tokens",
            &mockNomadJobs,
            dispatchJobId,
        )
        ph.InstallHandlers(router.PathPrefix("/notify/push").Subrouter())
    })

    Describe("for GitHub", func() {
        endpoint := "http://example.com/notify/push/github/some-auth-token"

        BeforeEach(func() {
            mockVaultLogical.
                On("Read", "webhook-tokens/github/some-auth-token").
                Return(&vaultapi.Secret{
                    Data: map[string]interface{} {
                        // tied to the payload signatures
                        "secret": "011746565c10e8c64df18d8724bc542da584433c",
                    },
                }, nil)
        })

        It("should handle a ping", func() {
            req, err := http.NewRequest(
                "POST",
                endpoint,
                strings.NewReader(githubWebhookPingExamplePayload),
            )
            Expect(err).ShouldNot(HaveOccurred())

            req.Header.Add("Content-Type", "application/json")
            req.Header.Add("X-Github-Event", "ping")
            req.Header.Add("X-Github-Delivery", "some-uuid")
            req.Header.Add("X-Hub-Signature", "sha1=d9fd3f2b1dd74386ece71aec95df0b442b1a8e61")

            router.ServeHTTP(resp, req)
            Expect(resp.Code).To(Equal(http.StatusNoContent))

            mockVaultLogical.AssertExpectations(GinkgoT())
        })

        It("should handle a push event", func() {
            mockNomadJobs.
                On(
                    "Dispatch",
                    dispatchJobId,
                    map[string]string{},
                    mock.AnythingOfType("[]uint8"), // https://github.com/stretchr/testify/issues/387
                    mock.AnythingOfType("*api.WriteOptions"),
                ).
                Return(
                    &nomadapi.JobDispatchResponse{
                        EvalID: "cafedead-beef-cafe-dead-beefcafedead",
                        DispatchedJobID: dispatchJobId + "/dispatch-1234",
                    },
                    &nomadapi.WriteMeta{},
                    nil,
                )

            req, err := http.NewRequest(
                "POST",
                endpoint,
                strings.NewReader(githubPushEventExamplePayload),
            )
            Expect(err).ShouldNot(HaveOccurred())

            req.Header.Add("Content-Type", "application/json")
            req.Header.Add("X-Github-Event", "push")
            req.Header.Add("X-Github-Delivery", "some-uuid")
            req.Header.Add("X-Hub-Signature", "sha1=93308d96ce42201626ede454a2b420cd21b9df71")

            router.ServeHTTP(resp, req)
            Expect(resp.Code).To(Equal(http.StatusAccepted))

            mockVaultLogical.AssertExpectations(GinkgoT())
            mockNomadJobs.AssertExpectations(GinkgoT())

            // verify payload
            var dispatchPayload structs.CloneDispatchPayload
            Expect(json.Unmarshal(mockNomadJobs.Calls[0].Arguments[2].([]byte), &dispatchPayload)).ShouldNot(HaveOccurred())

            Expect(dispatchPayload).To(Equal(structs.CloneDispatchPayload{
                CloneURL: "https://github.com/nomad-ci/push-handler-service.git",
                Ref:      "refs/heads/master",
                SHA:      "024acfdef6b2f11d8b9b2d1e49b9dc401e64ffd7",
            }))
        })

        It("should return 403 for an invalid signature", func() {
            req, err := http.NewRequest(
                "POST",
                endpoint,
                strings.NewReader(githubPushEventExamplePayload),
            )
            Expect(err).ShouldNot(HaveOccurred())

            req.Header.Add("Content-Type", "application/json")
            req.Header.Add("X-Github-Event", "push")
            req.Header.Add("X-Github-Delivery", "some-uuid")
            req.Header.Add("X-Hub-Signature", "sha1=totallynotavalidsignature")

            router.ServeHTTP(resp, req)
            Expect(resp.Code).To(Equal(http.StatusForbidden))

            mockVaultLogical.AssertExpectations(GinkgoT())
        })
    })

    Describe("for GitHub invalid webhooks", func() {
        endpoint := "http://example.com/notify/push/github/invalid-auth-token"

        BeforeEach(func() {
            mockVaultLogical.
                On("Read", "webhook-tokens/github/invalid-auth-token").
                Return(nil, nil)
        })

        It("should return 404 for an unknown auth token", func() {
            req, err := http.NewRequest(
                "POST",
                endpoint,
                strings.NewReader(githubPushEventExamplePayload),
            )
            Expect(err).ShouldNot(HaveOccurred())

            req.Header.Add("Content-Type", "application/json")
            req.Header.Add("X-Github-Event", "push")
            req.Header.Add("X-Github-Delivery", "some-uuid")
            req.Header.Add("X-Hub-Signature", "sha1=93308d96ce42201626ede454a2b420cd21b9df71")

            router.ServeHTTP(resp, req)
            Expect(resp.Code).To(Equal(http.StatusNotFound))

            mockVaultLogical.AssertExpectations(GinkgoT())
        })

    })
})
