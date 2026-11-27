package eol

import "testing"

func TestLF(t *testing.T) {
	casos := map[string]string{
		"":               "",
		"a\nb\n":         "a\nb\n",
		"a\r\nb\r\n":     "a\nb\n",
		"misto\r\nlf\n":  "misto\nlf\n",
		"cr solto\rfica": "cr solto\rfica",
	}
	for in, want := range casos {
		if got := string(LF([]byte(in))); got != want {
			t.Errorf("LF(%q) = %q, quer %q", in, got, want)
		}
	}
}

func TestIguais(t *testing.T) {
	if !Iguais([]byte("x\r\ny\r\n"), []byte("x\ny\n")) {
		t.Fatal("CRLF e LF com o mesmo texto devem ser iguais")
	}
	if Iguais([]byte("x\ny\n"), []byte("x\nz\n")) {
		t.Fatal("textos diferentes não podem ser iguais")
	}
}
