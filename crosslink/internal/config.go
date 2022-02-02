package crosslink

import (
	"log"

	"go.uber.org/zap"
)

type moduleInfo struct {
	moduleFilePath            string
	moduleContents            []byte
	requiredReplaceStatements map[string]struct{}
}

func newModuleInfo() *moduleInfo {
	var mi moduleInfo
	mi.requiredReplaceStatements = make(map[string]struct{})
	return &mi
}

type runConfig struct {
	RootPath string
	Verbose  bool
	// TODO: callout excluded path should be original module name not replaced module name. aka go.opentelemetry.io not ../replace
	ExcludedPaths map[string]struct{}
	Overwrite     bool
	Prune         bool
	logger        *zap.Logger
}

func DefaultRunConfig() runConfig {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Printf("Could not create zap logger: %v", err)
	}
	ep := make(map[string]struct{})
	rc := runConfig{
		logger:        lg,
		ExcludedPaths: ep,
	}
	return rc
}
