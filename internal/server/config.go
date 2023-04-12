/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.

 * This file is based on the links found at https://github.com/kubernetes/kubernetes/tree/release-1.21/test/images/agnhost/webhook
 * and is meant as a basic "I'm learning how to Go via mutating admission controller" type project - according to the Kubernetes
 * docs at https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#write-an-admission-webhook-server
 * this is a working and valid example of how to make a Golang webserver that handles admission control requests, and can serve
 * as a model for picking this activity up. This is **not** production ready, this is **not** something you can plug and play.
 */

package server

import (
	"github.com/spf13/viper"
)

type IConfig interface {
	getCertFile() string
	setCertFile(certFile string)
	getKeyFile() string
	setKeyFile(keyFile string)
	getMessage() string
	setMessage(message string)
	getHttpPort() int
	setHttpPort(port int)
	getHttpsPort() int
	setHttpsPort(port int)
}

type Config struct {
	CertFile  string
	KeyFile   string
	Message   string
	HttpPort  int
	HttpsPort int
}

func (c *Config) getCertFile() string {
	return c.CertFile
}

func (c *Config) setCertFile(certFile string) {
	c.CertFile = certFile
}

func (c *Config) getKeyFile() string {
	return c.KeyFile
}

func (c *Config) setKeyFile(keyFile string) {
	c.KeyFile = keyFile
}

func (c *Config) getMessage() string {
	return c.Message
}

func (c *Config) setMessage(message string) {
	c.Message = message
}

func (c *Config) getHttpPort() int {
	return c.HttpPort
}

func (c *Config) setHttpPort(port int) {
	c.HttpPort = port
}

func (c *Config) getHttpsPort() int {
	return c.HttpsPort
}

func (c *Config) setHttpsPort(port int) {
	c.HttpsPort = port
}

func newConfig() IConfig {
	// set viper config defaults
	return &Config{
		CertFile:  viper.GetString("certFile"),
		KeyFile:   viper.GetString("keyFile"),
		Message:   viper.GetString("message"),
		HttpPort:  viper.GetInt("http-port"),
		HttpsPort: viper.GetInt("https-port"),
	}
}

func setConfigDefaults() {
	viper.SetDefault("cert-file", "")
	viper.SetDefault("key-file", "")
	viper.SetDefault("message", "Hello World!")
	viper.SetDefault("port", 5001)
}
