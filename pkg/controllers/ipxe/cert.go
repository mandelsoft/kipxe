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
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/logger"

	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type CertMapper struct {
	reconciler *reconciler
}

func NewCertMapper(r *reconciler) kipxe.MetaDataMapper {
	return &CertMapper{
		r,
	}
}

func (this *CertMapper) String() string {
	return "cert enricher"
}

func (this *CertMapper) Weight() int {
	return 0
}

func (this *CertMapper) Map(logger logger.LogContext, values kipxe.MetaData, req *http.Request) (kipxe.MetaData, error) {
	ca := this.reconciler.cert.GetCertificateInfo().CACert()
	values["CACERT"] = string(ca)
	return values, nil
}
