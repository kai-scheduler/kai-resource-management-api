// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

// queueMinRuntimePatterns returns the patterns of the queue minRuntime fields in the CRD as it ships.
func queueMinRuntimePatterns(path string) map[string]*regexp.Regexp {
	data, err := os.ReadFile(filepath.Clean(path))
	Expect(err).NotTo(HaveOccurred())

	var crd apiextensionsv1.CustomResourceDefinition
	Expect(yaml.Unmarshal(data, &crd)).To(Succeed())
	Expect(crd.Spec.Versions).To(HaveLen(1))

	queue := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties["queues"].Items.Schema
	patterns := map[string]*regexp.Regexp{}
	for name, field := range queue.Properties["minRuntime"].Properties {
		patterns[name] = regexp.MustCompile(field.Pattern)
	}
	Expect(patterns).To(HaveKey("preemptMinRuntime"))
	Expect(patterns).To(HaveKey("reclaimMinRuntime"))
	return patterns
}

var _ = Describe("Queue minRuntime", func() {
	for _, crd := range []string{
		"../../config/crd/kai.resources_projects.yaml",
		"../../config/crd/kai.resources_departments.yaml",
	} {
		DescribeTable(filepath.Base(crd),
			func(value string, accepted bool) {
				for name, pattern := range queueMinRuntimePatterns(crd) {
					Expect(pattern.MatchString(value)).To(Equal(accepted), name)
				}
				if accepted {
					// Whatever the schema accepts must decode, or the controller cannot read the object.
					data, err := json.Marshal(map[string]string{"preemptMinRuntime": value, "reclaimMinRuntime": value})
					Expect(err).NotTo(HaveOccurred())
					Expect(json.Unmarshal(data, &QueueMinRuntime{})).To(Succeed())
				}
			},
			Entry("accepts minutes", "5m", true),
			Entry("accepts combined units", "1h30m", true),
			Entry("accepts a fraction", "1.5h", true),
			Entry("accepts zero", "0", true),
			Entry("rejects days, which Go durations lack", "1d", false),
			Entry("rejects a negative duration", "-5m", false),
			Entry("rejects words", "5 minutes", false),
		)
	}
})
