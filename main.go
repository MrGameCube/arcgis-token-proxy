package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
)

type Service struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	Token                string `json:"token"`
	TokenExpiration      string `json:"tokenExpiration"`
	DefinitionExpression string `json:"definitionExpression"`
}

func main() {
	router := gin.Default()

	// Static const Map from ID to url
	idToURL := map[string]Service{
		"1-abc": {
			ID:    "1",
			Name:  "USA_ZIP_Code_Points_analysis",
			URL:   "https://services.arcgis.com/P3ePLMYs2RVChkJx/ArcGIS/rest/services/USA_ZIP_Code_Points_analysis/FeatureServer/0",
			Token: "1234567890",
		},
		// Add more mappings here
	}

	router.GET("/service/:id/*action", func(c *gin.Context) {
		id := c.Param("id")
		action := c.Param("action")
		query := c.Request.URL.Query()

		token := query.Get("token")
		// Find the URL for the given ID
		urlInfo, exists := idToURL[id+"-"+token]
		if !exists {
			c.String(http.StatusNotFound, "No URL found for ID: %s", id)
			return
		}
		baseURL := urlInfo.URL
		query.Set("token", urlInfo.Token)
		// Construct new URL with action and query parameters
		newURL, err := url.Parse(baseURL)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error parsing base URL: %v", err)
			return
		}
		newURL.Path += action
		rawQuery := query.Encode()
		if rawQuery != "" {
			newURL.RawQuery = rawQuery
		}

		// Forward the request
		resp, err := http.Get(newURL.String())
		if err != nil {
			c.String(http.StatusInternalServerError, "Error forwarding request: %v", err)
			return
		}
		defer resp.Body.Close()

		// Stream the response back
		c.Status(resp.StatusCode)
		for key, value := range resp.Header {
			c.Header(key, value[0])
		}
		io.Copy(c.Writer, resp.Body)
	})

	// Run the server
	router.RunTLS(":8080", "cert.pem", "key.pem") // By default, it serves on :8080
}
