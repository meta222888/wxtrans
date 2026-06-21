package importer_test

import (
	"path/filepath"
	"testing"

	"wxtrans/internal/database"
	"wxtrans/internal/importer"
)

func TestImportWeChatExcel(t *testing.T) {
	matches, _ := filepath.Glob(filepath.Join("..", "..", "*.xlsx"))
	if len(matches) == 0 {
		t.Skip("no sample xlsx in project root")
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := importer.ImportWeChatExcel(db, matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported == 0 {
		t.Fatalf("expected imports, got %+v", result)
	}

	result2, err := importer.ImportWeChatExcel(db, matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if result2.Imported != 0 {
		t.Fatalf("expected no new imports on duplicate, got %+v", result2)
	}
	if result2.Skipped != result.Imported {
		t.Fatalf("expected skipped=%d, got %+v", result.Imported, result2)
	}

	count, err := db.CountAll()
	if err != nil {
		t.Fatal(err)
	}
	if count != result.Imported {
		t.Fatalf("count=%d imported=%d", count, result.Imported)
	}
}
