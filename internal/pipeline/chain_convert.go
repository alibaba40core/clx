package pipeline

import (
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/providers"
)

func buildGeneratedFromAI(resp *providers.CommandResponse, shellHint string, profile environment.SystemProfile) (generator.GeneratedCommand, error) {
	if resp.HasChain() {
		return buildFromChain(resp.Chain, shellHint, resp.Explanation, profile)
	}
	if chain := providers.ChainFromArgv(resp.Argv); chain != nil {
		return buildFromChain(chain, shellHint, resp.Explanation, profile)
	}
	if vErr := executor.ValidateGeneratedArgv(resp.Argv, shellHint); vErr != nil {
		return generator.GeneratedCommand{}, vErr
	}
	return generator.NewAICommand(resp.Argv, resp.Shell, resp.Explanation, profile), nil
}

func buildFromChain(chain *generator.CommandChain, shell, explanation string, profile environment.SystemProfile) (generator.GeneratedCommand, error) {
	if shell == "" {
		shell = profile.Shell
	}
	if vErr := executor.ValidateCommandChain(chain, shell); vErr != nil {
		return generator.GeneratedCommand{}, vErr
	}
	script, err := executor.BuildValidatedChainScript(shell, chain, profile)
	if err != nil {
		return generator.GeneratedCommand{}, err
	}
	return generator.NewCommandFromChain(chain, shell, explanation, script, profile), nil
}
