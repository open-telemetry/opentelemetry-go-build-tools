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

type runConfig struct {
	rootPath string
	verbose  bool
	// TODO: callout excluded path should be original module name not replaced module name. aka go.opentelemetry.io not ../replace
	excludedPaths map[string]struct{}
	overwrite     bool
	prune         bool
	logger        *zap.Logger
}

func newModuleInfo() *moduleInfo {
	var mi moduleInfo
	mi.requiredReplaceStatements = make(map[string]struct{})
	return &mi
}

func DefaultRunConfig() runConfig {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Printf("Could not create zap logger: %v", err)
	}
	ep := make(map[string]struct{})
	rc := runConfig{
		logger:        lg,
		excludedPaths: ep,
	}
	return rc
}
