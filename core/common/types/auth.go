/**
 * @Author:         yi
 * @Description:    auth
 * @Version:        1.0.0
 * @Date:           2025/4/22 10:01
 */

package types

import (
	"skeyevss/core/repositories/models/admins"
	"skeyevss/core/repositories/models/departments"
	"skeyevss/core/repositories/models/roles"
)

type AuthRes struct {
	Admins      []*admins.Item      `json:"admins"`
	Departments []*departments.Item `json:"departments"`
	Roles       []*roles.Item       `json:"roles"`
}
