/*
 Copyright 2022 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package checksum

import (
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// Checksums is a list of checksums.
type Checksums struct {
	checksums []Interface
	value     string
}

// NewChecksums returns a new Checksums.
func NewChecksums(checksums ...Interface) *Checksums {
	return &Checksums{checksums: checksums}
}

// Get gets the checksums. It will iterate through the Get() of the []Interface array and store the first fetched value.
func (c *Checksums) Get() error {
	if c.value != "" {
		return nil
	}

	var errs []error
	for _, v := range c.checksums {
		if err := v.Get(); err != nil {
			errs = append(errs, err)
			continue
		}
		c.value = v.Value()
		return nil
	}
	return kerrors.NewAggregate(errs)
}

// Value returns the checksums.
func (c *Checksums) Value() string {
	return c.value
}

// Append appends checksums.
func (c *Checksums) Append(checksums ...Interface) {
	c.checksums = append(c.checksums, checksums...)
}
