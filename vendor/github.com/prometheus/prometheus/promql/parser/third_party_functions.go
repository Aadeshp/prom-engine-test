// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

// NB: this is done in an init method to have the least amount of interference
// with the rest of the Prometheus parser, allowing for easier merging of
// upstream changes.
func init() {
	functionsToUpdate := []string{
		"avg_over_time",
		"changes",
		"count_over_time",
		"delta",
		"deriv",
		"holt_winters",
		"idelta",
		"increase",
		"irate",
		"max_over_time",
		"min_over_time",
		"predict_linear",
		"quantile_over_time",
		"rate",
		"resets",
		"stddev_over_time",
		"stdvar_over_time",
		"sum_over_time",
	}

	functionsLock.Lock()
	defer functionsLock.Unlock()
	for _, fn := range functionsToUpdate {
		// Sanity check to ensure the key exists (it should unless Prometheus
		// renames functions).
		if _, ok := Functions[fn]; !ok {
			continue
		}

		Functions[fn].Variadic = 1
		Functions[fn].ArgTypes = append(Functions[fn].ArgTypes, ValueTypeString)
	}
}
