package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Request struct {
	Method string `json:"method"`
	Path []string `json:"path"`
	User string `json:"user"`
}

type InputRequest struct {
	Input Request `json:"input"`
}

func main(){
	r := gin.Default()
	opaUrlString, found := os.LookupEnv("OPA_ADDR")
	matchBasicAuth := regexp.MustCompile("^Basic (.*)$")

	if !found {
		opaUrlString = "http://opa:8181"
	}

	opaUrl, err := url.Parse(opaUrlString)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		return
	}



	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Handle /api/v1 route
	apiGroup := r.Group( "/api/v1/")

	// Create a custom middleware to handle the authorization
	apiGroup.Use(func(c *gin.Context) {
		// Eval policy
		req := c.Request
		auth := req.Header.Get("Authorization")
		user := ""

		if matchBasicAuth.MatchString(auth) {
			basicData := matchBasicAuth.FindAllStringSubmatch(auth, -1)[0][1]
			byteArr, err := base64.StdEncoding.DecodeString(basicData)
			if err == nil {
				userPass := strings.Split(string(byteArr), ":")
				user = userPass[0]
			}
		}

		// Evaluate policy using OPA
		result := evalPolicy(opaUrl, InputRequest{
			Request{
				Method: c.Request.Method,
				Path:   strings.Split(c.Request.URL.Path, "/")[1:],
				User:   user,
			},
		})

		logrus.Infof("policy result is %v", result)


		if result {
			// Policy says that we're allowed to continue
			c.Next()
		} else {
			// Return a 403
			c.Status(http.StatusForbidden)
			c.YAML(http.StatusForbidden, map[string]interface{}{
				"status": false,
				"error": "You do not have access to this resource",
			})
			c.Abort()
		}
	})

	// Dummy endpoint that returns :name (e.g: /employees/alice returns alice
	apiGroup.GET("/employees/:name", func(c *gin.Context) {
		name, _ := c.Params.Get("name")
		c.YAML(http.StatusOK, map[string]string{
			"name": name,
		})
	})
	r.Run()
}

type PolicyResponse struct {
	Result bool `json:"result"`
}

func evalPolicy(opaUrl *url.URL, input InputRequest) bool {
	relativeUrl, err := url.Parse("/v1/data/swisscom/example/allow")
	if err != nil {
		panic(err)
	}
	body, err := json.Marshal(input)

	if err != nil {
		logrus.Error(err)
		return false
	}

	logrus.Infof("policy req: %v", string(body))

	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, opaUrl.ResolveReference(relativeUrl).String(), bodyReader)

	if err != nil {
		logrus.Error(err)
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Error(err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("invalid response code %d, %d expected", resp.StatusCode, http.StatusOK)
	}

	jsonDecoder := json.NewDecoder(resp.Body)
	var response *PolicyResponse
	err = jsonDecoder.Decode(&response)

	if err != nil {
		logrus.Error(err)
		return false
	}

	return response.Result
}
