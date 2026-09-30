package identity

import "testing"

func TestSignVerify(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	sig, err := Sign(priv, []byte("abc"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	ok, err := Verify(pub, []byte("abc"), sig)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("expected valid signature")
	}
}
