package flags

type FeatureFlag = string

const (
	ExperimentalOverdraftFunctionFeatureFlag FeatureFlag = "experimental-overdraft-function"
	ExperimentalGetAssetFunctionFeatureFlag  FeatureFlag = "experimental-get-asset-function"
	ExperimentalGetAmountFunctionFeatureFlag FeatureFlag = "experimental-get-amount-function"
	ExperimentalOneofFeatureFlag             FeatureFlag = "experimental-oneof"
	ExperimentalAccountInterpolationFlag     FeatureFlag = "experimental-account-interpolation"
	ExperimentalMidScriptFunctionCall        FeatureFlag = "experimental-mid-script-function-call"
	ExperimentalAssetColors                  FeatureFlag = "experimental-asset-colors"
	ExperimentalMinOfFunctionFeatureFlag     FeatureFlag = "experimental-min-of-function"
	ExperimentalMaxOfFunctionFeatureFlag     FeatureFlag = "experimental-max-of-function"
	ExperimentalMultiplyFunctionFeatureFlag  FeatureFlag = "experimental-multiply-of-function"
)

var AllFlags []string = []string{
	ExperimentalOverdraftFunctionFeatureFlag,
	ExperimentalGetAssetFunctionFeatureFlag,
	ExperimentalGetAmountFunctionFeatureFlag,
	ExperimentalOneofFeatureFlag,
	ExperimentalAccountInterpolationFlag,
	ExperimentalMidScriptFunctionCall,
	ExperimentalAssetColors,
	ExperimentalMinOfFunctionFeatureFlag,
	ExperimentalMaxOfFunctionFeatureFlag,
	ExperimentalMultiplyFunctionFeatureFlag,
}
