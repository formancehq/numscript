package flags

type FeatureFlag = string

const (
	ExperimentalOverdraftFunctionFeatureFlag FeatureFlag = "experimental-overdraft-function"
	ExperimentalGetAssetFunctionFeatureFlag  FeatureFlag = "experimental-get-asset-function"
	ExperimentalOneofFeatureFlag             FeatureFlag = "experimental-oneof"
	ExperimentalAccountInterpolationFlag     FeatureFlag = "experimental-account-interpolation"
	ExperimentalMidScriptFunctionCall        FeatureFlag = "experimental-mid-script-function-call"
)

var AllFlags []string = []string{
	ExperimentalOverdraftFunctionFeatureFlag,
	ExperimentalGetAssetFunctionFeatureFlag,
	ExperimentalOneofFeatureFlag,
	ExperimentalAccountInterpolationFlag,
	ExperimentalMidScriptFunctionCall,
}
