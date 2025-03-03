//go:build test
// +build test

// All *_inject.go files are meant to be used by tests only. Purpose of this
// files is to provide a way to inject mocked data into the current setup.

package config

import (
	"github.com/cri-o/cri-o/internal/config/cnimgr"
	"github.com/cri-o/ocicni/pkg/ocicni"
)

// SetCNIPlugin sets the network plugin for the Configuration. The function
// errors if a sane shutdown of the initially created network plugin failed.
func (c *Config) SetCNIPlugin(plugin ocicni.CNIPlugin) error {
	if c.cniManager == nil {
		c.cniManager = &cnimgr.CNIManager{}
	}
	return c.cniManager.SetCNIPlugin(plugin)
}
