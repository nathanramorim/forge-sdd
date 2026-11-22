package gate

import (
	"sort"
	"strings"
)

// SelecionaIDs escolhe, no catálogo da Constituição, os IDs pedidos (ou todos,
// quando nenhum é pedido). Devolve os IDs válidos em ordem e os pedidos que não
// existem na Constituição, em maiúsculas.
func SelecionaIDs(catalogo map[string]string, pedidos []string) (ids []string, ausentes []string) {
	if len(pedidos) == 0 {
		for id := range catalogo {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		return ids, nil
	}
	for _, pedido := range pedidos {
		upper := strings.ToUpper(pedido)
		if _, ok := catalogo[upper]; ok {
			ids = append(ids, upper)
		} else {
			ausentes = append(ausentes, upper)
		}
	}
	sort.Strings(ids)
	return ids, ausentes
}
