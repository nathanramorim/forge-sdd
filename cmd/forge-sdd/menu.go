package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// menuOption é uma linha do menu.
type menuOption = survey.Option

// errBack é o Esc do menu: sobe um nível (no menu principal, sai).
var errBack = survey.ErrBack

// prompter isola a interação para os testes simularem a entrada.
type prompter interface {
	Select(title string, opts []menuOption, def string) (string, error)
	// SelectGesture é o Select que também informa →/← (survey.GestureRight/GestureLeft).
	SelectGesture(title string, opts []menuOption, def string) (string, string, error)
	Confirm(title string, def bool) (bool, error)
	Pause(msg string) error
}

type huhPrompter struct{}

func (huhPrompter) Select(title string, opts []menuOption, def string) (string, error) {
	return survey.SelectOne(title, opts, def)
}
func (huhPrompter) SelectGesture(title string, opts []menuOption, def string) (string, string, error) {
	return survey.SelectGesture(title, opts, def)
}
func (huhPrompter) Confirm(title string, def bool) (bool, error) {
	return survey.ConfirmYesNo(title, def)
}
func (huhPrompter) Pause(msg string) error { return survey.Pause(msg) }

// isInteractive diz se há terminal para o menu; variável para os testes.
var isInteractive = func() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// menuEnv é tudo de que o menu precisa; injetável nos testes.
type menuEnv struct {
	dir string
	out io.Writer
	p   prompter
	// run executa o RunE do subcomando (o mesmo código da linha de comando, sem subprocesso).
	run          func(args ...string) error
	preflightEnv pipeline.PreflightEnv
}

// runSubcommand executa o RunE de um subcomando com as flags dele zeradas, sem reexecutar o binário.
// O reset (Value.Set(DefValue)) vale para flags escalares; um flag de lista (StringSlice etc.) o
// quebraria, e o teste TestRunSubcommand_NenhumFlagDeLista impede que um apareça sem ajuste.
func runSubcommand(args ...string) error {
	sub, rest, err := rootCmd.Find(args)
	if err != nil || sub == rootCmd {
		return fmt.Errorf("comando %q não encontrado", strings.Join(args, " "))
	}
	sub.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	if err := sub.ParseFlags(rest); err != nil {
		return err
	}
	if sub.RunE != nil {
		return sub.RunE(sub, sub.Flags().Args())
	}
	if sub.Run != nil {
		sub.Run(sub, sub.Flags().Args())
		return nil
	}
	return sub.Help()
}

// O RunE do rootCmd é ligado aqui (e não no literal) para evitar o ciclo de inicialização
// rootCmd → runRootMenu → runSubcommand → rootCmd.
func init() {
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runRootMenu(cmd)
		}
		return cmd.Help()
	}
}

// runRootMenu é a porta de entrada de `forge` sem argumentos.
func runRootMenu(cmd *cobra.Command) error {
	if !isInteractive() {
		fmt.Fprintln(cmd.OutOrStdout(), shortHelp())
		return nil
	}
	dir, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	return runMenu(&menuEnv{
		dir: dir, out: cmd.OutOrStdout(), p: huhPrompter{}, run: runSubcommand,
		preflightEnv: pipeline.DefaultPreflightEnv(),
	})
}

func shortHelp() string {
	return `Forge SDD ` + version + `
Sem terminal interativo o menu não abre. Comandos mais usados:
  forge init | update          instala ou atualiza a estrutura
  forge pipeline init          ativa o pipeline multi-sessão
  forge up [--discovery X --builder Y --reviewer Z]   abre as 3 sessões
  forge status | gate | doctor
  forge --help                 todos os comandos`
}

const (
	valUp       = "up"
	valLive     = "live"
	valStatus   = "status"
	valQueue    = "queue"
	valDiag     = "diag"
	valConfig   = "config"
	valSettings = "settings"
	valHelp     = "help"
	valInstall  = "install"
	valExit     = "exit"
	valBack     = "back"
)

