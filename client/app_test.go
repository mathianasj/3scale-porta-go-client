package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/3scale/3scale-porta-go-client/fake"
)

func TestCreateApp(t *testing.T) {
	const (
		credential = "123"
		accountID  = "321"
		planID     = "abc"
		name       = "test"
	)

	inputs := []struct {
		name       string
		returnErr  bool
		expectCode int
		expectErr  string
	}{
		{
			name:      "Test app creation fail",
			returnErr: true,
			expectErr: `error calling 3scale system - reason: { "error": "Your access token does not have the correct permissions" } - code: 403`,
		},
		{
			name: "Test app creation success",
		},
	}

	for _, input := range inputs {
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			if req.Method != http.MethodPost {
				t.Fatalf("wrong helper called for create app api")
			}

			if req.URL.Path != "/admin/api/accounts/321/applications.json" {
				t.Fatal("wrong url generated by CreateApp function")
			}

			if input.returnErr {
				return fake.CreateAppError()
			}
			return fake.CreateAppSuccess(input.name)
		})

		c := NewThreeScale(NewTestAdminPortal(t), credential, httpClient)

		t.Run(input.name, func(t *testing.T) {
			a, b := c.CreateApp(accountID, planID, name, input.name)
			if input.returnErr {
				e := b.(ApiErr)
				if e.Code() != http.StatusForbidden {
					t.Fatal("unexpected code returned in error")
				}
				if b.Error() != input.expectErr {
					t.Fatalf("unexpected error message. Error received: %s", b.Error())
				}
				return
			}

			if a.Error != "" {
				t.Fatal("expected error to be empty")
			}
			if a.Description != input.name {
				t.Fatal("xml has not decoded correctly")
			}
		})
	}
}

func TestListApp(t *testing.T) {
	const (
		accessToken = "someAccessToken"
		accountID   = int64(321)
	)

	inputs := []struct {
		Name             string
		ExpectErr        bool
		ResponseCode     int
		ResponseBodyFile string
		ExpectedErrorMsg string
	}{
		{
			Name:             "ListAppOK",
			ExpectErr:        false,
			ResponseCode:     200,
			ResponseBodyFile: "app_list_response_fixture.json",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "ListAppErr",
			ExpectErr:        true,
			ResponseCode:     400,
			ResponseBodyFile: "error_response_fixture.json",
			ExpectedErrorMsg: "Test Error",
		},
	}

	for _, input := range inputs {
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			if req.Method != http.MethodGet {
				t.Fatalf("wrong helper called")
			}

			if req.URL.Path != fmt.Sprintf(appList, accountID) {
				t.Fatalf("wrong url generated")
			}

			bodyReader := bytes.NewReader(helperLoadBytes(t, input.ResponseBodyFile))
			return &http.Response{
				StatusCode: input.ResponseCode,
				Body:       ioutil.NopCloser(bodyReader),
				Header:     make(http.Header),
			}
		})

		c := NewThreeScale(NewTestAdminPortal(t), accessToken, httpClient)

		t.Run(input.Name, func(subTest *testing.T) {
			appList, err := c.ListApplications(accountID)
			if input.ExpectErr {
				if err == nil {
					subTest.Fatalf("client operation did not return error")
				}

				apiError, ok := err.(ApiErr)
				if !ok {
					subTest.Fatalf("expected ApiErr error type")
				}

				if !strings.Contains(apiError.Error(), input.ExpectedErrorMsg) {
					subTest.Fatalf("Expected [%s]: got [%s] ", input.ExpectedErrorMsg, apiError.Error())
				}

			} else {
				if err != nil {
					subTest.Fatal(err)
				}
				if appList == nil {
					subTest.Fatalf("appList not parsed")
				}
				if len(appList.Applications) == 0 {
					subTest.Fatalf("appList empty")
				}
				if appList.Applications[0].Application.ID != 146 {
					subTest.Fatalf("appList not parsed")
				}
			}
		})
	}
}
