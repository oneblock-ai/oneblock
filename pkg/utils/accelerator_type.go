package utils

import "strings"

const (
	NvidiaTeslaV100      = "V100"
	NvidiaTeslaP100      = "P100"
	NvidiaTeslaT4        = "T4"
	NvidiaTeslaP4        = "P4"
	NvidiaTeslaK80       = "K80"
	NvidiaTeslaA10g      = "A10G"
	NvidiaL4             = "L4"
	NvidiaA100           = "A100"
	IntelMax1550         = "Intel-GPU-Max-1550"
	IntelMax1100         = "Intel-GPU-Max-1100"
	IntelGaudi           = "Intel-GAUDI"
	AmdInstinctMi100     = "AMD-Instinct-MI100"
	AmdInstinctMi250x    = "AMD-Instinct-MI250X"
	AmdInstinctMi250     = "AMD-Instinct-MI250X-MI250"
	AmdInstinctMi210     = "AMD-Instinct-MI210"
	AmdInstinctMi300x    = "AMD-Instinct-MI300X-OAM"
	AmdRadeonR9200Hd7900 = "AMD-Radeon-R9-200-HD-7900"
	AmdRadeonHd7900      = "AMD-Radeon-HD-7900"
	AwsNeuronCore        = "aws-neuron-core"
	GoogleTpuV2          = "TPU-V2"
	GoogleTpuV3          = "TPU-V3"
	GoogleTpuV4          = "TPU-V4"
	NvidiaA10040g        = "A100-40G"
	NvidiaA10080g        = "A100-80G"
)

// GetAcceleratorTypeByProductName returns the accelerator type by label if the value contains the accelerator type.
func GetAcceleratorTypeByProductName(gpuProductName string) string {
	if strings.Contains(gpuProductName, NvidiaTeslaV100) {
		return NvidiaTeslaV100
	} else if strings.Contains(gpuProductName, NvidiaTeslaP100) {
		return NvidiaTeslaP100
	} else if strings.Contains(gpuProductName, NvidiaTeslaT4) {
		return NvidiaTeslaT4
	} else if strings.Contains(gpuProductName, NvidiaTeslaP4) {
		return NvidiaTeslaP4
	} else if strings.Contains(gpuProductName, NvidiaTeslaK80) {
		return NvidiaTeslaK80
	} else if strings.Contains(gpuProductName, NvidiaTeslaA10g) {
		return NvidiaTeslaA10g
	} else if strings.Contains(gpuProductName, NvidiaL4) {
		return NvidiaL4
	} else if strings.Contains(gpuProductName, NvidiaA100) {
		return NvidiaA100
	} else if strings.Contains(gpuProductName, IntelMax1550) {
		return IntelMax1550
	} else if strings.Contains(gpuProductName, IntelMax1100) {
		return IntelMax1100
	} else if strings.Contains(gpuProductName, IntelGaudi) {
		return IntelGaudi
	} else if strings.Contains(gpuProductName, AmdInstinctMi100) {
		return AmdInstinctMi100
	} else if strings.Contains(gpuProductName, AmdInstinctMi250x) {
		return AmdInstinctMi250x
	} else if strings.Contains(gpuProductName, AmdInstinctMi250) {
		return AmdInstinctMi250
	} else if strings.Contains(gpuProductName, AmdInstinctMi210) {
		return AmdInstinctMi210
	} else if strings.Contains(gpuProductName, AmdInstinctMi300x) {
		return AmdInstinctMi300x
	} else if strings.Contains(gpuProductName, AmdRadeonR9200Hd7900) {
		return AmdRadeonR9200Hd7900
	} else if strings.Contains(gpuProductName, AmdRadeonHd7900) {
		return AmdRadeonHd7900
	} else if strings.Contains(gpuProductName, AwsNeuronCore) {
		return AwsNeuronCore
	} else if strings.Contains(gpuProductName, GoogleTpuV2) {
		return GoogleTpuV2
	} else if strings.Contains(gpuProductName, GoogleTpuV3) {
		return GoogleTpuV3
	} else if strings.Contains(gpuProductName, GoogleTpuV4) {
		return GoogleTpuV4
	} else if strings.Contains(gpuProductName, NvidiaA10040g) {
		return NvidiaA10040g
	} else if strings.Contains(gpuProductName, NvidiaA10080g) {
		return NvidiaA10080g
	}

	return ""
}
