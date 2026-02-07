package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/catalisaio/workflow-core/pkg/types"
)

// HTTPRequestNode implements the HTTP Request node.
type HTTPRequestNode struct{}

// Type returns the n8n node type.
func (n *HTTPRequestNode) Type() string {
	return "n8n-nodes-base.httpRequest"
}

// Execute performs HTTP requests.
func (n *HTTPRequestNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		input = types.EmptyNodeData()
	}

	var results []types.NodeData

	for _, item := range input {
		// Set up evaluator with current item data
		ctx.Evaluator().SetData(item.JSON)

		// Get request parameters
		method := strings.ToUpper(getStringParam(params, "method", "GET"))
		urlStr := getStringParam(params, "url", "")

		// Evaluate URL expression
		if urlStr != "" {
			evaluated, err := ctx.Evaluator().Evaluate(urlStr)
			if err == nil {
				urlStr = interfaceToString(evaluated)
			}
		}

		if urlStr == "" {
			return nil, fmt.Errorf("URL is required")
		}

		// Build request body
		var bodyReader io.Reader
		contentType := "application/json"

		if method != "GET" && method != "HEAD" && method != "DELETE" {
			bodyType := getStringParam(params, "bodyContentType", "json")

			switch bodyType {
			case "json":
				if jsonBody := getMapParam(params, "body"); jsonBody != nil {
					// Evaluate expressions in body
					evaluated, err := ctx.Evaluator().Evaluate(jsonBody)
					if err == nil {
						jsonBody = evaluated.(map[string]interface{})
					}
					bodyBytes, _ := json.Marshal(jsonBody)
					bodyReader = bytes.NewReader(bodyBytes)
					contentType = "application/json"
				} else if bodyStr := getStringParam(params, "body", ""); bodyStr != "" {
					evaluated, _ := ctx.Evaluator().Evaluate(bodyStr)
					bodyReader = strings.NewReader(interfaceToString(evaluated))
				}
			case "raw":
				bodyStr := getStringParam(params, "body", "")
				evaluated, _ := ctx.Evaluator().Evaluate(bodyStr)
				bodyReader = strings.NewReader(interfaceToString(evaluated))
				contentType = getStringParam(params, "rawContentType", "text/plain")
			case "form-urlencoded":
				formData := url.Values{}
				if formParams := getSliceParam(params, "bodyParameters"); formParams != nil {
					for _, p := range formParams {
						param, ok := p.(map[string]interface{})
						if !ok {
							continue
						}
						name := getStringParam(param, "name", "")
						value := getStringParam(param, "value", "")
						if name != "" {
							evaluated, _ := ctx.Evaluator().Evaluate(value)
							formData.Set(name, interfaceToString(evaluated))
						}
					}
				}
				bodyReader = strings.NewReader(formData.Encode())
				contentType = "application/x-www-form-urlencoded"
			}
		}

		// Create HTTP request
		reqCtx, cancel := context.WithTimeout(ctx.Context(), 30*time.Second)

		req, err := http.NewRequestWithContext(reqCtx, method, urlStr, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("User-Agent", "workflow-core/1.0")

		if headers := getSliceParam(params, "headerParameters"); headers != nil {
			for _, h := range headers {
				header, ok := h.(map[string]interface{})
				if !ok {
					continue
				}
				name := getStringParam(header, "name", "")
				value := getStringParam(header, "value", "")
				if name != "" {
					evaluated, _ := ctx.Evaluator().Evaluate(value)
					req.Header.Set(name, interfaceToString(evaluated))
				}
			}
		}

		// Add query parameters
		if queryParams := getSliceParam(params, "queryParameters"); queryParams != nil {
			q := req.URL.Query()
			for _, qp := range queryParams {
				param, ok := qp.(map[string]interface{})
				if !ok {
					continue
				}
				name := getStringParam(param, "name", "")
				value := getStringParam(param, "value", "")
				if name != "" {
					evaluated, _ := ctx.Evaluator().Evaluate(value)
					q.Add(name, interfaceToString(evaluated))
				}
			}
			req.URL.RawQuery = q.Encode()
		}

		// Execute request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var output map[string]interface{}
		var jsonResponse interface{}
		if err := json.Unmarshal(respBody, &jsonResponse); err == nil {
			if obj, ok := jsonResponse.(map[string]interface{}); ok {
				output = obj
			} else {
				output = map[string]interface{}{"data": jsonResponse}
			}
		} else {
			output = map[string]interface{}{"data": string(respBody)}
		}

		results = append(results, types.NodeData{
			JSON: output,
		})
	}

	return results, nil
}
