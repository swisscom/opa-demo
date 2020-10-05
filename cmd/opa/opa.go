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

type InputRequest struct {
	Method string `json:"method"`
	Path []string `json:"path"`
	User string `json:"user"`
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

	apiGroup := r.Group( "/api/v1/")
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

		result := evalPolicy(opaUrl, InputRequest{
			Method:  c.Request.Method,
			Path:    strings.Split(c.Request.URL.Path, "/")[1:],
			User: user,
		})

		logrus.Infof("policy result is %v", result)

		if result {
			c.Next()
		} else {
			c.Status(http.StatusForbidden)
			c.Abort()
		}
	})
	apiGroup.GET("/employees/:name", func(c *gin.Context) {
		c.YAML(http.StatusOK, map[string]string{
			"name": "Fake Name",
		})
	})
	r.Run()
}

type PolicyResponse struct {
	Result bool `json:"result"`
}

func evalPolicy(opaUrl *url.URL, input InputRequest) bool {
	relativeUrl, err := url.Parse("/v1/data/example/authz/allow")
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
