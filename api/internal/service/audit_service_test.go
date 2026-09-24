// api/internal/service/audit_service_test.go

// Tests unitarios del AuditService SIN arrancar el worker (mismo patrón que
// mail_service_resend_test.go): se inspecciona la cola directamente para no
// introducir carreras, y se vale de la nil-safety del registro global.
package service

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// TestAuditService_RecordEnqueuesSinWorker verifica que Record captura el evento
// en la cola sin necesidad del worker (escritura diferida).
func TestAuditService_RecordEnqueuesSinWorker(t *testing.T) {
	s := &AuditService{queue: make(chan AuditEvent, 4)}
	defer s.Close()

	evt := AuditEvent{
		Entity:        domain.AuditEntityPsi,
		EntityID:      uuid.New().String(),
		EntityLabel:   "Juan Pérez (FPV 12345)",
		Action:        domain.AuditActionUpdate,
		ActorID:       uuid.New(),
		ActorRole:     "sudo",
		ActorUsername: "admin",
		IP:            "127.0.0.1",
	}
	s.Record(evt)

	select {
	case got := <-s.queue:
		if got.Action != domain.AuditActionUpdate {
			t.Fatalf("acción encolada inesperada: %s", got.Action)
		}
		if got.ActorUsername != "admin" {
			t.Fatalf("actor encolado inesperado: %s", got.ActorUsername)
		}
	default:
		t.Fatal("el evento debió encolarse")
	}
}

// TestAuditService_RecordTrasCloseEsNoOp verifica que tras Close() la captura
// queda deshabilitada: el canal se cierra sin eventos pendientes (ok=false).
func TestAuditService_RecordTrasCloseEsNoOp(t *testing.T) {
	s := &AuditService{queue: make(chan AuditEvent, 4)}
	s.Close()

	// Tras Close, Record es un no-op: no debe encolar, ni panic.
	s.Record(AuditEvent{Entity: "psi", Action: "login"})

	// Leer de un canal cerrado y vacío devuelve ok=false → nada encolado.
	_, ok := <-s.queue
	if ok {
		t.Fatal("no debe aceptar eventos tras Close()")
	}
}

// TestAuditService_RecordDescartaSiColaLlena verifica el drop + warn sin bloquear.
func TestAuditService_RecordDescartaSiColaLlena(t *testing.T) {
	s := &AuditService{queue: make(chan AuditEvent, 1)}
	defer s.Close()

	s.Record(AuditEvent{Entity: "psi", Action: "a"})
	// El segundo llena el buffer; el tercero debe descartarse (no panic, no bloqueo).
	s.Record(AuditEvent{Entity: "psi", Action: "b"})
	s.Record(AuditEvent{Entity: "psi", Action: "c"})

	n := len(s.queue)
	if n != 1 {
		t.Fatalf("la cola con buffer 1 debía quedar con 1 evento, quedó %d", n)
	}
}

// TestAuditService_ToRowSerializaChangesYMetadata verifica la conversión del
// evento al modelo persistible (JSON por campo + metadata).
func TestAuditService_ToRowSerializaChangesYMetadata(t *testing.T) {
	evt := AuditEvent{
		Entity:   domain.AuditEntityNotification,
		EntityID: uuid.New().String(),
		Action:   domain.AuditActionCreate,
		Changes: map[string]domain.AuditChange{
			"status": {From: "pending", To: "sent"},
		},
		Metadata: map[string]any{"target_type": "global"},
	}

	row := evt.toRow()

	if row.ID == uuid.Nil {
		t.Fatal("s.toRow debe generar un ID v7")
	}
	if row.CreatedAt.IsZero() {
		t.Fatal("s.toRow debe estampar CreatedAt")
	}
	if len(row.Changes) == 0 {
		t.Fatal("changes no se serializaron")
	}
	var changes map[string]domain.AuditChange
	if err := json.Unmarshal(row.Changes, &changes); err != nil {
		t.Fatalf("changes no es JSON válido: %v", err)
	}
	chg, ok := changes["status"]
	if !ok || chg.From != "pending" || chg.To != "sent" {
		t.Fatalf("diff inesperado: %+v", changes)
	}
	if string(row.Metadata) == "" {
		t.Fatal("metadata no se serializó")
	}
}

