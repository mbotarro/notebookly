package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

type CloneCmd struct {
	ID    int64      `help:"ID of the notebook to clone."`
	Name  string     `help:"New name of the cloned notebook."`
	Var   []string   `help:"Template vars to create, format: name,tag,default." sep:";"`
	Start *time.Time `help:"Start time"`
	End   *time.Time `help:"End time"`
}

// These structs are copied since we need to add template vars.
// NotebookCreateDataAttributes The data attributes of a notebook.
type NotebookCreateDataAttributes struct {
	// List of cells to display in the notebook.
	Cells []datadogV1.NotebookCellCreateRequest `json:"cells"`
	// Metadata associated with the notebook.
	Metadata *datadogV1.NotebookMetadata `json:"metadata,omitempty"`
	// The name of the notebook.
	Name string `json:"name"`
	// Publication status of the notebook. For now, always "published".
	Status *datadogV1.NotebookStatus `json:"status,omitempty"`
	// Notebook global timeframe.
	Time datadogV1.NotebookGlobalTime `json:"time"`

	// Template vars
	TemplateVariables []NotebookVar `json:"template_variables"`

	// UnparsedObject contains the raw value of the object if there was an error when deserializing into the struct
	UnparsedObject       map[string]interface{} `json:-`
	AdditionalProperties map[string]interface{}
}

// NotebookCreateData The data for a notebook create request.
type NotebookCreateData struct {
	// The data attributes of a notebook.
	Attributes NotebookCreateDataAttributes `json:"attributes"`
	// Type of the Notebook resource.
	Type datadogV1.NotebookResourceType `json:"type"`
	// UnparsedObject contains the raw value of the object if there was an error when deserializing into the struct
	UnparsedObject       map[string]interface{} `json:-`
	AdditionalProperties map[string]interface{}
}

// NotebookCreateRequest The description of a notebook create request.
type NotebookCreateRequest struct {
	// The data for a notebook create request.
	Data NotebookCreateData `json:"data"`
	// UnparsedObject contains the raw value of the object if there was an error when deserializing into the struct
	UnparsedObject       map[string]interface{} `json:-`
	AdditionalProperties map[string]interface{}
}

type NotebookVar struct {
	Prefix          string   `json:"prefix"`
	Name            string   `json:"name"`
	Default         string   `json:"default"`
	AvailableValues []string `json:"available_values"`
}

func (c *CloneCmd) Run(opts *Options) error {
	fmt.Printf("Cloning notebook %d\n", c.ID)

	ctx := context.WithValue(
        context.Background(),
        datadog.ContextAPIKeys,
        map[string]datadog.APIKey{
            "apiKeyAuth": {
                Key: os.Getenv("DD_CLIENT_API_KEY"),
            },
            "appKeyAuth": {
                Key: os.Getenv("DD_CLIENT_APP_KEY"),
            },
        },
    )
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewNotebooksApi(apiClient)
	resp, r, err := api.GetNotebook(ctx, c.ID)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotebooksApi.GetNotebook`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")

	createRequest := &NotebookCreateRequest{}

	err = json.Unmarshal(responseContent, createRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to deserialize notebook: %v", err)
		return err
	}

	if len(c.Name) > 0 {
		createRequest.Data.Attributes.Name = c.Name
	}

	live := true
	if c.Start != nil && c.End != nil {
		createRequest.Data.Attributes.Time.NotebookAbsoluteTime.Start = *c.Start
		createRequest.Data.Attributes.Time.NotebookAbsoluteTime.End = *c.End
		live = false
	}
	createRequest.Data.Attributes.Time.NotebookAbsoluteTime.Live = &live

	for _, v := range c.Var {
		parts := strings.SplitN(v, ",", 3)
		if len(parts) < 3 {
			continue
		}
		createRequest.Data.Attributes.TemplateVariables = append(createRequest.Data.Attributes.TemplateVariables, NotebookVar{
			Name:    parts[0],
			Prefix:  parts[1],
			Default: parts[2],
		})
	}

	resp, r, err = createNotebookExecute(api, ctx, createRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotebooksApi.CreateNotebook`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	return nil
}

func createNotebookExecute(a *datadogV1.NotebooksApi, ctx context.Context, body *NotebookCreateRequest) (datadogV1.NotebookResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPost
		localVarPostBody    interface{}
		localVarReturnValue datadogV1.NotebookResponse
	)

	localBasePath, err := a.Client.Cfg.ServerURLWithContext(ctx, "v1.NotebooksApi.CreateNotebook")
	if err != nil {
		return localVarReturnValue, nil, datadog.GenericOpenAPIError{ErrorMessage: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/notebooks"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if body == nil {
		return localVarReturnValue, nil, datadog.ReportError("body is required and must be specified")
	}
	localVarHeaderParams["Content-Type"] = "application/json"
	localVarHeaderParams["Accept"] = "application/json"

	// body params
	localVarPostBody = body
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(datadog.ContextAPIKeys).(map[string]datadog.APIKey); ok {
			if apiKey, ok := auth["apiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["DD-API-KEY"] = key
			}
		}
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(datadog.ContextAPIKeys).(map[string]datadog.APIKey); ok {
			if apiKey, ok := auth["appKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["DD-APPLICATION-KEY"] = key
			}
		}
	}
	req, err := a.Client.PrepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, nil)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.Client.CallAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := datadog.GenericOpenAPIError{
			ErrorBody:    localVarBody,
			ErrorMessage: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v datadogV1.APIErrorResponse
			err = a.Client.Decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.ErrorModel = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v datadogV1.APIErrorResponse
			err = a.Client.Decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.ErrorModel = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v datadogV1.APIErrorResponse
			err = a.Client.Decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.ErrorModel = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.Client.Decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := datadog.GenericOpenAPIError{
			ErrorBody:    localVarBody,
			ErrorMessage: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
