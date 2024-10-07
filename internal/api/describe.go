package api

import (
	"regexp"
	"sort"
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
)

type DescribeRequest struct {
	core.APIRequest
}

type DescribeResponse struct {
	Endpoints []string `json:"endpoints"`
}

func HandleDescribe(request *DescribeRequest) (*DescribeResponse, *core.APIError) {
	re := regexp.MustCompile(`(\^/api/v\d{2}/)|(\$)`)

	routes := GetRoutes()

	descriptions := make([]string, 0, len(*routes))
	for _, route := range *routes {
		description := route.Describe()

		if strings.Contains(description, "api") {
			cleanedDescription := re.ReplaceAllString(description, "")
			descriptions = append(descriptions, cleanedDescription)
		}
	}

	sort.Strings(descriptions)

	response := DescribeResponse{descriptions}

	return &response, nil
}
