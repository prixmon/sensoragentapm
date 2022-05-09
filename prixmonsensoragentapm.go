package prixmonsensoragentapm

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type SensorAPMIdentity struct {
	SensorAgentCommunicatorURL string `default:"prxsensor.prixmon.com"` // Sensor Agent Communicator URL
	OwnerId                    string // Owner ID which is provided inside Prixmon dashboard
	ProjectId                  string // Project Id
	Username                   string // APM Username
	Password                   string // APM Password
	IsSecure                   bool   //HTTP or HTTPS
}

var sensorAPMData map[string]time.Time

var prvSensorAPMID SensorAPMIdentity

//InitSensorAPM - To initialize APM
func InitSensorAPM(sensorAPMIdentity SensorAPMIdentity) {
	prvSensorAPMID = sensorAPMIdentity
	sensorAPMData = make(map[string]time.Time)
}

//StartMetric - Start Timer
func StartMetric(methodName string) bool {
	result := false
	if _, itemExists := sensorAPMData[methodName]; !itemExists {
		sensorAPMData[methodName] = time.Now()
		result = !itemExists
	}
	return result
}

//FinilizeMetric - End Timer and submit Data
func FinilizeMetric(methodName string) time.Duration {
	// Submit to Sensor Agent APM Communicator
	methodRunDuration := time.Since(sensorAPMData[methodName])
	// HTTP or HTTPS
	//in case of any other web server errors returns Communicator result
	fmt.Println("Log for: " + methodName)
	go submitToCommunicator(methodName, methodRunDuration, sensorAPMData[methodName], time.Now())
	// Clear Selected Metric
	delete(sensorAPMData, methodName)
	return methodRunDuration
}

func submitToCommunicator(methodName string, methodRunDuration time.Duration, startTime, EndTime time.Time) {
	authData := bytes.NewBuffer([]byte(`U=` + prvSensorAPMID.Username + "&P=" + prvSensorAPMID.Password))

	communicatorURL := prvSensorAPMID.SensorAgentCommunicatorURL + "/saveapm?oid=" + prvSensorAPMID.OwnerId + "&pid=" + url.QueryEscape(prvSensorAPMID.ProjectId) + "&pm=" + url.QueryEscape(methodName) + "&tm=" + strconv.FormatInt(methodRunDuration.Milliseconds(), 10) + "&st=" + startTime.Format("2006-01-02 15:04") + "&et=" + EndTime.Format("2006-01-02 15:04")

	if prvSensorAPMID.IsSecure {
		communicatorURL = "https://" + communicatorURL
	} else {
		communicatorURL = "http://" + communicatorURL
	}

	serviceCallResponse, err := http.Post(communicatorURL, "application/x-www-form-urlencoded", authData)

	if err != nil || serviceCallResponse.StatusCode != 200 {

		if err == nil {
			err = errors.New("Http Error " + strconv.Itoa(serviceCallResponse.StatusCode))
		}
	}

}
