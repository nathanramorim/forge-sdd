package pipeline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
)

// Estados de uma task na junção do leque.
const (
	JuncaoIntegrada = "integrada"
	JuncaoBloqueada = "bloqueada"
)

const (
	// chaveLockPrincipal é a chave do lock (o mesmo de harness.AdquirirLock usado pelo laço) que
	// protege o worktree principal enquanto um leque está em curso.
	chaveLockPrincipal = "worktree-principal"
	leaseLeque         = 15 * time.Minute
	maxTasksLeque      = 50
	maxArquivosConf    = 20
)

// ParamsLeque configura um leque: um worktree isolado por task, partindo de Base.
type ParamsLeque struct {
	Iniciativa string
	Dono       string
	Base       string   // branch da iniciativa; precisa ser a que está em checkout no worktree principal
	Tasks      []string // na ordem do plano; a junção respeita essa ordem
	Agora      func() time.Time
}

// Leque é um lote de tasks com worktree próprio, sob o lock do worktree principal.
type Leque struct {
	dir, iniciativa, base string
	tasks                 []string
	caminhos              map[string]string
	lock                  *harness.Lock
	agora                 func() time.Time
}

// ResultadoJuncao é o desfecho de uma task na junção.
type ResultadoJuncao struct {
	Task     string   `json:"task"`
	Estado   string   `json:"estado"`
	Motivo   string   `json:"motivo,omitempty"`
	Arquivos []string `json:"arquivos,omitempty"` // arquivos em conflito
}

// RetornoJuncao é o que a junção entrega ao chat: o resultado por task e, havendo bloqueio, o menu (RN-9).
type RetornoJuncao struct {
	Resultados  []ResultadoJuncao `json:"resultados"`
	Opcoes      []Opcao           `json:"opcoes,omitempty"`
	Recomendada string            `json:"recomendada,omitempty"`
}

// Bloqueadas diz se alguma task ficou bloqueada.
func (r *RetornoJuncao) Bloqueadas() bool {
	for _, x := range r.Resultados {
		if x.Estado == JuncaoBloqueada {
			return true
		}
	}
	return false
}

func refSegura(ref string) bool {
	return refRe.MatchString(ref) && !strings.HasPrefix(ref, "-") && !strings.Contains(ref, "..") && !strings.HasSuffix(ref, ".lock") && len(ref) <= 200
}

func gitLimpo(dir string, args ...string) (string, error) {
	out, err := layout.Git(dir, args...)
	if err != nil {
		return "", errors.New(notify.Sanitize(err.Error(), 400, false))
	}
	return strings.TrimSpace(out), nil
}

