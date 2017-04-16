package bench

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frk/route"
	"github.com/gin-gonic/gin"
	"github.com/pressly/chi"
)

func BenchmarkServeHTTPStatic_Chi(b *testing.B) {
	chiHandlerFunc := func(s string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Handled-By", s)
		}
	}

	router := chi.NewMux()
	for _, a := range githubStaticAPI {
		switch a.Method {
		case "GET":
			router.Get(a.Pattern, chiHandlerFunc(a.Pattern))
		case "PUT":
			router.Put(a.Pattern, chiHandlerFunc(a.Pattern))
		case "POST":
			router.Post(a.Pattern, chiHandlerFunc(a.Pattern))
		case "PATCH":
			router.Patch(a.Pattern, chiHandlerFunc(a.Pattern))
		case "DELETE":
			router.Delete(a.Pattern, chiHandlerFunc(a.Pattern))
		}
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for _, req := range staticBenchRequests {
			r.Method = req.Method
			r.RequestURI = req.Path
			r.URL.Path = req.Path
			router.ServeHTTP(w, r)
		}
	}
}

func BenchmarkServeHTTPStatic_Gin(b *testing.B) {
	ginHandlerFunc := func(s string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Writer.Header().Set("Handled-By", s)
		}
	}

	router := gin.New()
	for _, a := range githubStaticAPI {
		router.Handle(a.Method, a.Pattern, ginHandlerFunc(a.Pattern))
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for _, req := range staticBenchRequests {
			r.Method = req.Method
			r.RequestURI = req.Path
			r.URL.Path = req.Path
			router.ServeHTTP(w, r)
		}
	}
}

func BenchmarkServeHTTPStatic_Frk(b *testing.B) {
	frkHandlerFunc := func(s string) route.HandlerFunc {
		return func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Handled-By", s)
		}
	}

	router := route.NewRouter()
	for _, a := range githubStaticAPI {
		router.HandleFunc(a.Method, a.Pattern, frkHandlerFunc(a.Pattern))
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for _, req := range staticBenchRequests {
			r.Method = req.Method
			r.RequestURI = req.Path
			r.URL.Path = req.Path
			router.ServeHTTP(w, r)
		}
	}
}

type strHandler string

func (h strHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Handled-By", string(h))
}

var staticBenchRequests = func() []benchReq {
	rs := make([]benchReq, len(githubStaticAPI), len(githubStaticAPI))
	for i, a := range githubStaticAPI {
		rs[i].Method = a.Method
		rs[i].Path = a.Pattern
	}
	return rs
}()

type benchReq struct{ Method, Path, Pattern string }

