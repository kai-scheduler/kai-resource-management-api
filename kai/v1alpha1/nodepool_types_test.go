// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel/model"
	"k8s.io/apimachinery/pkg/util/validation/field"
	celconfig "k8s.io/apiserver/pkg/apis/cel"
	"k8s.io/apiserver/pkg/cel/environment"
	"sigs.k8s.io/yaml"
)

func TestKaiV1alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "kai.resources/v1alpha1 Custom Resources")
}

// crdPath points at the generated NodePool CRD, relative to this package dir.
const crdPath = "../../config/crd/kai.resources_nodepools.yaml"

// immutabilityRuleFromCRD reads the labelKey/labelValue immutability rule
// straight out of the generated CRD so this test always exercises whatever rule
// actually ships. If the marker is ever changed to a form that breaks the
// default nodepool (e.g. a plain "self.labelKey == oldSelf.labelKey" that errors
// on absent fields), the specs below fail.
func immutabilityRuleFromCRD() string {
	data, err := os.ReadFile(filepath.Clean(crdPath))
	Expect(err).NotTo(HaveOccurred())

	var crd apiextensionsv1.CustomResourceDefinition
	Expect(yaml.Unmarshal(data, &crd)).To(Succeed())

	for _, v := range crd.Spec.Versions {
		if v.Schema == nil || v.Schema.OpenAPIV3Schema == nil {
			continue
		}
		spec, ok := v.Schema.OpenAPIV3Schema.Properties["spec"]
		if !ok {
			continue
		}
		for _, r := range spec.XValidations {
			if strings.Contains(r.Message, "immutable") {
				return r.Rule
			}
		}
	}
	Fail("no labelKey/labelValue immutability rule found in " + crdPath)
	return ""
}

// evaluateSpecRule runs the given CEL rule (as an x-kubernetes-validations rule
// on the spec) through the same validator the API server uses, and reports
// whether the transition from oldObj to newObj is allowed.
func evaluateSpecRule(rule string, oldObj, newObj map[string]interface{}) (allowed bool, msg string) {
	props := &apiextensions.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensions.JSONSchemaProps{
			"labelKey":   {Type: "string"},
			"labelValue": {Type: "string"},
		},
		XValidations: apiextensions.ValidationRules{{Rule: rule, Message: "labelKey and labelValue are immutable"}},
	}
	s, err := structuralschema.NewStructural(props)
	Expect(err).NotTo(HaveOccurred())

	v := cel.NewValidator(s, true, math.MaxUint32)
	errs, _ := v.Validate(context.Background(), field.NewPath("spec"), s, newObj, oldObj, math.MaxInt64)
	if len(errs) == 0 {
		return true, ""
	}
	return false, errs.ToAggregate().Error()
}

// Field representations. With omitempty, an unset field is absent from the
// stored object (empty map); a field sent explicitly as "" is present-but-empty.
var (
	labelsAbsent  = map[string]interface{}{}
	labelsGPUA100 = map[string]interface{}{"labelKey": "gpu", "labelValue": "a100"}
	labelsGPUH100 = map[string]interface{}{"labelKey": "gpu", "labelValue": "h100"}
)

