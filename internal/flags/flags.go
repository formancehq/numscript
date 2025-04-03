package flags

type FeatureFlag = string

const (
	ExperimentalOverdraftFunctionFeatureFlag FeatureFlag = "experimental-overdraft-function"
	ExperimentalOneofFeatureFlag             FeatureFlag = "experimental-oneof"
	ExperimentalAccountInterpolationFlag     FeatureFlag = "experimental-account-interpolation"
	ExperimentalMidScriptFunctionCall        FeatureFlag = "experimental-mid-script-function-call"
)

var AllFlags []string = []string{
	ExperimentalOverdraftFunctionFeatureFlag,
	ExperimentalOneofFeatureFlag,
	ExperimentalAccountInterpolationFlag,
	ExperimentalMidScriptFunctionCall,
}
