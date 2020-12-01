/*
 * Copyright 2020 Mandelsoft. All rights reserved.
 *  This file is licensed under the Apache Software License, v. 2 except as noted
 *  otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package ipxe

import (
	"fmt"
	"strings"
	"time"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cert"
)

const CERT_NONE = "none"
const CERT_MANAGE = "manage"
const CERT_USE = "use"

type Config struct {
	set config.OptionSet

	LocalNamespaceOnly bool
	PXEPort            int
	CacheDir           string
	CacheTTL           time.Duration

	TraceRequest bool

	CertMode string
	TLS      bool
	BasePath string
	Cert     *cert.CertConfig
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	this.set = set
	set.AddStringOption(&this.CacheDir, "cache-dir", "", "", "enable URL caching in a dedicated directory")
	set.AddDurationOption(&this.CacheTTL, "cache-ttl", "", 10*time.Minute, "TTL for cache entries")
	set.AddBoolOption(&this.LocalNamespaceOnly, "local-namespace-only", "", false, "server only resources in local namespace")
	set.AddBoolOption(&this.TraceRequest, "trace-requests", "", false, "trace mapping of request data")
	set.AddIntOption(&this.PXEPort, "pxe-port", "", 8081, "pxe server port")
	set.AddStringOption(&this.BasePath, "base-path", "", "", "pxe server URL base path")

	set.AddBoolOption(&this.TLS, "use-tls", "", false, "use https")
	set.AddStringOption(&this.CertMode, "certificate-mode", "", "manage", "mode for cert management")
	this.Cert = cert.NewCertConfig("kipxe", "")
	this.Cert.AddOptionsToSet(set)
}

func (this *Config) Prepare() error {
	if this.BasePath == "" {
		this.BasePath = "/"
	} else {
		if !strings.HasPrefix(this.BasePath, "/") {
			this.BasePath = "/" + this.BasePath
		}
	}
	if this.TLS {
		if this.set != nil {
			opt := this.set.GetOption("pxe-port")
			if !opt.Changed() {
				this.PXEPort = 8443
			}
		}
		if this.CertMode == CERT_NONE {
			return fmt.Errorf("certificate handling required for TLS mode")
		}
		if this.Cert.Secret == "" && this.Cert.CertFile == "" {
			return fmt.Errorf("secret or cerificate file required for TLS mode")
		}
	} else {
		opt := this.set.GetOption("certificate-mode")
		if !opt.Changed() && this.Cert.Secret == "" && this.Cert.CertFile == "" {
			this.CertMode = CERT_NONE
		}
	}
	if this.CertMode == CERT_MANAGE {
		if this.Cert.Secret == "" {
			return fmt.Errorf("secret required for managed certificate")
		}
	}
	if this.CertMode != CERT_NONE {
		if this.Cert.Secret == "" && this.Cert.CertFile == "" {
			return fmt.Errorf("secret or cerificate file required for providing certificate")
		}
	}

	return nil
}