// github's api
var githubStaticAPI = []struct {
	Method  string
	Pattern string
}{
	// auth
	{"GET", "/authorizations"},
	{"GET", "/authorizations/id"},
	{"POST", "/authorizations"},
	{"PUT", "/authorizations/clients/client_id"},
	{"PUT", "/authorizations/clients/client_id/fingerprint"},
	{"PATCH", "/authorizations/id"},
	{"DELETE", "/authorizations/id"},
	{"GET", "/applications/client_id/tokens/access_token"},
	{"POST", "/applications/client_id/tokens/access_token"},
	{"DELETE", "/applications/client_id/tokens"},
	{"DELETE", "/applications/client_id/tokens/access_token"},
	// events
	{"GET", "/events"},
	{"GET", "/repos/owner/repo/events"},
	{"GET", "/networks/owner/repo/events"},
	{"GET", "/orgs/org/events"},
	{"GET", "/users/username/received_events"},
	{"GET", "/users/username/received_events/public"},
	{"GET", "/users/username/events"},
	{"GET", "/users/username/events/public"},
	{"GET", "/users/username/events/orgs/org"},
	// feeds
	{"GET", "/feeds"},
	// notifications
	{"GET", "/notifications"},
	{"GET", "/repos/owner/repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/owner/repo/notifications"},
	{"GET", "/notifications/threads/id"},
	{"PATCH", "/notifications/threads/id"},
	{"GET", "/notifications/threads/id/subscription"},
	{"PUT", "/notifications/threads/id/subscription"},
	{"DELETE", "/notifications/threads/id/subscription"},
	// starring
	{"GET", "/repos/owner/repo/stargazers"},
	{"GET", "/users/username/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/owner/repo"},
	{"PUT", "/user/starred/owner/repo"},
	{"DELETE", "/user/starred/owner/repo"},
	// watching
	{"GET", "/repos/owner/repo/subscribers"},
	{"GET", "/users/username/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/owner/repo/subscription"},
	{"PUT", "/repos/owner/repo/subscription"},
	{"DELETE", "/repos/owner/repo/subscription"},
	// watching (legacy)
	{"GET", "/user/subscriptions/owner/repo"},
	{"PUT", "/user/subscriptions/owner/repo"},
	{"DELETE", "/user/subscriptions/owner/repo"},
	// gists
	{"GET", "/users/username/gists"},
	{"GET", "/gists"},
	{"GET", "/gists/public"},
	{"GET", "/gists/starred"},
	{"GET", "/gists/gist_id"},
	{"POST", "/gists"},
	{"PATCH", "/gists/gist_id"},
	{"GET", "/gists/gist_id/commits"},
	{"PUT", "/gists/gist_id/star"},
	{"DELETE", "/gists/gist_id/star"},
	{"GET", "/gists/gist_id/star"},
	{"POST", "/gists/gist_id/forks"},
	{"GET", "/gists/gist_id/forks"},
	{"DELETE", "/gists/gist_id"},
	// gitst comments
	{"GET", "/gists/gist_id/comments"},
	{"GET", "/gists/gist_id/comments/id"},
	{"POST", "/gists/gist_id/comments"},
	{"PATCH", "/gists/gist_id/comments/id"},
	{"DELETE", "/gists/gist_id/comments/id"},
	// blobs
	{"GET", "/repos/owner/repo/git/blobs/sha"},
	{"POST", "/repos/owner/repo/git/blobs"},
	// commits
	{"GET", "/repos/owner/repo/git/commits/sha"},
	{"POST", "/repos/owner/repo/git/commits"},
	// references
	{"GET", "/repos/owner/repo/git/refs/ref"},
	{"GET", "/repos/owner/repo/git/refs"},
	{"GET", "/repos/owner/repo/git/refs/tags"},
	{"POST", "/repos/owner/repo/git/refs"},
	{"PATCH", "/repos/owner/repo/git/refs/ref"},
	{"DELETE", "/repos/owner/repo/git/refs/ref"},
	// tags
	{"GET", "/repos/owner/repo/git/tags/sha"},
	{"POST", "/repos/owner/repo/git/tags"},
	// trees
	{"GET", "/repos/owner/repo/git/trees/sha"},
	{"POST", "/repos/owner/repo/git/trees"},
	// issues
	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/org/issues"},
	{"GET", "/repos/owner/repo/issues"},
	{"GET", "/repos/owner/repo/issues/issue_number"},
	{"POST", "/repos/owner/repo/issues"},
	{"PATCH", "/repos/owner/repo/issues/issue_number"},
	// issues assignees
	{"GET", "/repos/owner/repo/assignees"},
	{"GET", "/repos/owner/repo/assignees/assignee"},
	// issue comments
	{"GET", "/repos/owner/repo/issues/issue_number/comments"},
	{"GET", "/repos/owner/repo/issues/comments"},
	{"GET", "/repos/owner/repo/issues/comments/id"},
	{"POST", "/repos/owner/repo/issues/issue_number/comments"},
	{"PATCH", "/repos/owner/repo/issues/comments/id"},
	{"DELETE", "/repos/owner/repo/issues/comments/id"},
	// issue events
	{"GET", "/repos/owner/repo/issues/issue_number/events"},
	{"GET", "/repos/owner/repo/issues/events"},
	{"GET", "/repos/owner/repo/issues/events/id"},
	// issue labels
	{"GET", "/repos/owner/repo/labels"},
	{"GET", "/repos/owner/repo/labels/name"},
	{"POST", "/repos/owner/repo/labels"},
	{"PATCH", "/repos/owner/repo/labels/name"},
	{"DELETE", "/repos/owner/repo/labels/name"},
	{"GET", "/repos/owner/repo/issues/issue_number/labels"},
	{"POST", "/repos/owner/repo/issues/issue_number/labels"},
	{"DELETE", "/repos/owner/repo/issues/issue_number/labels/name"},
	{"PUT", "/repos/owner/repo/issues/issue_number/labels"},
	{"DELETE", "/repos/owner/repo/issues/issue_number/labels"},
	{"GET", "/repos/owner/repo/milestones/number/labels"},
	// issue milestones
	{"GET", "/repos/owner/repo/milestones"},
	{"GET", "/repos/owner/repo/milestones/number"},
	{"POST", "/repos/owner/repo/milestones"},
	{"PATCH", "/repos/owner/repo/milestones/number"},
	{"DELETE", "/repos/owner/repo/milestones/number"},
	// misc
	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/C"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},
	// orgs
	{"GET", "/user/orgs"},
	{"GET", "/users/username/orgs"},
	{"GET", "/orgs/org"},
	{"PATCH", "/orgs/org"},
	// org members
	{"GET", "/orgs/org/members"},
	{"GET", "/orgs/org/members/username"},
	{"DELETE", "/orgs/org/members/username"},
	{"GET", "/orgs/org/public_members"},
	{"GET", "/orgs/org/public_members/username"},
	{"PUT", "/orgs/org/public_members/username"},
	{"DELETE", "/orgs/org/public_members/username"},
	{"GET", "/orgs/org/memberships/username"},
	{"PUT", "/orgs/org/memberships/username"},
	{"DELETE", "/orgs/org/memberships/username"},
	{"GET", "/user/memberships/orgs"},
	{"GET", "/user/memberships/orgs/org"},
	{"PATCH", "/user/memberships/orgs/org"},
	// org teams
	{"GET", "/orgs/org/teams"},
	{"GET", "/teams/id"},
	{"POST", "/orgs/org/teams"},
	{"PATCH", "/teams/id"},
	{"DELETE", "/teams/id"},
	{"GET", "/teams/id/members"},
	{"GET", "/teams/id/members/username"},
	{"PUT", "/teams/id/members/username"},
	{"DELETE", "/teams/id/members/username"},
	{"GET", "/teams/id/memberships/username"},
	{"PUT", "/teams/id/memberships/username"},
	{"DELETE", "/teams/id/memberships/username"},
	{"GET", "/teams/id/repos"},
	{"GET", "/teams/id/repos/org/repo"},
	{"PUT", "/teams/id/repos/org/repo"},
	{"DELETE", "/teams/id/repos/org/repo"},
	{"GET", "/user/teams"},
	// org webhooks
	{"GET", "/orgs/org/hooks"},
	{"GET", "/orgs/org/hooks/id"},
	{"POST", "/orgs/org/hooks"},
	{"PATCH", "/orgs/org/hooks/id"},
	{"POST", "/orgs/org/hooks/id/pings"},
	{"DELETE", "/orgs/org/hooks/id"},
	// pulls
	{"GET", "/repos/owner/repo/pulls"},
	{"GET", "/repos/owner/repo/pulls/number"},
	{"POST", "/repos/owner/repo/pulls"},
	{"PATCH", "/repos/owner/repo/pulls/number"},
	{"GET", "/repos/owner/repo/pulls/number/commits"},
	{"GET", "/repos/owner/repo/pulls/number/files"},
	{"GET", "/repos/owner/repo/pulls/number/merge"},
	{"PUT", "/repos/owner/repo/pulls/number/merge"},
	// pull review comments
	{"GET", "/repos/owner/repo/pulls/number/comments"},
	{"GET", "/repos/owner/repo/pulls/comments"},
	{"GET", "/repos/owner/repo/pulls/comments/number"},
	{"POST", "/repos/owner/repo/pulls/number/comments"},
	{"PATCH", "/repos/owner/repo/pulls/comments/number"},
	{"DELETE", "/repos/owner/repo/pulls/comments/number"},
	// repos
	{"GET", "/user/repos"},
	{"GET", "/users/username/repos"},
	{"GET", "/orgs/org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/org/repos"},
	{"GET", "/repos/owner/repo"},
	{"PATCH", "/repos/owner/repo"},
	{"GET", "/repos/owner/repo/contributors"},
	{"GET", "/repos/owner/repo/languages"},
	{"GET", "/repos/owner/repo/teams"},
	{"GET", "/repos/owner/repo/tags"},
	{"GET", "/repos/owner/repo/branches"},
	{"GET", "/repos/owner/repo/branches/branch"},
	{"DELETE", "/repos/owner/repo"},
	// repo collaborators
	{"GET", "/repos/owner/repo/collaborators"},
	{"GET", "/repos/owner/repo/collaborators/username"},
	{"PUT", "/repos/owner/repo/collaborators/username"},
	{"DELETE", "/repos/owner/repo/collaborators/username"},
	// repo comments
	{"GET", "/repos/owner/repo/comments"},
	{"GET", "/repos/owner/repo/commits/ref/comments"},
	{"POST", "/repos/owner/repo/commits/ref/comments"},
	{"GET", "/repos/owner/repo/comments/id"},
	{"PATCH", "/repos/owner/repo/comments/id"},
	{"DELETE", "/repos/owner/repo/comments/id"},
	// repo commits
	{"GET", "/repos/owner/repo/commits"},
	{"GET", "/repos/owner/repo/commits/ref"},
	{"GET", "/repos/owner/repo/compare/base...{head}"},
	// repo contents
	{"GET", "/repos/owner/repo/readme"},
	{"GET", "/repos/owner/repo/contents/path"},
	{"PUT", "/repos/owner/repo/contents/path"},
	{"PATCH", "/repos/owner/repo/contents/path"},
	{"DELETE", "/repos/owner/repo/contents/path"},
	{"GET", "/repos/owner/repo/archive_format/ref"},
	// repo deploy keys
	{"GET", "/repos/owner/repo/keys"},
	{"GET", "/repos/owner/repo/keys/id"},
	{"POST", "/repos/owner/repo/keys"},
	{"DELETE", "/repos/owner/repo/keys/id"},
	// repo deployments
	{"GET", "/repos/owner/repo/deployments"},
	{"POST", "/repos/owner/repo/deployments"},
	{"GET", "/repos/owner/repo/deployments/id/statuses"},
	{"POST", "/repos/owner/repo/deployments/id/statuses"},
	// repo downloads
	{"GET", "/repos/owner/repo/downloads"},
	{"GET", "/repos/owner/repo/downloads/id"},
	{"DELETE", "/repos/owner/repo/downloads/id"},
	// repo forks
	{"GET", "/repos/owner/repo/forks"},
	{"POST", "/repos/owner/repo/forks"},
	// repo webhooks
	{"GET", "/repos/owner/repo/hooks"},
	{"GET", "/repos/owner/repo/hooks/id"},
	{"POST", "/repos/owner/repo/hooks"},
	{"PATCH", "/repos/owner/repo/hooks/id"},
	{"POST", "/repos/owner/repo/hooks/id/tests"},
	{"POST", "/repos/owner/repo/hooks/id/pings"},
	{"DELETE", "/repos/owner/repo/hooks/id"},
	// repo merging
	{"POST", "/repos/owner/repo/merges"},
	// repo pages
	{"GET", "/repos/owner/repo/pages"},
	{"GET", "/repos/owner/repo/pages/builds"},
	{"GET", "/repos/owner/repo/pages/builds/latest"},
	// repo releases
	{"GET", "/repos/owner/repo/releases"},
	{"GET", "/repos/owner/repo/releases/id"},
	{"POST", "/repos/owner/repo/releases"},
	{"PATCH", "/repos/owner/repo/releases/id"},
	{"DELETE", "/repos/owner/repo/releases/id"},
	{"GET", "/repos/owner/repo/releases/id/assets"},
	{"GET", "/repos/owner/repo/releases/assets/id"},
	{"PATCH", "/repos/owner/repo/releases/assets/id"},
	{"DELETE", "/repos/owner/repo/releases/assets/id"},
	// repo satistics
	{"GET", "/repos/owner/repo/stats/contributors"},
	{"GET", "/repos/owner/repo/stats/commit_activity"},
	{"GET", "/repos/owner/repo/stats/code_frequency"},
	{"GET", "/repos/owner/repo/stats/participation"},
	{"GET", "/repos/owner/repo/stats/punch_card"},
	// repo statuses
	{"POST", "/repos/owner/repo/statuses/sha"},
	{"GET", "/repos/owner/repo/commits/ref/statuses"},
	{"GET", "/repos/owner/repo/commits/ref/status"},
	// search
	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	// users
	{"GET", "/users/username"},
	{"GET", "/user"},
	{"PATCH", "/user"},
	{"GET", "/users"},
	// user emails
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	// user followers
	{"GET", "/users/username/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/username/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/username"},
	{"GET", "/users/username/following/target_user"},
	{"PUT", "/user/following/username"},
	{"DELETE", "/user/following/username"},
	// user public keys
	{"GET", "/users/username/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/id"},
	{"POST", "/user/keys"},
	{"DELETE", "/user/keys/id"},
	// user administration (enterprise)
	{"PUT", "/users/username/site_admin"},
	{"DELETE", "/users/username/site_admin"},
	{"PUT", "/users/username/suspendeds"},
	{"DELETE", "/users/username/suspended"},
	// enterprise
	{"GET", "/enterprise/stats/type"},
	{"GET", "/enterprise/settings/license"},
	{"POST", "/staff/indexing_jobs"},
	// enterprise management console
	{"POST", "/setup/api/start"},
	{"POST", "/setup/api/upgrade"},
	{"GET", "/setup/api/configcheck"},
	{"POST", "/setup/api/configure"},
	{"GET", "/setup/api/settings"},
	{"PUT", "/setup/api/settings"},
	{"GET", "/setup/api/maintenance"},
	{"POST", "/setup/api/maintenance"},
	{"GET", "/setup/api/settings/authorized-keys"},
	{"POST", "/setup/api/settings/authorized-keys"},
	{"DELETE", "/setup/api/settings/authorized-keys"},
	// enterprise LDAP
	{"PATCH", "/admin/ldap/user/username/mapping"},
	{"POST", "/admin/ldap/user/username/sync"},
	{"PATCH", "/admin/ldap/teams/team_id/mapping"},
	{"POST", "/admin/ldap/teams/team_id/sync"},
}