// runMenu mostra o menu principal até a pessoa sair. O menu roda uma ação por vez e não é painel ao vivo.
func runMenu(env *menuEnv) error {
	installed := config.DetectProject(env.dir)
	for {
		proj := filepath.Base(env.dir)
		state := "não instalado"
		if installed {
			state = "pronta"
			if _, err := os.Stat(filepath.Join(env.dir, "sdd", "memory", "constitution.md")); err == nil {
				if b, err := gate.BuildBrief(env.dir, 0); err == nil {
					state = b.State
				}
			}
		}
		br, err := layout.CurrentBranch(env.dir)
		if err != nil || br == "" {
			br = "—"
		}
		header := fmt.Sprintf("Forge SDD %s — %s (%s · branch: %s)", version, proj, state, br)

		var opts []menuOption
		def := valUp
		if !installed {
			opts = append(opts, menuOption{Label: "Instalar o Forge neste projeto", Value: valInstall})
			def = valInstall
		} else if state == gate.BriefRascunho {
			opts = append(opts, menuOption{Label: "Alinhar o projeto (passo 1: /constitution no seu agente de IA)", Value: "constitution"})
			def = "constitution"
		}

		opts = append(opts,
			menuOption{Label: "Iniciar esteira (up)", Value: valUp},
			menuOption{Label: "Painel vivo da esteira (status --live)", Value: valLive},
			menuOption{Label: "Status das iniciativas e esteira", Value: valStatus},
			menuOption{Label: "Fila de tasks e inbox de notificações", Value: valQueue},
			menuOption{Label: "Diagnóstico de saúde e regras (Doctor & Gate)", Value: valDiag},
			menuOption{Label: "Configurações (agentes, terminal, modo ao vivo/headless)", Value: valConfig},
			menuOption{Label: "Settings (chaves ativado/desativado)", Value: valSettings},
			menuOption{Label: "Ajuda e comandos do CLI", Value: valHelp},
			menuOption{Label: "Sair", Value: valExit},
		)

		choice, gesture, err := env.p.SelectGesture(header, opts, def)
		if errors.Is(err, errBack) || choice == valExit {
			return nil
		}
		if err != nil {
			return err
		}
		switch choice {
		case "constitution":
			fmt.Fprintln(env.out, "\nPasso 1: Alinhar o projeto")
			fmt.Fprintln(env.out, "Abra o seu agente de IA (Claude, Gemini ou Codex) e execute:")
			fmt.Fprintln(env.out, "  /constitution")
			fmt.Fprintln(env.out, "O comando vai alinhar stack, arquitetura, módulos e regras, validando com `forge brief --check`.")
			_ = env.p.Pause("Enter para voltar ao menu")
		case valInstall:
			installed = env.installOrUpdate() == nil && config.DetectProject(env.dir)
		case valUp:
			env.startSessions(gesture)
		case valLive:
			env.act("status", "--live", "--dir", env.dir)
		case valStatus:
			env.act("status", "--dir", env.dir)
		case valQueue:
			for _, a := range [][]string{
				{"queue", "panel", "--dir", env.dir},
				{"bus", "list", "--dir", env.dir, "--role", "builder"},
				{"bus", "list", "--dir", env.dir, "--role", "reviewer"},
				{"inbox", "--dir", env.dir},
			} {
				if err := env.run(a...); err != nil {
					fmt.Fprintf(env.out, "\n✗ %v\n", err)
				}
			}
			_ = env.p.Pause("Enter para voltar ao menu")
		case valDiag:
			if err := env.diagMenu(); err != nil && !errors.Is(err, errBack) {
				return err
			}
		case valSettings:
			if err := env.settingsMenu(); err != nil && !errors.Is(err, errBack) {
				return err
			}
		case valConfig:
			if err := env.configMenu(); err != nil && !errors.Is(err, errBack) {
				return err
			}
		case valHelp:
			_ = rootCmd.Help()
			_ = env.p.Pause("Enter para voltar ao menu")
		}
	}
}

// act executa uma ação, mostra o erro sem derrubar o menu e espera Enter para a saída não sumir.
func (e *menuEnv) act(args ...string) {
	if err := e.run(args...); err != nil {
		fmt.Fprintf(e.out, "\n✗ %v\n", err)
	}
	_ = e.p.Pause("Enter para voltar ao menu")
}

func (e *menuEnv) installOrUpdate() error {
	cmd := "init"
	if config.DetectProject(e.dir) {
		cmd = "update"
	}
	err := e.run(cmd, e.dir)
	if err != nil {
		fmt.Fprintf(e.out, "\n✗ %v\n", err)
	}
	return err
}

