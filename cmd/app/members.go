package app

import (
	"encoding/json"
	"net/http"

	"github.com/xanzy/go-gitlab"
)

type ProjectMembersResponse struct {
	SuccessResponse
	ProjectMembers []*gitlab.ProjectMember
}

type ProjectMemberLister interface {
	ListAllProjectMembers(pid interface{}, opt *gitlab.ListProjectMembersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectMember, *gitlab.Response, error)
}

type projectMemberService struct {
	data
	client ProjectMemberLister
}

/* projectMembersHandler returns all members of the current Gitlab project */
func (a projectMemberService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectMemberOptions := &gitlab.ListProjectMembersOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	var allProjectMembers []*gitlab.ProjectMember
	for {
		projectMembers, res, err := a.client.ListAllProjectMembers(a.projectInfo.ProjectId, projectMemberOptions)
		if err != nil {
			handleError(w, err, "Could not retrieve project members", http.StatusInternalServerError)
			return
		}

		if res.StatusCode >= 300 {
			handleError(w, GenericError{r.URL.Path}, "Could not retrieve project members", res.StatusCode)
			return
		}

		allProjectMembers = append(allProjectMembers, projectMembers...)

		if res.NextPage == 0 {
			break
		}

		projectMemberOptions.Page = res.NextPage
	}

	w.WriteHeader(http.StatusOK)

	response := ProjectMembersResponse{
		SuccessResponse: SuccessResponse{Message: "Project members retrieved"},
		ProjectMembers:  allProjectMembers,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, "Could not encode response", http.StatusInternalServerError)
	}
}
