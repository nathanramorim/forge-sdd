# Catálogo de Conceitos e Contratos (`DUP-01`) — demo

Regra `DUP-01`: **Um conceito = uma casa.** Antes de criar uma nova entidade, módulo ou utilitário, consulte esta tabela para identificar o módulo dono do conceito e reutilizar ou estender seu contrato público (`DUP-02`).

| Conceito | Pacote/Domínio Dono | Entidade/Struct Mestra | Contrato Público | Observação |
| --- | --- | --- | --- | --- |
| [Ex.: Configuração] | [Ex.: `internal/config`] | [Ex.: `Config`] | [Ex.: `Load()`] | [Fonte única de configuração do sistema] |
