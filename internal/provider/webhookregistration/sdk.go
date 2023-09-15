package webhookregistration

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"text/template"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *webhookRegistrationResourceModel) Refresh(response *string) {
	if response != nil {
		m.Register.Response = types.StringValue(*response)
	} else {
		m.Register.Response = types.StringValue(m.Register.Response.ValueString())
	}
}

func (m *webhookRegistrationResourceModel) ToRegisterRequest() (*http.Request, error) {
	method := getMethod(m.Register.Request.Method.ValueString())
	requestURL := m.Register.Request.URL.ValueString()

	var bodyReader io.Reader = nil
	if !m.Register.Request.Body.IsUnknown() && !m.Register.Request.Body.IsNull() {
		body := m.Register.Request.Body.ValueString()
		bodyReader = bytes.NewReader([]byte(body))
	}

	request, err := http.NewRequest(method, requestURL, bodyReader)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if !m.Register.Request.Headers.IsUnknown() && !m.Register.Request.Headers.IsNull() {
		if err := json.Unmarshal([]byte(m.Register.Request.Headers.ValueString()), &headers); err != nil {
			return nil, err
		}
	}
	for headerKey, headerValue := range headers {
		request.Header.Set(headerKey, headerValue)
	}

	return request, nil
}

func (m *webhookRegistrationResourceModel) ToUnregisterRequest() (*http.Request, error) {
	if m.Unregister == nil {
		return nil, nil
	}

	method := getMethod(m.Unregister.Request.Method.ValueString())
	requestURL, err := m.ExecuteRegistrationResponseTemplate(m.Unregister.Request.URL.ValueString())
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader = nil
	if !m.Unregister.Request.Body.IsUnknown() && !m.Unregister.Request.Body.IsNull() {
		body, err := m.ExecuteRegistrationResponseTemplate(m.Unregister.Request.Body.ValueString())
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader([]byte(*body))
	}

	request, err := http.NewRequest(method, *requestURL, bodyReader)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if !m.Unregister.Request.Headers.IsUnknown() && !m.Unregister.Request.Headers.IsNull() {
		if err := json.Unmarshal([]byte(m.Unregister.Request.Headers.ValueString()), &headers); err != nil {
			return nil, err
		}
	}
	for headerKey, headerValue := range headers {
		request.Header.Set(headerKey, headerValue)
	}

	return request, nil
}

func (m *webhookRegistrationResourceModel) DoRequest(request *http.Request, shouldClose bool) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if shouldClose {
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			return nil, err
		}
		return nil, errors.New(string(bytes))
	}

	return resp, nil
}

func (m *webhookRegistrationResourceModel) ParseRegisterResponse(response *http.Response) (*string, error) {
	defer response.Body.Close()

	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var responseBodyJSON map[string]any
	_ = json.Unmarshal(responseBodyBytes, &responseBodyJSON)

	headers := map[string][]string{}
	for k, v := range response.Header {
		headers[k] = v
	}

	responseJSON := map[string]any{
		"body":   responseBodyJSON,
		"header": headers,
	}
	responseBytes, err := json.Marshal(responseJSON)
	if err != nil {
		return nil, err
	}
	responseString := string(responseBytes)

	return &responseString, nil
}

func (m *webhookRegistrationResourceModel) MarshallRegistrationResponse() (*map[string]any, error) {
	responseJSON := map[string]any{}
	if err := json.Unmarshal([]byte(m.Register.Response.ValueString()), &responseJSON); err != nil {
		return nil, err
	}
	return &responseJSON, nil
}

func (m *webhookRegistrationResourceModel) CreateRegistrationResponseTemplate(text string) *template.Template {
	return template.Must(template.New("template").Parse(text))
}

func (m *webhookRegistrationResourceModel) ExecuteRegistrationResponseTemplate(text string) (*string, error) {
	responseJSON, err := m.MarshallRegistrationResponse()
	if err != nil {
		return nil, err
	}

	var data bytes.Buffer
	registrationReponseTemplate := m.CreateRegistrationResponseTemplate(text)

	var templateData map[string]any = map[string]any{
		"register": map[string]any{
			"response": responseJSON,
		},
	}
	if err := registrationReponseTemplate.Execute(&data, templateData); err != nil {
		return nil, err
	}

	str := data.String()
	return &str, nil
}

func getMethod(method string) string {
	switch method {
	case "GET":
		return http.MethodGet
	case "POST":
		return http.MethodPost
	case "PUT":
		return http.MethodPut
	case "PATCH":
		return http.MethodPatch
	case "DELETE":
		return http.MethodDelete
	default:
		return http.MethodPost
	}
}