func (e *menuEnv) diagMenu() error {
	for {
		choice, err := e.p.Select("Diagnóstico de saúde e regras", []menuOption{
			{Label: "Doctor (check-up do projeto)", Value: "doctor"},
			{Label: "Gate de governança (Camada A)", Value: "gate"},
			{Label: "Processos ativos e consumo (ps)", Value: "ps"},
			{Label: "Diagnóstico do pipeline (CLI instalada, arquivos do agente)", Value: "pipeline"},
			{Label: "Recuperar jobs presos", Value: "recover"},
			{Label: "Voltar", Value: valBack},
		}, "doctor")
		if errors.Is(err, errBack) || choice == valBack {
			return errBack
		}
		if err != nil {
			return err
		}
		switch choice {
		case "doctor":
			e.act("doctor", "--dir", e.dir)
		case "gate":
			e.act("gate", "--dir", e.dir)
		case "ps":
			e.act("ps", "--dir", e.dir)
		case "pipeline":
			e.diagnose()
		case "recover":
			for _, r := range []string{"builder", "reviewer"} {
				if err := e.run("bus", "recover", "--dir", e.dir, "--role", r); err != nil {
					fmt.Fprintf(e.out, "\n✗ %v\n", err)
				}
			}
			_ = e.p.Pause("Enter para voltar ao menu")
		}
	}
}

// loadPipeline devolve a configuração do pipeline; sem forge.json oferece ativar.
func (e *menuEnv) loadPipeline() (config.ForgeConfig, bool) {
	cfg, found, err := config.LoadForgeConfig(e.dir)
	if err != nil {
		fmt.Fprintf(e.out, "\n✗ %v\n", err)
		return cfg, false
	}
	if found {
		return cfg, true
	}
	ok, err := e.p.Confirm("O pipeline ainda não está ativo neste projeto. Ativar agora (cria .forge/config/forge.json)?", true)
	if err != nil || !ok {
		return cfg, false
	}
	if err := e.run("pipeline", "init", "--dir", e.dir); err != nil {
		fmt.Fprintf(e.out, "\n✗ %v\n", err)
		return cfg, false
	}
	cfg, _, err = config.LoadForgeConfig(e.dir)
	return cfg, err == nil
}

func roleLabel(role string) string {
	return map[string]string{
		config.SessionRoleDiscovery: "forge-discovery (conversa e decide)",
		config.SessionRoleBuilder:   "forge-builder (constrói em paralelo)",
		config.SessionRoleReviewer:  "forge-reviewer (revisa o PR)",
	}[role]
}

