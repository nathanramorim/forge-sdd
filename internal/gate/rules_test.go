package gate

import (
	"reflect"
	"testing"
)

func TestSelecionaIDs(t *testing.T) {
	catalogo := map[string]string{"ARQ-01": "a", "FLX-08": "b", "GOV-04": "c"}

	ids, ausentes := SelecionaIDs(catalogo, nil)
	if !reflect.DeepEqual(ids, []string{"ARQ-01", "FLX-08", "GOV-04"}) || ausentes != nil {
		t.Fatalf("sem pedidos: ids=%v ausentes=%v", ids, ausentes)
	}

	ids, ausentes = SelecionaIDs(catalogo, []string{"gov-04", "xyz-99", "arq-01"})
	if !reflect.DeepEqual(ids, []string{"ARQ-01", "GOV-04"}) {
		t.Fatalf("ids = %v", ids)
	}
	if !reflect.DeepEqual(ausentes, []string{"XYZ-99"}) {
		t.Fatalf("ausentes = %v", ausentes)
	}
}