var _ = Describe("NodePool labelKey/labelValue immutability CRD rule", func() {
	var rule string

	BeforeEach(func() {
		rule = immutabilityRuleFromCRD()
	})

	It("allows a no-op update of the default nodepool (both fields absent)", func() {
		allowed, msg := evaluateSpecRule(rule, labelsAbsent, labelsAbsent)
		Expect(allowed).To(BeTrue(), "default no-op must be allowed, got: %s", msg)
	})

	It("allows a no-op update of a labeled nodepool", func() {
		allowed, msg := evaluateSpecRule(rule, labelsGPUA100, labelsGPUA100)
		Expect(allowed).To(BeTrue(), "labeled no-op must be allowed, got: %s", msg)
	})

	It("rejects changing the labelValue", func() {
		allowed, _ := evaluateSpecRule(rule, labelsGPUA100, labelsGPUH100)
		Expect(allowed).To(BeFalse(), "changing the pair must be rejected")
	})

	It("rejects clearing a labeled nodepool's pair", func() {
		allowed, _ := evaluateSpecRule(rule, labelsGPUA100, labelsAbsent)
		Expect(allowed).To(BeFalse(), "clearing the pair must be rejected")
	})

	It("rejects adding a label pair to the default nodepool", func() {
		allowed, _ := evaluateSpecRule(rule, labelsAbsent, labelsGPUA100)
		Expect(allowed).To(BeFalse(), "adding a pair to the default nodepool must be rejected")
	})
})

// staticEstimatedCostLimit is apiextensions validation.StaticEstimatedCostLimit, the highest estimated
// cost the API server accepts for one CRD rule. Importing it would pull the API server into go.mod.
const staticEstimatedCostLimit = 10000000

// minRuntimeSchemaFromCRD returns the schema of schedulingShardConfig.minRuntime as it ships.
func minRuntimeSchemaFromCRD() *structuralschema.Structural {
	data, err := os.ReadFile(filepath.Clean(crdPath))
	Expect(err).NotTo(HaveOccurred())

	var crd apiextensionsv1.CustomResourceDefinition
	Expect(yaml.Unmarshal(data, &crd)).To(Succeed())
	Expect(crd.Spec.Versions).To(HaveLen(1))

	props := &apiextensions.JSONSchemaProps{}
	Expect(apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(
		crd.Spec.Versions[0].Schema.OpenAPIV3Schema, props, nil)).To(Succeed())
	s, err := structuralschema.NewStructural(props)
	Expect(err).NotTo(HaveOccurred())
	minRuntime := s.Properties["spec"].Properties["schedulingShardConfig"].Properties["minRuntime"]
	return &minRuntime
}

var _ = Describe("NodePool minRuntime CRD rules", func() {
	var s *structuralschema.Structural

	BeforeEach(func() {
		s = minRuntimeSchemaFromCRD()
	})

	It("compiles within the API server's cost limit", func() {
		results, err := cel.Compile(s, model.SchemaDeclType(s, false), celconfig.PerCallLimit,
			environment.MustBaseEnvSet(environment.DefaultCompatibilityVersion()), cel.StoredExpressionsEnvLoader())
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))
		for _, result := range results {
			Expect(result.Error).To(BeNil())
			Expect(result.MaxCost).To(BeNumerically("<=", staticEstimatedCostLimit))
		}
	})

	DescribeTable("validates both values as durations",
		func(value string, accepted bool) {
			obj := map[string]interface{}{"preemptMinRuntime": value, "reclaimMinRuntime": value}
			errs, _ := cel.NewValidator(s, false, celconfig.PerCallLimit).
				Validate(context.Background(), field.NewPath("minRuntime"), s, obj, nil, celconfig.RuntimeCELCostBudget)
			if accepted {
				Expect(errs).To(BeEmpty())
			} else {
				Expect(errs).To(HaveLen(2))
			}
		},
		Entry("accepts minutes", "5m", true),
		Entry("accepts combined units", "1h30m", true),
		Entry("accepts a fraction", "1.5h", true),
		Entry("accepts zero", "0", true),
		Entry("rejects days, which KAI cannot parse", "1d", false),
		Entry("rejects a negative duration", "-5m", false),
		Entry("rejects words", "5 minutes", false),
	)

	It("accepts a minRuntime with neither value set", func() {
		errs, _ := cel.NewValidator(s, false, celconfig.PerCallLimit).
			Validate(context.Background(), field.NewPath("minRuntime"), s, map[string]interface{}{}, nil, celconfig.RuntimeCELCostBudget)
		Expect(errs).To(BeEmpty())
	})
})
