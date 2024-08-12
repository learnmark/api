package validator

import (
	"github.com/learnmark/learnmark/internal/model"
)

func IsOrgAdmin(orgUser []model.OrgUser, user model.User) bool {
	if user.IsSuperAdmin {
		return true
	}
	for _, m := range orgUser {
		if m.ForOrg.CreatedBy == user.Id {
			return true
		}
		if m.UserId == user.Id && m.IsAdmin == true {
			return true
		}
	}
	return false
}

func IsOrgMember(orgUser []model.OrgUser, user model.User) bool {
	if user.IsSuperAdmin {
		return true
	}
	for _, m := range orgUser {
		if m.ForOrg.CreatedBy == user.Id {
			return true
		}
		if m.UserId == user.Id {
			return true
		}
	}
	return false
}
