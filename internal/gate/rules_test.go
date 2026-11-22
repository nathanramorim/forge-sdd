package gate

import (
	"reflect"
	"testing"
)

func TestSelecionaIDs(t *testing.T) {
	catalogo := map[string]string{"ARQ-01": "a", "FLX-08": "b", "GOV-04": "c"}

	ids, ausentes, err := SelecionaIDs(catalogo, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []string{"ARQ-01", "FLX-08", "GOV-04"}) || ausentes != nil {
		t.Fatalf("sem pedidos: ids=%v ausentes=%v", ids, ausentes)
	}

	ids, ausentes, err = SelecionaIDs(catalogo, []string{"gov-04", "xyz-99", "arq-01"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []string{"ARQ-01", "GOV-04"}) {
		t.Fatalf("ids = %v", ids)
	}
	if !reflect.DeepEqual(ausentes, []string{"XYZ-99"}) {
		t.Fatalf("ausentes = %v", ausentes)
	}

	if _, _, err := SelecionaIDs(catalogo, []string{"xyz-99"}); err == nil {
		t.Fatal("sem nenhum ID válido deveria falhar")
	}
}
