package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pdkovacs/igo-repo/api"
	"github.com/pdkovacs/igo-repo/config"
	"github.com/pdkovacs/igo-repo/domain"
	"github.com/pdkovacs/igo-repo/security/authr"
	"github.com/pdkovacs/igo-repo/test/api/testdata"
)

type apiTestSession struct {
	apiTestClient
	cjar http.CookieJar
}

func (client *apiTestClient) login(credentials *requestCredentials) (*apiTestSession, error) {
	if credentials == nil {
		calculatedCredentials, credError := makeRequestCredentials(config.BasicAuthentication, testdata.DefaultCredentials.Username, testdata.DefaultCredentials.Password)
		if credError != nil {
			return &apiTestSession{}, fmt.Errorf("failed to create default request credentials: %w", credError)
		}
		credentials = &calculatedCredentials
	}
	cjar := client.MustCreateCookieJar()

	resp, postError := client.post(&testRequest{
		path:        "/login",
		credentials: credentials,
		jar:         cjar,
		json:        true,
		body:        credentials,
	})
	if postError != nil {
		return &apiTestSession{}, fmt.Errorf("failed to login: %w", postError)
	}
	if resp.statusCode != 200 {
		return &apiTestSession{}, fmt.Errorf(
			"failed to login with status code %d: %w",
			resp.statusCode,
			errors.New("authentication error"),
		)
	}

	return &apiTestSession{
		apiTestClient: apiTestClient{
			serverPort: client.serverPort,
		},
		cjar: cjar,
	}, nil
}

func (client *apiTestClient) mustLogin(credentials *requestCredentials) *apiTestSession {
	session, err := client.login(credentials)
	if err != nil {
		panic(err)
	}
	return session
}

func (client *apiTestClient) mustLoginSetAllPerms() *apiTestSession {
	session := client.mustLogin(nil)
	session.mustSetAuthorization(authr.GetPermissionsForGroup(authr.ICON_EDITOR))
	return session
}

func (session *apiTestSession) get(request *testRequest) (testResponse, error) {
	request.jar = session.cjar
	return session.sendRequest("GET", request)
}

func (session *apiTestSession) put(request *testRequest) (testResponse, error) {
	request.jar = session.cjar
	return session.sendRequest("PUT", request)
}

func (session *apiTestSession) setAuthorization(requestedAuthorization []authr.PermissionID) (testResponse, error) {
	var err error
	var resp testResponse
	credentials := session.makeRequestCredentials(testdata.DefaultCredentials)
	if err != nil {
		panic(err)
	}
	resp, err = session.sendRequest("PUT", &testRequest{
		path:        authenticationBackdoorPath,
		credentials: &credentials,
		jar:         session.cjar,
		json:        true,
		body:        requestedAuthorization,
	})
	return resp, err
}

func (session *apiTestSession) mustSetAllPermsExcept(toExclude []authr.PermissionID) {
	all := authr.GetPermissionsForGroup(authr.ICON_EDITOR)
	filtered := []authr.PermissionID{}
	for _, oneOfAll := range all {
		include := true
		for _, oneOfExcept := range toExclude {
			if oneOfExcept == oneOfAll {
				include = false
				break
			}
		}
		if include {
			filtered = append(filtered, oneOfAll)
		}
	}
	session.mustSetAuthorization(filtered)
}

func (session *apiTestSession) mustSetAuthorization(requestedPermissions []authr.PermissionID) {
	resp, err := session.setAuthorization(requestedPermissions)
	if err != nil {
		panic(err)
	}
	if resp.statusCode != 200 {
		panic(fmt.Sprintf("Unexpected status code: %d", resp.statusCode))
	}
}

func (session *apiTestSession) mustAddTestData(testData []domain.Icon) {
	var err error
	var statusCode int
	for _, testIcon := range testData {
		statusCode, _, err = session.createIcon(testIcon.Name, testIcon.Iconfiles[0].Content)
		if err != nil {
			panic(err)
		}
		if statusCode != 201 {
			panic(fmt.Sprintf("Unexpected status code %d, expected %d", statusCode, 201))
		}
		for i := 1; i < len(testIcon.Iconfiles); i++ {
			statusCode, _, err = session.addIconfile(testIcon.Name, testIcon.Iconfiles[i])
			if err != nil {
				panic(fmt.Errorf("failed to add iconfile to %s with status code %d: %w", testIcon.Name, statusCode, err))
			}
		}
	}
}

func (session *apiTestSession) describeAllIcons() ([]api.ResponseIcon, error) {
	resp, err := session.get(&testRequest{
		path:          "/icon",
		jar:           session.cjar,
		respBodyProto: &[]api.ResponseIcon{},
	})
	if err != nil {
		return []api.ResponseIcon{}, fmt.Errorf("GET /icon failed: %w", err)
	}
	if resp.statusCode != 200 {
		return []api.ResponseIcon{}, fmt.Errorf("%w: got %d", errUnexpecteHTTPStatus, resp.statusCode)
	}
	icons, ok := resp.body.(*[]api.ResponseIcon)
	if !ok {
		return []api.ResponseIcon{}, fmt.Errorf("failed to cast %T as []api.ResponseIcon", resp.body)
	}
	return *icons, err
}

func (session *apiTestSession) mustDescribeAllIcons() []api.ResponseIcon {
	respIcons, err := session.describeAllIcons()
	if err != nil {
		panic(err)
	}
	return respIcons
}

