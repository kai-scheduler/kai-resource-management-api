// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

// +kubebuilder:object:generate:=true
package v1alpha1

// QueueResourcesConfig defines resource allocations for GPU, CPU, and Memory
type QueueResourcesConfig struct {
	// GPU resources in fractions (e.g. 0.7 = 70% of a GPU)
	GPU SystemResource `json:"gpu,omitempty"`

	// CPU resources in millicpus (e.g. 1000 = 1 CPU)
	CPU SystemResource `json:"cpu,omitempty"`

	// Memory resources in megabytes (1 MB = 10^6 bytes)
	Memory SystemResource `json:"memory,omitempty"`
}

// SystemResource defines resource limits and quotas
type SystemResource struct {
	// +optional
	Deserved float64 `json:"deserved"`
	// +optional
	OverQuotaWeight float64 `json:"overQuotaWeight"`
	// +optional
	Limit float64 `json:"limit"`
}

type QueueConfig struct {
	// Queue name
	Name string `json:"name"`

	// Nodepool associated with the queue
	Nodepool string `json:"nodepool"`

	// Queue resources definition
	Resources *QueueResourcesConfig `json:"resources,omitempty"`

	// Queue priority
	Priority *int32 `json:"priority,omitempty"`

	// Minimum runtime of a workload in this queue before it can be preempted or reclaimed.
	// Unset, the value comes from the parent department's queue, then from the node pool.
	// +optional
	MinRuntime *QueueMinRuntime `json:"minRuntime,omitempty"`
}

// QueueMinRuntime sets how long a workload runs before it can be interrupted.
// Values use the node pool's duration format, where d is 24h and w is 7d.
type QueueMinRuntime struct {
	// How long a workload runs before a higher-priority workload may preempt it, such as "5m" or "2d".
	// +optional
	// +kubebuilder:validation:Pattern=`^(0|([0-9]+(\.[0-9]+)?(ns|us|ms|s|m|h|d|w))+)$`
	PreemptMinRuntime *string `json:"preemptMinRuntime,omitempty"`

	// How long a workload runs before another queue may reclaim the quota it borrowed, such as "10m".
	// +optional
	// +kubebuilder:validation:Pattern=`^(0|([0-9]+(\.[0-9]+)?(ns|us|ms|s|m|h|d|w))+)$`
	ReclaimMinRuntime *string `json:"reclaimMinRuntime,omitempty"`
}
