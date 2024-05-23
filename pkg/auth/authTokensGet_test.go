/*
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/galasa-dev/cli/pkg/api"
	"github.com/galasa-dev/cli/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func NewPropertiesNamespaceServletMock(t *testing.T, state string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockPropertiesNamespaceServlet(t, w, r, state)
	}))
	return server
}

func mockPropertiesNamespaceServlet(t *testing.T, writer http.ResponseWriter, request *http.Request, state string) {
	writer.Header().Set("Content-Type", "application/json")
	var statusCode int
	var body string
	statusCode = 200
	if state == "populated" {
		body = `{
    "tokens":[
        {
            "token_id":"098234980123-1283182389",
            "creation_time":"2023-12-03T18:25:43.511Z",
            "owner": {
                "login_id":"mcobbett"
            },
            "description":"So I can access ecosystem1 from my laptop."
        },
        {
            "token_id":"8218971d287s1-dhj32er2323",
            "creation_time":"2024-03-03T09:36:50.511Z",
            "owner": {
                "login_id":"mcobbett"
            },
            "description":"Automated build of example repo can change CPS properties"
        },
        {
            "token_id":"87a6sd87ahq2-2y8hqwdjj273",
            "creation_time":"2023-08-04T23:00:23.511Z",
            "owner": {
                "login_id":"savvas"
            },
            "description":"CLI access from vscode"
        }
	]
}`
	} else if state == "empty" {
		body = `{
    "tokens":[]
}`
	} else {
		statusCode = 500
		body = `{"error_code": 5000,"error_message": "GAL5000E: Error occured when trying to access the endpoint. Report the problem to your Galasa Ecosystem owner."}`
	}
	writer.WriteHeader(statusCode)
	writer.Write([]byte(body))
}

func TestMultipleNamespacesPathReturnsOk(t *testing.T) {
	//Given...
	serverState := "populated"
	server := NewPropertiesNamespaceServletMock(t, serverState)
	apiClient := api.InitialiseAPI(server.URL)
	defer server.Close()

	console := utils.NewMockConsole()
	expectedOutput := `tokenid                   created(YYYY-MM-DD) user     description
098234980123-1283182389   2023-12-03          mcobbett So I can access ecosystem1 from my laptop.
8218971d287s1-dhj32er2323 2024-03-03          mcobbett Automated build of example repo can change CPS properties
87a6sd87ahq2-2y8hqwdjj273 2023-08-04          savvas   CLI access from vscode

Total:3
`

	//When
	err := GetTokens(apiClient, console)

	//Then
	assert.Nil(t, err)
	assert.Equal(t, expectedOutput, console.ReadText())
}

func TestNoTokensPathReturnsOk(t *testing.T) {
	//Given...
	serverState := "empty"
	server := NewPropertiesNamespaceServletMock(t, serverState)
	apiClient := api.InitialiseAPI(server.URL)
	defer server.Close()

	console := utils.NewMockConsole()
	expectedOutput := "Total:0\n"

	//When
	err := GetTokens(apiClient, console)

	//Then
	assert.Nil(t, err)
	assert.Equal(t, expectedOutput, console.ReadText())
}

func TestInvalidPathReturnsError(t *testing.T) {
	//Given...
	serverState := ""
	server := NewPropertiesNamespaceServletMock(t, serverState)
	apiClient := api.InitialiseAPI(server.URL)
	defer server.Close()

	console := utils.NewMockConsole()

	//When
	err := GetTokens(apiClient, console)

	//Then
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "GAL1145E")
}