func (session *apiTestSession) describeIcon(iconName string) (int, api.ResponseIcon, error) {
	resp, err := session.get(&testRequest{
		path:          fmt.Sprintf("/icon/%s", iconName),
		jar:           session.cjar,
		respBodyProto: &api.ResponseIcon{},
	})
	if err != nil {
		return resp.statusCode, api.ResponseIcon{}, fmt.Errorf("GET /icon/%s failed: %w", iconName, err)
	}
	icon, ok := resp.body.(*api.ResponseIcon)
	if !ok {
		return resp.statusCode, api.ResponseIcon{}, fmt.Errorf("failed to cast %T as api.ResponseIcon", resp.body)
	}
	return resp.statusCode, *icon, err
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func (session *apiTestSession) createIcon(iconName string, initialIconfile []byte) (int, api.ResponseIcon, error) {
	var err error
	var resp testResponse

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var fw io.Writer
	if fw, err = w.CreateFormField("iconName"); err != nil {
		panic(err)
	}
	if _, err = io.Copy(fw, strings.NewReader(iconName)); err != nil {
		panic(err)
	}

	if fw, err = w.CreateFormFile("iconfile", iconName); err != nil {
		panic(err)
	}
	if _, err = io.Copy(fw, bytes.NewReader(initialIconfile)); err != nil {
		panic(err)
	}
	w.Close()

	headers := map[string]string{
		"Content-Type": w.FormDataContentType(),
	}

	resp, err = session.sendRequest("POST", &testRequest{
		path:          "/icon",
		jar:           session.cjar,
		headers:       headers,
		body:          b.Bytes(),
		respBodyProto: &api.ResponseIcon{},
	})
	if err != nil {
		return resp.statusCode, api.ResponseIcon{}, err
	}

	statusCode := resp.statusCode

	if respIconfile, ok := resp.body.(*api.ResponseIcon); ok {
		return statusCode, *respIconfile, err
	}

	return statusCode, api.ResponseIcon{}, fmt.Errorf("failed to cast %T to api.ResponseIcon", resp.body)
}

func (session *apiTestSession) deleteIcon(iconName string) (int, error) {
	resp, deleteError := session.sendRequest(
		"DELETE",
		&testRequest{
			path: fmt.Sprintf("/icon/%s", iconName),
			jar:  session.cjar,
		},
	)
	return resp.statusCode, deleteError
}

func (s *apiTestSession) GetIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor) (domain.Iconfile, error) {
	iconfile := domain.Iconfile{}
	resp, reqErr := s.get(&testRequest{
		path:          getFilePath(iconName, iconfileDescriptor),
		respBodyProto: &iconfile,
	})
	if reqErr != nil {
		return iconfile, fmt.Errorf("failed to retrieve iconfile %v of %s: %w", iconfileDescriptor, iconName, reqErr)
	}

	if respIconfile, ok := resp.body.(*domain.Iconfile); ok {
		iconfile.Content = respIconfile.Content
		return iconfile, nil
	}

	return iconfile, fmt.Errorf("failed to cast the reply %T to []byte while retrieving iconfile %v of %s", resp.body, iconfileDescriptor, iconName)
}

func (session *apiTestSession) addIconfile(iconName string, iconfile domain.Iconfile) (int, api.IconPath, error) {
	var err error
	var resp testResponse

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var fw io.Writer
	if fw, err = w.CreateFormField("iconName"); err != nil {
		panic(err)
	}
	if _, err = io.Copy(fw, strings.NewReader(iconName)); err != nil {
		panic(err)
	}

	if fw, err = w.CreateFormFile("iconfile", iconName); err != nil {
		panic(err)
	}
	if _, err = io.Copy(fw, bytes.NewReader(iconfile.Content)); err != nil {
		panic(err)
	}
	w.Close()

	headers := map[string]string{
		"Content-Type": w.FormDataContentType(),
	}

	resp, err = session.sendRequest("POST", &testRequest{
		path:          fmt.Sprintf("/icon/%s", iconName),
		jar:           session.cjar,
		headers:       headers,
		body:          b.Bytes(),
		respBodyProto: &api.IconPath{},
	})
	if err != nil {
		return resp.statusCode, api.IconPath{}, err
	}

	if respIconfile, ok := resp.body.(*api.IconPath); ok {
		return resp.statusCode, *respIconfile, nil
	}

	return resp.statusCode, api.IconPath{}, fmt.Errorf("failed to cast %T to domain.Icon", resp.body)
}

func (session *apiTestSession) deleteIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor) (int, error) {
	resp, err := session.sendRequest("DELETE", &testRequest{
		path: api.CreateIconPath("/icon", iconName, iconfileDescriptor).Path,
		jar:  session.cjar,
	})

	if err != nil {
		return 0, err
	}

	return resp.statusCode, err
}

func (session *apiTestSession) addTag(iconName string, tag string) (int, error) {
	requestData := api.AddServiceRequestData{Tag: tag}
	resp, err := session.sendRequest("POST", &testRequest{
		path: fmt.Sprintf("/icon/%s/tag", iconName),
		jar:  session.cjar,
		json: true,
		body: requestData,
	})
	if err != nil {
		return 0, err
	}

	return resp.statusCode, err
}

func (session *apiTestSession) removeTag(iconName string, tag string) (int, error) {
	resp, err := session.sendRequest("DELETE", &testRequest{
		path: fmt.Sprintf("/icon/%s/tag/%s", iconName, tag),
		jar:  session.cjar,
	})
	if err != nil {
		return 0, err
	}

	return resp.statusCode, err
}