// TestAuditService_BuildDiff verifica que el diff solo reporta lo que cambió
// (incluyendo campos presentes en un solo lado).
func TestAuditService_BuildDiff(t *testing.T) {
	before := map[string]any{"nombre": "Juan", "genero": "M", "solvent": false, "borrado": "x"}
	after := map[string]any{"nombre": "María", "genero": "M", "solvent": true, "nuevo": "v"}

	diff := BuildDiff(before, after)

	// nombre (cambió), solvent (cambió), borrado (solo before), nuevo (solo after).
	if len(diff) != 4 {
		t.Fatalf("se esperaban 4 cambios, hubo %d: %+v", len(diff), diff)
	}
	if d, ok := diff["nombre"]; !ok || d.From != "Juan" || d.To != "María" {
		t.Fatalf("diff de nombre incorrecto: %+v", diff["nombre"])
	}
	if _, ok := diff["genero"]; ok {
		t.Fatal("genero no cambió y no debe aparecer")
	}
	if d, ok := diff["solvent"]; !ok || d.From != false || d.To != true {
		t.Fatalf("diff de solvent incorrecto: %+v", diff["solvent"])
	}
	if d, ok := diff["borrado"]; !ok || d.From != "x" {
		t.Fatalf("campo borrado debe constar con from: %+v", diff["borrado"])
	}
	if d, ok := diff["nuevo"]; !ok || d.To != "v" {
		t.Fatalf("campo nuevo debe constar con to: %+v", diff["nuevo"])
	}
}

// TestAuditService_NilSafe verifica que Record sobre receptor nil no panic
// (defensa ante el registro global no inicializado).
func TestAuditService_NilSafe(t *testing.T) {
	var s *AuditService
	s.Record(AuditEvent{Entity: "psi", Action: "login"}) // no debe panic
}

// TestRecordAudit_SinInicializarEsNoOp verifica que el registro global no
// inicializado (nil) no reviente el flujo de la aplicación.
func TestRecordAudit_SinInicializarEsNoOp(t *testing.T) {
	// Asegurar estado limpio: sin InitAuditLogs en este test.
	InitAuditLogs(nil)
	RecordAudit(t.Context(), AuditEvent{Entity: "staff", Action: "create"}) // no debe panic
}

// TestPsiAuditLabel verifica el formato de la etiqueta de psicólogo.
func TestPsiAuditLabel(t *testing.T) {
	psi := &domain.PsiUserModel{FirstName: " Juan ", LastName: "Pérez", FPV: 12345}
	got := psiAuditLabel(psi)
	want := "Juan Pérez (FPV 12345)"
	if got != want {
		t.Fatalf("psiAuditLabel = %q, se esperaba %q", got, want)
	}
	if psiAuditLabel(nil) != "" {
		t.Fatal("psiAuditLabel(nil) debe ser vacío")
	}
}

// TestAuditRoleOfAdmin verifica la resolución del rol del actor admin.
func TestAuditRoleOfAdmin(t *testing.T) {
	sudo := &domain.UserAdmin{Sudo: true}
	secretaria := &domain.UserAdmin{Role: string(RoleSecretaria)}
	comun := &domain.UserAdmin{}

	if got := auditRoleOfAdmin(sudo); got != "sudo" {
		t.Fatalf("sudo = %q", got)
	}
	if got := auditRoleOfAdmin(secretaria); got != "secretaria" {
		t.Fatalf("secretaria = %q", got)
	}
	if got := auditRoleOfAdmin(comun); got != "staff" {
		t.Fatalf("comun = %q", got)
	}
	if auditRoleOfAdmin(nil) != "" {
		t.Fatal("nil debe dar vacío")
	}
}

// TestWithAuditMetaInyectaYExtrae verifica el transporte IP/UA en el contexto.
func TestWithAuditMetaInyectaYExtrae(t *testing.T) {
	ctx := WithAuditMeta(t.Context(), "10.0.0.1", "curl/8")
	meta := auditMetaFrom(ctx)
	if meta.IP != "10.0.0.1" || meta.UserAgent != "curl/8" {
		t.Fatalf("meta extraída incorrecta: %+v", meta)
	}
	if auditMetaFrom(nil).IP != "" {
		t.Fatal("auditMetaFrom(nil) debe ser vacío")
	}
}