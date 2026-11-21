import contextlib
import io
import tempfile
import unittest
from pathlib import Path
import sys

RAIZ_SCRIPTS = Path(__file__).resolve().parent.parent
if str(RAIZ_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(RAIZ_SCRIPTS))

import licoes  # noqa: E402


class TestLicoes(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.raiz = Path(self.tmp.name)
        mem = self.raiz / "sdd" / "memory"
        mem.mkdir(parents=True, exist_ok=True)
        (mem / "constitution.md").write_text(
            "# Constituição\n- **COD-02:** go vet\n- **DUP-10:** fonte única\n- **FLX-04:** telemetria\n",
            encoding="utf-8",
        )

    def tearDown(self):
        self.tmp.cleanup()

    def rodar(self, *args):
        out = io.StringIO()
        err = io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            rc = licoes.main(list(args))
        return rc, out.getvalue(), err.getvalue()

    def test_ler_rule_existente_sem_arquivo_imprime_id(self):
        rc, out, err = self.rodar("--dir", str(self.raiz), "--rule", "COD-02")
        self.assertEqual(rc, 0, err)
        self.assertIn("COD-02", out)

    def test_adicionar_e_ler_licao_por_rule(self):
        rc, out, err = self.rodar(
            "--dir",
            str(self.raiz),
            "--add",
            "--rule",
            "DUP-10",
            "--pattern",
            "duplicação entre prompts",
            "--fix",
            "extrair para skill única",
            "--ref",
            "fix-50",
        )
        self.assertEqual(rc, 0, err)
        arq = self.raiz / "sdd" / "memory" / "licoes" / "DUP-10.md"
        self.assertTrue(arq.is_file())

        rc2, out2, err2 = self.rodar("--dir", str(self.raiz), "--rule", "DUP-10")
        self.assertEqual(rc2, 0, err2)
        self.assertIn("DUP-10", out2)
        self.assertIn("duplicação entre prompts → extrair para skill única (fix-50)", out2)

    def test_adicionar_com_texto_posicional_e_multiplas_rules(self):
        rc, _, err = self.rodar(
            "--dir",
            str(self.raiz),
            "--add",
            "--rule",
            "COD-02,FLX-04",
            "padrão X → correção Y",
        )
        self.assertEqual(rc, 0, err)
        rc2, out2, _ = self.rodar("--dir", str(self.raiz), "--rule", "COD-02,FLX-04")
        self.assertEqual(rc2, 0)
        self.assertIn("COD-02", out2)
        self.assertIn("FLX-04", out2)
        self.assertIn("padrão X → correção Y", out2)

    def test_negativo_rule_inexistente_recusada(self):
        rc, _, err = self.rodar("--dir", str(self.raiz), "--rule", "ZZZ-99")
        self.assertNotEqual(rc, 0)
        self.assertIn("ZZZ-99", err)

    def test_negativo_path_traversal_e_controle_recusados(self):
        rc, _, _ = self.rodar("--dir", str(self.raiz) + "/../outro", "--rule", "COD-02")
        self.assertNotEqual(rc, 0)

        rc2, _, _ = self.rodar("--dir", str(self.raiz), "--rule", "../COD-02")
        self.assertNotEqual(rc2, 0)

        rc3, _, _ = self.rodar(
            "--dir",
            str(self.raiz),
            "--add",
            "--rule",
            "COD-02",
            "texto com \x1b[31m controle",
        )
        self.assertNotEqual(rc3, 0)


if __name__ == "__main__":
    unittest.main()