// PrepararLeque recusa começar se o worktree principal está sujo, na branch errada ou ocupado por
// outra sessão (lock) e cria um worktree por task em .forge/runtime/worktrees/<iniciativa>/<task>.
// Worktrees órfãos de uma execução interrompida são reaproveitados só se não guardam trabalho;
// se guardam, a preparação para e pede decisão da pessoa. Chame Encerrar ao final.
func PrepararLeque(dir string, p ParamsLeque) (*Leque, error) {
	if p.Agora == nil {
		p.Agora = time.Now
	}
	if !layout.ValidInitiative(p.Iniciativa) || !wtIDRe.MatchString(p.Iniciativa) {
		return nil, fmt.Errorf("iniciativa %q inválida", p.Iniciativa)
	}
	if !refSegura(p.Base) {
		return nil, fmt.Errorf("branch base %q inválida", p.Base)
	}
	if len(p.Tasks) == 0 || len(p.Tasks) > maxTasksLeque {
		return nil, fmt.Errorf("leque exige de 1 a %d tasks", maxTasksLeque)
	}
	vistas := map[string]bool{}
	for _, t := range p.Tasks {
		if !wtIDRe.MatchString(t) || vistas[t] {
			return nil, fmt.Errorf("task %q inválida ou repetida", t)
		}
		vistas[t] = true
	}
	topo, err := gitLimpo(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	a, e1 := filepath.EvalSymlinks(topo)
	b, e2 := filepath.EvalSymlinks(dir)
	if e1 != nil || e2 != nil || a != b {
		return nil, fmt.Errorf("%s não é a raiz do worktree principal", notify.Sanitize(dir, 200, false))
	}
	lock, err := harness.AdquirirLock(dirLocks(dir), chaveLockPrincipal, p.Dono, leaseLeque, p.Agora())
	if err != nil {
		return nil, fmt.Errorf("worktree principal ocupado: %w", err)
	}
	lq := &Leque{dir: dir, iniciativa: p.Iniciativa, base: p.Base, tasks: p.Tasks, caminhos: map[string]string{}, lock: lock, agora: p.Agora}
	pronto := false
	defer func() {
		if !pronto { // erro ou panic: o lock nunca fica preso até a lease vencer
			_ = lock.Liberar()
		}
	}()
	fail := func(err error) (*Leque, error) {
		for _, t := range lq.tasks {
			if path, ok := lq.caminhos[t]; ok {
				_ = removerSeguro(dir, p.Iniciativa, t, path)
			}
		}
		_ = lock.Liberar()
		return nil, err
	}
	atual, err := gitLimpo(dir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return fail(err)
	}
	if atual != p.Base {
		return fail(fmt.Errorf("o worktree principal está em %q, não em %q: recusado", notify.Sanitize(atual, 100, false), p.Base))
	}
	if sujo, err := gitLimpo(dir, "status", "--porcelain", "--untracked-files=no"); err != nil {
		return fail(err)
	} else if sujo != "" {
		return fail(errors.New("o worktree principal tem alterações não commitadas: recusado"))
	}
	for _, t := range p.Tasks {
		if err := lq.prepararTask(t); err != nil {
			return fail(err)
		}
	}
	pronto = true
	return lq, nil
}

// removerSeguro só remove worktrees recém-criados nesta chamada (limpeza de falha de preparação).
// O chamador descarta o erro de propósito: é limpeza em melhor esforço e, se falhar, o que sobra é
// um órfão vazio que a próxima preparação reconhece e recria; nada com trabalho é apagado aqui.
func removerSeguro(dir, job, task, path string) error {
	if n, err := commitsAdiante(dir, "forge/"+job+"-"+task, ""); err == nil && n > 0 {
		return nil
	}
	return RemoveWorktree(dir, job, task)
}

func commitsAdiante(dir, branch, base string) (int, error) {
	if base == "" { // sem base explícita: compara com HEAD do principal
		base = "HEAD"
	}
	out, err := gitLimpo(dir, "rev-list", "--count", base+".."+branch)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(out)
}

func (l *Leque) prepararTask(task string) error {
	path, err := WorktreePath(l.dir, l.iniciativa, task)
	if err != nil {
		return err
	}
	branch := "forge/" + l.iniciativa + "-" + task
	fi, errPath := os.Lstat(path)
	if errPath == nil && fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("caminho do worktree de %q é um link simbólico: recusado", task)
	}
	existeBranch := false
	if _, err := layout.Git(l.dir, "rev-parse", "--verify", "-q", "refs/heads/"+branch); err == nil {
		existeBranch = true
	}
	if errPath == nil || existeBranch {
		// Órfão de uma execução interrompida: só descarta se provar estar vazio e sem commits fora da base.
		if l.orfaoTemTrabalho(path, branch, errPath == nil, existeBranch) {
			return &ErroOrfao{Task: task, Branch: branch, Opcoes: []Opcao{
				{ID: "retomar", Rotulo: "Retomar o trabalho", Descricao: "Continue na branch/worktree existente; nada foi tocado."},
				{ID: "preservar", Rotulo: "Salvar em branch e descartar depois", Descricao: "Crie uma branch no commit do worktree, confira, e só então remova o worktree à mão."},
				{ID: "abortar", Rotulo: "Parar o lote", Descricao: "Encerra sem alterar nada."},
			}}
		}
		if errPath == nil {
			if err := RemoveWorktree(l.dir, l.iniciativa, task); err != nil {
				return err
			}
		}
		if _, err := layout.Git(l.dir, "rev-parse", "--verify", "-q", "refs/heads/"+branch); err == nil {
			// -d (não -D): o git recusa apagar branch com commits não integrados.
			if _, err := gitLimpo(l.dir, "branch", "-d", branch); err != nil {
				return err
			}
		}
	}
	path, err = CreateWorktree(l.dir, l.iniciativa, task, l.base)
	if err != nil {
		return err
	}
	l.caminhos[task] = path
	return nil
}

// ErroOrfao indica um worktree ou branch órfão que guarda trabalho; nada é removido e a pessoa
// escolhe pelo menu (RN-9).
type ErroOrfao struct {
	Task, Branch string
	Opcoes       []Opcao
}

func (e *ErroOrfao) Error() string {
	return fmt.Sprintf("worktree órfão de %q guarda trabalho (branch %s): nada foi removido; decida entre retomar ou preservar", e.Task, e.Branch)
}

// orfaoTemTrabalho é conservador: qualquer dúvida ou erro conta como trabalho.
func (l *Leque) orfaoTemTrabalho(path, branch string, temPath, temBranch bool) bool {
	if temBranch {
		if n, err := commitsAdiante(l.dir, branch, l.base); err != nil || n > 0 {
			return true
		}
	}
	if !temPath {
		return false
	}
	if sujo, err := gitLimpo(path, "status", "--porcelain"); err != nil || sujo != "" { // inclui não rastreados
		return true
	}
	if head, err := gitLimpo(path, "symbolic-ref", "-q", "HEAD"); err != nil || head != "refs/heads/"+branch {
		return true // HEAD destacado ou em outra branch
	}
	out, err := gitLimpo(path, "rev-list", "--count", "HEAD", "^"+l.base)
	if n, cerr := strconv.Atoi(out); err != nil || cerr != nil || n > 0 {
		return true
	}
	return false
}

// Caminho devolve o worktree da task.
func (l *Leque) Caminho(task string) (string, bool) { p, ok := l.caminhos[task]; return p, ok }

