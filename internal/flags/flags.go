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
	ExperimentalVirtualAccount               FeatureFlag = "experimental-virtual-account"
)

var AllFlags []string = []string{
	ExperimentalOverdraftFunctionFeatureFlag,
	ExperimentalGetAssetFunctionFeatureFlag,
	ExperimentalGetAmountFunctionFeatureFlag,
	ExperimentalOneofFeatureFlag,
	ExperimentalAccountInterpolationFlag,
	ExperimentalMidScriptFunctionCall,
	ExperimentalAssetColors,
	ExperimentalVirtualAccount,
}
