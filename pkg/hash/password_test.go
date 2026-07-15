package hash

import "testing"

// A low cost keeps these tests fast; production cost comes from configuration.
const testCost = 4

func TestHasher_HashAndMatch(t *testing.T) {
	h := NewHasher(testCost)
	const password = "correct horse battery staple"

	hashed, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if string(hashed) == password {
		t.Fatal("hash must not equal the plaintext password")
	}

	ok, err := h.Matches(password, hashed)
	if err != nil {
		t.Fatalf("Matches: %v", err)
	}
	if !ok {
		t.Error("Matches = false for the correct password, want true")
	}
}

func TestHasher_MatchesRejectsWrongPassword(t *testing.T) {
	h := NewHasher(testCost)

	hashed, err := h.Hash("right-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := h.Matches("wrong-password", hashed)
	if err != nil {
		t.Fatalf("Matches: %v", err)
	}
	if ok {
		t.Error("Matches = true for a wrong password, want false")
	}
}

func TestHasher_SaltsProduceDistinctHashes(t *testing.T) {
	h := NewHasher(testCost)

	a, err := h.Hash("same-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	b, err := h.Hash("same-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if string(a) == string(b) {
		t.Error("bcrypt must salt each hash; identical hashes indicate a problem")
	}
}
