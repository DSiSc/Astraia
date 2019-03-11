package config

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestGetApiGatewayHostName(t *testing.T) {
	hostname := GetApiGatewayHostName()
	assert.Equal(t, "127.0.0.1", hostname)
}

func TestGetApiGatewayPort(t *testing.T) {
	port := GetApiGatewayPort()
	assert.Equal(t, "47768", port)
}
