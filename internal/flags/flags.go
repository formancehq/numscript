package flags

type FeatureFlag = string

const (
	ExperimentalOverdraftFunctionFeatureFlag FeatureFlag = "experimental-overdraft-function"
	ExperimentalScopedFunction               FeatureFlag = "experimental-scoped-function"
	ExperimentalOneofFeatureFlag             FeatureFlag = "experimental-oneof"
	ExperimentalAccountInterpolationFlag     FeatureFlag = "experimental-account-interpolation"
	AssetScaling                             FeatureFlag = "experimental-asset-scaling"
)

var AllFlags []string = []string{
	ExperimentalOverdraftFunctionFeatureFlag,
	ExperimentalScopedFunction,
	ExperimentalOneofFeatureFlag,
	ExperimentalAccountInterpolationFlag,
	AssetScaling,
}
