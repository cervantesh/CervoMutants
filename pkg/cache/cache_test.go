package cache

import (
	"context"
	"path/filepath"
	"testing"

	"gitea.cervbox.synology.me/CervoSoft/cervo-mutant/pkg/engine"
)

func TestKeySeparatesParts(t *testing.T) {
	if Key("ab", "c") == Key("a", "bc") {
		t.Fatal("key should include part boundaries")
	}
	first := Key("same", "parts")
	second := Key("same", "parts")
	if first != second {
		t.Fatal("key should be deterministic")
	}
}

func TestStoreGetMissingAndPut(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "cache")}
	ctx := context.Background()
	if cached, ok, err := store.Get(ctx, "missing"); err != nil || ok {
		t.Fatalf("missing cache get = cached:%+v ok:%t err:%v", cached, ok, err)
	}
	result := engine.MutantResult{
		MutantID:    "m1",
		Status:      engine.StatusKilled,
		TestCommand: []string{"go", "test", "./..."},
		Mutant:      engine.Mutant{Fingerprint: "fingerprint"},
	}
	if err := store.Put(ctx, result); err != nil {
		t.Fatalf("Put returned error: %v", err)
	}
	key := Key("fingerprint", "go test ./...")
	cached, ok, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok || cached.Key != key || cached.Result.MutantID != "m1" {
		t.Fatalf("cached result mismatch: ok=%t cached=%+v", ok, cached)
	}
}
