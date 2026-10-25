// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"math"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

const departmentsCRDPath = "../../config/crd/kai.resources_departments.yaml"

// departmentSchemaFromCRD returns the Department schema as it ships, so the specs below run the
// generated root-level rule rather than a copy of it.
func departmentSchemaFromCRD() *structuralschema.Structural {
	data, err := os.ReadFile(filepath.Clean(departmentsCRDPath))
	Expect(err).NotTo(HaveOccurred())

	var crd apiextensionsv1.CustomResourceDefinition
	Expect(yaml.Unmarshal(data, &crd)).To(Succeed())
	Expect(crd.Spec.Versions).To(HaveLen(1))

	props := &apiextensions.JSONSchemaProps{}
	Expect(apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(
		crd.Spec.Versions[0].Schema.OpenAPIV3Schema, props, nil)).To(Succeed())
	s, err := structuralschema.NewStructural(props)
	Expect(err).NotTo(HaveOccurred())
	return s
}

// createDepartment runs the CRD's CEL rules on a create of a Department with the given name and spec.
// The validator is built as the resource root, which is what makes metadata.name visible to the rule.
func createDepartment(s *structuralschema.Structural, name string, spec map[string]interface{}) field.ErrorList {
	obj := map[string]interface{}{"metadata": map[string]interface{}{"name": name}}
	if spec != nil {
		obj["spec"] = spec
	}
	errs, _ := cel.NewValidator(s, true, math.MaxUint32).
		Validate(context.Background(), field.NewPath(""), s, obj, nil, math.MaxInt64)
	return errs
}

var _ = Describe("Department parent CRD rule", func() {
	var s *structuralschema.Structural

	BeforeEach(func() {
		s = departmentSchemaFromCRD()
	})

	It("rejects a department that is its own parent", func() {
		errs := createDepartment(s, "research", map[string]interface{}{"parent": "research", "queues": []interface{}{}})
		Expect(errs.ToAggregate()).To(MatchError(ContainSubstring("a department cannot be its own parent")))
	})

	It("accepts a department under another department", func() {
		errs := createDepartment(s, "ml", map[string]interface{}{"parent": "research", "queues": []interface{}{}})
		Expect(errs).To(BeEmpty())
	})

	It("accepts a top-level department", func() {
		Expect(createDepartment(s, "research", map[string]interface{}{"queues": []interface{}{}})).To(BeEmpty())
	})

	It("accepts a department with no spec", func() {
		Expect(createDepartment(s, "research", nil)).To(BeEmpty())
	})
})
