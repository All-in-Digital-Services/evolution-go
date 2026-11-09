package instance_repository

import (
	"os"
	"sort"
	"testing"

	instance_model "github.com/evolution-foundation/evolution-go/pkg/instance/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Roda contra um Postgres descartavel: EVOGO_TEST_POSTGRES_DSN aponta para um
// banco vazio que o teste pode recriar a tabela. Sem a variavel, o teste pula.
func TestInstanciasQueReligamNoBoot(t *testing.T) {
	dsn := os.Getenv("EVOGO_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("EVOGO_TEST_POSTGRES_DSN nao definido")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}
	if err := db.Migrator().DropTable(&instance_model.Instance{}); err != nil {
		t.Fatalf("limpar tabela: %v", err)
	}
	if err := db.AutoMigrate(&instance_model.Instance{}); err != nil {
		t.Fatalf("criar tabela: %v", err)
	}

	instancias := []instance_model.Instance{
		{Id: "00000000-0000-0000-0000-000000000001", Name: "conectada", Token: "t1", Connected: true, Jid: "5527@s.whatsapp.net", ClientName: "c"},
		{Id: "00000000-0000-0000-0000-000000000002", Name: "reconectando", Token: "t2", Jid: "5528@s.whatsapp.net", DisconnectReason: "Reconnecting", ClientName: "c"},
		{Id: "00000000-0000-0000-0000-000000000003", Name: "caiu-websocket", Token: "t3", Jid: "5529@s.whatsapp.net", DisconnectReason: "Disconnected emitted because the websocket is closed by the server.", ClientName: "c"},
		{Id: "00000000-0000-0000-0000-000000000004", Name: "logout", Token: "t4", Jid: "5530@s.whatsapp.net", DisconnectReason: "LoggedOut", ClientName: "c"},
		{Id: "00000000-0000-0000-0000-000000000005", Name: "caiu-sem-pareamento", Token: "t5", DisconnectReason: "Reconnecting", ClientName: "c"},
		{Id: "00000000-0000-0000-0000-000000000006", Name: "nunca-conectou", Token: "t6", ClientName: "c"},
		{Id: "00000000-0000-0000-0000-000000000007", Name: "outro-cliente", Token: "t7", Jid: "5531@s.whatsapp.net", DisconnectReason: "Reconnecting", ClientName: "outro"},
	}
	if err := db.Create(&instancias).Error; err != nil {
		t.Fatalf("semear: %v", err)
	}

	repo := NewInstanceRepository(db)
	nomes := func(lista []*instance_model.Instance) []string {
		out := make([]string, 0, len(lista))
		for _, i := range lista {
			out = append(out, i.Name)
		}
		sort.Strings(out)
		return out
	}

	doCliente, err := repo.GetAllConnectedInstancesByClientName("c")
	if err != nil {
		t.Fatalf("consulta por cliente: %v", err)
	}
	esperado := []string{"caiu-websocket", "conectada", "reconectando"}
	if got := nomes(doCliente); !equal(got, esperado) {
		t.Fatalf("por cliente: esperado %v, veio %v", esperado, got)
	}

	todas, err := repo.GetAllConnectedInstances()
	if err != nil {
		t.Fatalf("consulta geral: %v", err)
	}
	esperado = []string{"caiu-websocket", "conectada", "outro-cliente", "reconectando"}
	if got := nomes(todas); !equal(got, esperado) {
		t.Fatalf("geral: esperado %v, veio %v", esperado, got)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
