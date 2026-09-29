package manuales

import "testing"

func TestGetManuales(t *testing.T) {
	d, ok := Get("manual-admin.pdf")
	if !ok || len(d) < 1_000_000 {
		t.Fatalf("manual-admin.pdf: ok=%v len=%d (esperado >1MB embebido)", ok, len(d))
	}
	d2, ok2 := Get("manual-psiuser.pdf")
	if !ok2 || len(d2) < 1_000_000 {
		t.Fatalf("manual-psiuser.pdf: ok=%v len=%d (esperado >1MB embebido)", ok2, len(d2))
	}
	if _, bad := Get("secret.pdf"); bad {
		t.Fatal("secret.pdf no debería existir en la whitelist")
	}
}