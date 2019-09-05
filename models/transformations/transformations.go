package transformations

import (
	"strings"

	"gitlab.com/p-invent/mosoly-ledger-bridge/models/dbmodels"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mosolyapi"
)

// TransformUser transforms mosoly api user to database user.
func TransformUser(user *mosolyapi.User) *dbmodels.User {
	mentorees := make([]dbmodels.Mentoree, 0)
	mentors := make([]dbmodels.Mentor, 0)

	for _, mentoree := range user.Mentorees {
		mentorees = append(mentorees, dbmodels.Mentoree{
			ID:      mentoree.UserID,
			Account: strings.ToLower(mentoree.Account),
		})
	}

	for _, mentor := range user.Mentors {
		mentors = append(mentors, dbmodels.Mentor{
			ID:      mentor.UserID,
			Account: strings.ToLower(mentor.Account),
			Users:   make([]string, 0),
		})
	}

	return &dbmodels.User{
		ID:            user.ID,
		UpdatedAt:     user.UpdatedAt,
		InviteURLHash: user.InviteURLHash,
		Account:       strings.ToLower(user.Account),
		Validated:     user.Validated,
		Mentorees:     mentorees,
		Mentors:       mentors,
	}
}

// TransformProject transforms mosoly api project to database project.
func TransformProject(project *mosolyapi.Project) (*dbmodels.Project, error) {
	return &dbmodels.Project{
		ID:        project.ID,
		UpdatedAt: project.UpdatedAt,
		Name:      project.Name,
	}, nil
}
