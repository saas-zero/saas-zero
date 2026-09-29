package envconf

import "testing"

func TestString_Fallback(t *testing.T) {
	if got := String("ENVCONF_TEST_UNSET_XYZ", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestString_Override(t *testing.T) {
	t.Setenv("ENVCONF_TEST_SET", "from-env")
	if got := String("ENVCONF_TEST_SET", "fallback"); got != "from-env" {
		t.Fatalf("expected from-env, got %s", got)
	}
}

func TestString_EmptyEnvUsesFallback(t *testing.T) {
	t.Setenv("ENVCONF_TEST_EMPTY", "")
	if got := String("ENVCONF_TEST_EMPTY", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback for empty env, got %s", got)
	}
}
