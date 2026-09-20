package seeders

import (
	"fmt"
	"strings"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type iamSeedData struct {
	permissions []*iam.Permission
}

type iamSeedFile struct {
	Permissions    []*iam.Permission  `json:"permissions"`
	OrgPermissions []iamOrgPermission `json:"org_permissions"`
}

type iamOrgPermission struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

func loadIamSeed() (*iamSeedData, error) {
	file, err := loadSeedData[iamSeedFile]("data/iam_seed.json")
	if err != nil {
		return nil, err
	}

	data := &iamSeedData{}
	data.permissions = append(data.permissions, file.Permissions...)
	for _, def := range file.OrgPermissions {
		for _, action := range def.Actions {
			data.permissions = append(data.permissions, &iam.Permission{
				Name:        fmt.Sprintf("%s %s", capitalize(action), def.Resource),
				Code:        def.Resource + "." + action,
				Description: helper.Ptr("Permission to " + action + " " + def.Resource),
				Resource:    def.Resource,
				Action:      action,
			})
		}
	}

	return data, nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
