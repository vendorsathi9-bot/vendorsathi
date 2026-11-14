package auth

import "testing"

func TestHashPasswordPHCVerify(t *testing.T) {
	params := DefaultPasswordParams()

	hash, err := HashPasswordPHC("correct horse battery staple", params)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("password was not hashed")
	}

	ok, needsRehash, err := VerifyPasswordWithParams("correct horse battery staple", hash, params)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
	if needsRehash {
		t.Fatal("hash created with current params should not need rehash")
	}

	ok, _, err = VerifyPasswordWithParams("wrong password", hash, params)
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}
	if ok {
		t.Fatal("wrong password must not verify")
	}
}

func TestVerifyPasswordWithParamsLegacyFormat(t *testing.T) {
	params := DefaultPasswordParams()
	legacy, err := HashPassword("legacy-secret", params.Memory, params.Iterations, params.Parallelism, params.SaltLength, params.KeyLength)
	if err != nil {
		t.Fatalf("legacy hash: %v", err)
	}

	ok, needsRehash, err := VerifyPasswordWithParams("legacy-secret", legacy, params)
	if err != nil {
		t.Fatalf("verify legacy: %v", err)
	}
	if !ok {
		t.Fatal("expected legacy password to verify")
	}
	if !needsRehash {
		t.Fatal("legacy hashes should be flagged for rehash")
	}
}

func TestVerifyPasswordWithParamsRejectsMalformed(t *testing.T) {
	params := DefaultPasswordParams()
	for _, encoded := range []string{
		"",
		"$argon2id$v=19$m=65536,t=3,p=4$not-base64!!!$also-bad!!!",
		"$argon2id$v=16$m=65536,t=3,p=4$c2FsdA$aGFzaA",
	} {
		if _, _, err := VerifyPasswordWithParams("secret", encoded, params); err == nil {
			t.Fatalf("expected malformed hash %q to be rejected", encoded)
		}
	}
}