// chooseAgents pergunta o agente de cada sessão, com a escolha lembrada já marcada e o aviso das CLIs
// que não estão no PATH. Esc em qualquer pergunta cancela sem gravar nada.
func (e *menuEnv) chooseAgents(cfg config.ForgeConfig) (config.ForgeConfig, error) {
	sessions := map[string]string{}
	for _, role := range config.SessionRoles() {
		var opts []menuOption
		for _, cli := range config.OfferedCLIs() {
			label := cli
			profile, _ := cfg.Profile(cli)
			if e.preflightEnv.LookPath != nil {
				if _, err := e.preflightEnv.LookPath(profile.Command); err != nil {
					label += "  (não encontrado no PATH)"
				}
			}
			opts = append(opts, menuOption{Label: label, Value: cli})
		}
		cur := cfg.CLIFor(role)
		if !contains(config.OfferedCLIs(), cur) {
			cur = config.CLIClaude
		}
		choice, err := e.p.Select("Agente da sessão "+roleLabel(role), opts, cur)
		if err != nil {
			return cfg, err
		}
		sessions[role] = choice
	}
	cfg.Sessions = sessions
	return cfg, nil
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func (e *menuEnv) choosePermissions(cfg config.ForgeConfig) (config.ForgeConfig, error) {
	choice, err := e.p.Select("Permissões do Builder e do Reviewer (rodam sem você aprovar cada passo)", []menuOption{
		{Label: config.PermissionSummary(config.PermissionProject, cfg.AllowCommands), Value: config.PermissionProject},
		{Label: config.PermissionSummary(config.PermissionFull, cfg.AllowCommands), Value: config.PermissionFull},
	}, cfg.Permissions)
	if err != nil {
		return cfg, err
	}
	cfg.Permissions = choice
	return cfg, nil
}

func summary(cfg config.ForgeConfig) string {
	var parts []string
	for _, role := range config.SessionRoles() {
		parts = append(parts, role+" "+cfg.CLIFor(role))
	}
	return strings.Join(parts, " · ")
}

// applyAutonomyGesture liga (→) ou desliga (←) o modo avançado sobre cfg. Ligar exige a confirmação
// de ambiente isolado; sem ela devolve ok=false e nada é gravado (ADR-0013).
func (e *menuEnv) applyAutonomyGesture(cfg config.ForgeConfig, gesture string) (config.ForgeConfig, bool) {
	switch gesture {
	case survey.GestureRight:
		if cfg.IsAdvanced() {
			return cfg, true
		}
		ok, err := e.p.Confirm("Modo avançado (experimental): o Reviewer aprova a spec e as sessões rodam sem pedir permissão. O ambiente é isolado (contêiner, VM ou máquina descartável)?", false)
		if err != nil || !ok {
			fmt.Fprintln(e.out, "\nModo avançado não ligado: sem a confirmação de ambiente isolado nada foi gravado.")
			return cfg, false
		}
		cfg.Autonomy, cfg.IsolatedConfirmed, cfg.Permissions = config.AutonomyAdvanced, true, config.PermissionFull
	case survey.GestureLeft:
		cfg.Autonomy, cfg.IsolatedConfirmed = config.AutonomyNormal, false
	}
	return cfg, true
}

// startSessions é o fluxo principal: escolher, checar, confirmar, gravar e abrir. gesture vem do
// menu principal (→ liga e ← desliga o modo avançado).
func (e *menuEnv) startSessions(gesture string) {
	cfg, ok := e.loadPipeline()
	if !ok {
		_ = e.p.Pause("Enter para voltar ao menu")
		return
	}
	cfg, ok = e.applyAutonomyGesture(cfg, gesture)
	if !ok {
		_ = e.p.Pause("Enter para voltar ao menu")
		return
	}
	cfg, err := e.chooseAgents(cfg)
	if err != nil {
		return
	}
	if !cfg.IsAdvanced() { // o avançado força permissions full
		cfg, err = e.choosePermissions(cfg)
		if err != nil {
			return
		}
	}
	items := pipeline.Preflight(e.dir, cfg, e.preflightEnv)
	fmt.Fprintf(e.out, "\nChecagem antes de abrir:\n%s\n", pipeline.FormatPreflight(items))
	if pipeline.PreflightBlocked(items) {
		fmt.Fprintln(e.out, "✗ Há bloqueios: nada foi aberto nem gravado. Resolva e tente de novo (ou escolha outro agente).")
		_ = e.p.Pause("Enter para voltar ao menu")
		return
	}
	for _, it := range items {
		if it.Level == pipeline.LevelWarn {
			ok, err := e.p.Confirm("Há avisos acima. Abrir mesmo assim?", false)
			if err != nil || !ok {
				return
			}
			break
		}
	}
	mode := ""
	if cfg.IsAdvanced() {
		mode = " · avançado"
	}
	ok, err = e.p.Confirm("Abrir agora? "+summary(cfg)+" · permissões "+cfg.Permissions+mode+" · "+e.terminalLabel(cfg), true)
	if err != nil || !ok {
		return
	}
	if err := config.SaveForgeConfig(e.dir, cfg); err != nil {
		fmt.Fprintf(e.out, "\n✗ %v\n", err)
		_ = e.p.Pause("Enter para voltar ao menu")
		return
	}
	e.act("up", "--dir", e.dir)
}

func (e *menuEnv) configMenu() error {
	for {
		choice, err := e.p.Select("Ativar ou reconfigurar o pipeline", []menuOption{
			{Label: "Ativar o pipeline (cria .forge/config/forge.json)", Value: "init"},
			{Label: "Escolher o agente de cada sessão (sem abrir)", Value: "agents"},
			{Label: "Permissões do Builder e do Reviewer", Value: "perms"},
			{Label: "Modo de sessão (ao vivo no painel ou worker headless)", Value: "mode"},
			{Label: "Contexto entre jobs (limpar ou contínuo)", Value: "context"},
			{Label: "Voltar", Value: valBack},
		}, "agents")
		if errors.Is(err, errBack) || choice == valBack {
			return errBack
		}
		if err != nil {
			return err
		}
		if choice == "init" {
			e.act("pipeline", "init", "--dir", e.dir)
			continue
		}
		cfg, ok := e.loadPipeline()
		if !ok {
			continue
		}
		switch choice {
		case "agents":
			cfg, err = e.chooseAgents(cfg)
		case "perms":
			cfg, err = e.choosePermissions(cfg)
		case "mode":
			var mode string
			mode, err = e.p.Select("Modo de sessão do Builder e Reviewer", []menuOption{
				{Label: "Ao vivo (CLI aberta no painel desde o início; requer tmux)", Value: config.SessionModeLive},
				{Label: "Headless (worker aguarda jobs e roda o agente em segundo plano)", Value: config.SessionModeHeadless},
			}, cfg.SessionMode)
			if err == nil {
				cfg.SessionMode = mode
			}
		case "context":
			var ctx string
			ctx, err = e.p.Select("Contexto entre jobs", []menuOption{
				{Label: "Limpar contexto a cada job (recomendado)", Value: config.ContextClear},
				{Label: "Contínuo (acumula contexto; maior consumo de tokens)", Value: config.ContextContinuous},
			}, cfg.Context)
			if err == nil {
				cfg.Context = ctx
			}
		}
		if err != nil {
			continue
		}
		if err := config.SaveForgeConfig(e.dir, cfg); err != nil {
			fmt.Fprintf(e.out, "\n✗ %v\n", err)
		} else {
			fmt.Fprintln(e.out, "\n✓ Gravado em .forge/config/forge.json")
		}
	}
}

func (e *menuEnv) diagnose() {
	cfg, ok := e.loadPipeline()
	if ok {
		fmt.Fprintf(e.out, "\nAgentes: %s · permissões %s\n%s\n", summary(cfg), cfg.Permissions, pipeline.FormatPreflight(pipeline.Preflight(e.dir, cfg, e.preflightEnv)))
	}
	_ = e.p.Pause("Enter para voltar ao menu")
}

// terminalLabel diz onde as sessões vão abrir: "no tmux (3 painéis)", "em painéis do Windows Terminal" ou "em janelas do sistema".
func (e *menuEnv) terminalLabel(cfg config.ForgeConfig) string {
	look := e.preflightEnv.LookPath
	if look == nil {
		look = pipeline.DefaultPreflightEnv().LookPath
	}
	env := pipeline.LaunchEnv{GOOS: runtime.GOOS, LookPath: look}
	term, err := pipeline.ResolveTerminal("", cfg, env)
	if err != nil || term == config.TerminalNative {
		if runtime.GOOS == "windows" {
			if _, err := look("wt.exe"); err == nil {
				return "em painéis do Windows Terminal"
			}
			if _, err := look("wt"); err == nil {
				return "em painéis do Windows Terminal"
			}
		}
		return "em janelas do sistema"
	}
	return "no tmux (3 painéis)"
}

// settingKey é uma chave liga/desliga mostrada no menu Settings.
type settingKey struct {
	value string
	label string
	get   func(config.ForgeConfig) bool
	set   func(*config.ForgeConfig, bool)
}

// settingKeys é a casa única das chaves liga/desliga (DUP-01); o card e o teto diário de jobs.
var settingKeys = []settingKey{{
	value: "card",
	label: "Usa gerenciador de projeto (card obrigatório no /implementar)",
	get:   func(c config.ForgeConfig) bool { return c.PMMCP.RequireExistingCard },
	set:   func(c *config.ForgeConfig, v bool) { c.PMMCP.RequireExistingCard = v },
}, {
	value: "teto-diario",
	label: fmt.Sprintf("Teto diário de jobs (%d/dia; no teto a esteira pausa até o dia seguinte)", defaultDailyJobCap),
	get:   func(c config.ForgeConfig) bool { return c.Safeguards.MaxJobsPerDay > 0 },
	set: func(c *config.ForgeConfig, v bool) {
		c.Safeguards.MaxJobsPerDay = 0
		if v {
			c.Safeguards.MaxJobsPerDay = defaultDailyJobCap
		}
	},
}}

// defaultDailyJobCap é o valor gravado ao ligar o teto diário no Settings.
const defaultDailyJobCap = 30

func onOff(v bool) string {
	if v {
		return "ativado"
	}
	return "desativado"
}

// settingsMenu lista as chaves como ativado/desativado e alterna a escolhida, gravando com SaveForgeConfig.
func (e *menuEnv) settingsMenu() error {
	for {
		cfg, ok := e.loadPipeline()
		if !ok {
			return errBack
		}
		opts := make([]menuOption, 0, len(settingKeys)+1)
		for _, k := range settingKeys {
			opts = append(opts, menuOption{Label: fmt.Sprintf("%s: %s", k.label, onOff(k.get(cfg))), Value: k.value})
		}
		opts = append(opts, menuOption{Label: "Voltar", Value: valBack})
		choice, err := e.p.Select("Settings (Enter alterna a chave)", opts, settingKeys[0].value)
		if errors.Is(err, errBack) || choice == valBack {
			return errBack
		}
		if err != nil {
			return err
		}
		for _, k := range settingKeys {
			if k.value != choice {
				continue
			}
			k.set(&cfg, !k.get(cfg))
			if err := config.SaveForgeConfig(e.dir, cfg); err != nil {
				fmt.Fprintf(e.out, "\n✗ %v\n", err)
			} else {
				fmt.Fprintf(e.out, "\n✓ %s: %s (gravado em .forge/config/forge.json)\n", k.label, onOff(k.get(cfg)))
			}
		}
	}
}