// Juntar integra as branches das tasks na branch da iniciativa, na ordem do plano, com
// `git merge --no-ff`. Conflito nunca é resolvido aqui: o merge é abortado (sem perder nada, a
// branch e o worktree da task ficam) e a task fica bloqueada, com menu para a pessoa decidir.
func (l *Leque) Juntar() (*RetornoJuncao, error) {
	if l == nil || l.lock == nil {
		return nil, errors.New("leque não preparado")
	}
	ret := &RetornoJuncao{}
	for _, t := range l.tasks {
		path, ok := l.caminhos[t]
		if !ok {
			continue
		}
		if err := l.lock.Renovar(l.agora()); err != nil {
			return ret, err
		}
		r, err := l.juntarTask(t, path)
		if err != nil {
			return ret, err // resultados parciais: o que já foi integrado não pode se perder do relato
		}
		ret.Resultados = append(ret.Resultados, r)
	}
	if ret.Bloqueadas() {
		ret.Opcoes = []Opcao{
			{ID: "resolver", Rotulo: "Resolver o conflito", Descricao: "Resolva na branch da task (merge da base nela) e rode a junção de novo; nada foi descartado."},
			{ID: "pular", Rotulo: "Seguir sem a task", Descricao: "Mantém a task bloqueada com a branch preservada e conclui o resto."},
			{ID: "abortar", Rotulo: "Parar o lote", Descricao: "Encerra sem mais junções; branches e worktrees ficam intactos."},
		}
		ret.Recomendada = "resolver"
	}
	return ret, nil
}

func (l *Leque) juntarTask(task, path string) (ResultadoJuncao, error) {
	r := ResultadoJuncao{Task: task}
	branch := "forge/" + l.iniciativa + "-" + task
	bloquear := func(motivo string) (ResultadoJuncao, error) {
		r.Estado, r.Motivo = JuncaoBloqueada, motivo
		return r, nil
	}
	if sujo, err := gitLimpo(path, "status", "--porcelain"); err != nil {
		return bloquear(err.Error())
	} else if sujo != "" {
		return bloquear("worktree da task tem alterações não commitadas; nada foi integrado")
	}
	if atual, err := gitLimpo(l.dir, "symbolic-ref", "--short", "HEAD"); err != nil || atual != l.base {
		return r, errors.New("o worktree principal saiu da branch base durante a junção")
	}
	if sujo, err := gitLimpo(l.dir, "status", "--porcelain", "--untracked-files=no"); err != nil || sujo != "" {
		return r, errors.New("o worktree principal ficou sujo durante a junção")
	}
	args := []string{"merge", "--no-ff", "--no-edit", "-m", "merge(" + l.iniciativa + "): " + task, branch}
	if !l.temIdentidade() {
		// Só sem identidade resolvível (CI, contêiner): a do usuário, quando existe, nunca é trocada
		// e nada é gravado na config do repositório.
		args = append([]string{"-c", "user.name=forge-sdd", "-c", "user.email=forge-sdd@localhost"}, args...)
	}
	_, err := layout.Git(l.dir, args...)
	if err != nil {
		// Erro descartado de propósito: a lista de arquivos só enriquece o relato; o abort abaixo e o
		// bloqueio da task acontecem de qualquer forma.
		arqs, _ := layout.Git(l.dir, "diff", "--name-only", "--diff-filter=U")
		for _, a := range strings.Split(strings.TrimSpace(arqs), "\n") {
			if a != "" && len(r.Arquivos) < maxArquivosConf {
				r.Arquivos = append(r.Arquivos, notify.Sanitize(a, 200, false))
			}
		}
		if _, merr := layout.Git(l.dir, "rev-parse", "-q", "--verify", "MERGE_HEAD"); merr == nil { // só com merge em andamento
			if _, aerr := layout.Git(l.dir, "merge", "--abort"); aerr != nil {
				return r, fmt.Errorf("merge de %s falhou e não foi possível abortar: resolva manualmente", task)
			}
		}
		if len(r.Arquivos) > 0 {
			return bloquear("conflito de junção; merge abortado, branch " + branch + " preservada")
		}
		msg := err.Error()
		if i := strings.IndexByte(msg, '\n'); i >= 0 {
			msg = msg[:i]
		}
		return bloquear("falhou (não é conflito): " + notify.Sanitize(msg, 300, false))
	}
	r.Estado = JuncaoIntegrada
	if err := RemoveWorktree(l.dir, l.iniciativa, task); err != nil {
		r.Motivo = "integrada; limpeza do worktree falhou: " + notify.Sanitize(err.Error(), 200, false)
	}
	delete(l.caminhos, task)
	return r, nil
}

// temIdentidade diz se o git resolve autor e e-mail para este repositório.
func (l *Leque) temIdentidade() bool {
	n, e1 := gitLimpo(l.dir, "config", "--get", "user.name")
	m, e2 := gitLimpo(l.dir, "config", "--get", "user.email")
	return e1 == nil && e2 == nil && n != "" && m != ""
}

// Encerrar libera o lock do worktree principal. Worktrees e branches de tasks bloqueadas ficam.
func (l *Leque) Encerrar() error {
	if l == nil || l.lock == nil {
		return nil
	}
	return l.lock.Liberar()
}
