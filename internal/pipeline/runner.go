package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
)

var (
	agentNameRe   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)
	sessionNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,60}$`)
)

// RunSpec descreve uma execução não interativa do agente para um job.
type RunSpec struct {
	Agent   string // nome do agente (builder, reviewer)
	Name    string // nome da sessão (forge-builder)
	Prompt  string
	Dir     string
	Timeout time.Duration
}

// RunResult é o desfecho da execução.
type RunResult struct {
	ExitCode int
	TimedOut bool
}

// Runner executa o agente. A implementação real usa os/exec; os testes usam um executor falso.
type Runner interface {
	Run(ctx context.Context, spec RunSpec, out io.Writer) (RunResult, error)
}

// BuildArgv monta o comando e os argumentos do perfil substituindo {agent}, {name} e {prompt}.
// Nunca passa por shell (SEG-03). O prompt não pode virar opção da CLI: se o argumento resultante
// for exatamente o prompt, ele não pode começar com "-".
func BuildArgv(p config.CLIProfile, spec RunSpec) (string, []string, error) {
	if !agentNameRe.MatchString(spec.Agent) {
		return "", nil, fmt.Errorf("nome de agente %q inválido", spec.Agent)
	}
	if !sessionNameRe.MatchString(spec.Name) {
		return "", nil, fmt.Errorf("nome de sessão %q inválido", spec.Name)
	}
	if strings.ContainsRune(spec.Prompt, 0) {
		return "", nil, errors.New("prompt com byte nulo")
	}
	if strings.HasPrefix(strings.TrimSpace(spec.Prompt), "-") {
		return "", nil, errors.New("prompt não pode começar com '-' (poderia ser lido como opção da CLI)")
	}
	repl := strings.NewReplacer("{agent}", spec.Agent, "{name}", spec.Name, "{prompt}", spec.Prompt)
	args := make([]string, 0, len(p.Args))
	for _, a := range p.Args {
		args = append(args, repl.Replace(a))
	}
	return p.Command, args, nil
}

// ExecRunner executa o agente como processo filho.
type ExecRunner struct {
	Profile config.CLIProfile
}

// Run executa o comando do perfil com timeout, transmitindo a saída para out.
func (r ExecRunner) Run(ctx context.Context, spec RunSpec, out io.Writer) (RunResult, error) {
	command, args, err := BuildArgv(r.Profile, spec)
	if err != nil {
		return RunResult{}, err
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return RunResult{}, fmt.Errorf("CLI %q não encontrada no PATH (instale-a ou ajuste cli_profiles em forge.json)", command)
	}
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = spec.Dir
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Env = os.Environ()
	cmd.WaitDelay = 10 * time.Second
	err = cmd.Run()
	res := RunResult{}
	if ctx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
		res.ExitCode = -1
		return res, nil
	}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			res.ExitCode = ee.ExitCode()
			return res, nil
		}
		return res, fmt.Errorf("executar %s: %w", command, err)
	}
	return res, nil
}
