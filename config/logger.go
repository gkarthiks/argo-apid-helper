package config

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"time"
)

// InitializeLogger will initialize the logger with certain format.
// Uses the environment variable to determine the mode of starting the logger and gin router.
// This can't be dynamically overridden.
func InitializeLogger() {
	Log = logrus.New()

	if AppMode == "production" {
		gin.SetMode(gin.ReleaseMode)
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		// Global setting is required because of the discover-k8s module
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		Log.SetLevel(logrus.InfoLevel)
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			PrettyPrint:     true,
		})
		// Global setting is required because of the discover-k8s module
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			PrettyPrint:     true,
		})
		Log.SetLevel(logrus.DebugLevel)
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// Logger injects the middleware function for the gin-gonic router.
// This enables the additional fields with user preferred log framework; logrus in this case.
func Logger(logger logrus.FieldLogger) gin.HandlerFunc {
	serverAddr, _ := os.Hostname()
	appVersion := AppVersion

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		queryParam := c.Request.URL.RawQuery

		c.Next()

		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()

		if referer == "" {
			referer = "n/a"
		}
		if queryParam == "" {
			queryParam = "n/a"
		}
		if statusCode > 499 {
			logger.Errorf("appVersion: '%s', hostname: %s, clientIP: %s, method: %s, path: %s, queryParam: %s, statusCode: %d, dataLength: %d, latency: %dms, referer: %s, userAgent: %s", appVersion, serverAddr, clientIP, c.Request.Method, path, queryParam, statusCode, dataLength, latency, referer, userAgent)
		} else if statusCode > 399 {
			logger.Warnf("appVersion: '%s', hostname: %s, clientIP: %s, method: %s, path: %s, queryParam: %s, statusCode: %d, dataLength: %d, latency: %dms, referer: %s, userAgent: %s", appVersion, serverAddr, clientIP, c.Request.Method, path, queryParam, statusCode, dataLength, latency, referer, userAgent)
		} else {
			logger.Infof("appVersion: '%s', hostname: %s, clientIP: %s, method: %s, path: %s, queryParam: %s, statusCode: %d, dataLength: %d, latency: %dms, referer: %s, userAgent: %s", appVersion, serverAddr, clientIP, c.Request.Method, path, queryParam, statusCode, dataLength, latency, referer, userAgent)
		}
	}
}
