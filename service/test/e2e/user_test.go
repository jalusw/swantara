//go:build e2e

package e2e

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
)

var (
	publicUserFields  = []any{"id", "username", "first_name", "last_name", "avatar", "bio"}
	privateUserFields = []any{"email", "phone", "birthday", "address", "city", "postal_code", "active", "private"}
)

func TestCreateUserForbiddenForBasicUserE2E(t *testing.T) {
	user := registerAndLogin(t)

	newExpect(t).POST("/api/v1/users").
		WithHeader("Authorization", "Bearer "+user.tokens.AccessToken).
		WithJSON(map[string]any{
			"username":   gofakeit.Username(),
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
			"email":      gofakeit.Email(),
			"password":   "test-password",
		}).
		Expect().
		Status(http.StatusForbidden)
}

func TestCreateUserRequiresTokenE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/users").
		WithJSON(map[string]any{
			"username":   gofakeit.Username(),
			"first_name": gofakeit.FirstName(),
			"email":      gofakeit.Email(),
			"password":   "test-password",
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestPublicUserReadsE2E(t *testing.T) {
	user := registerAndLogin(t)

	e := newExpect(t)

	usersArray := e.GET("/api/v1/users").
		WithHeader("Authorization", "Bearer "+user.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("users").
		Array()

	usersArray.Length().Gt(0)
	assertPublicUserProjection(usersArray)

	userObject := e.GET("/api/v1/users/"+strconv.FormatUint(user.ID, 10)).
		WithHeader("Authorization", "Bearer "+user.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("user").
		Object()

	assertPublicUserProjectionObject(userObject)
}

func assertPublicUserProjection(usersArray *httpexpect.Array) {
	for i := 0; i < int(usersArray.Length().Raw()); i++ {
		assertPublicUserProjectionObject(usersArray.Element(i).Object())
	}
}

func assertPublicUserProjectionObject(userObject *httpexpect.Object) {
	userObject.Keys().Contains(publicUserFields...)
	userObject.Keys().NotContains(privateUserFields...)
}
