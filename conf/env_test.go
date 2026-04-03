package conf

import "testing"

func TestGetAppEnvDefaultsToTest(t *testing.T) {
	t.Setenv("APP_ENV", "")

	if GetAppEnv() != EnvTest {
		t.Fatalf("expected default env %q, got %q", EnvTest, GetAppEnv())
	}
}

func TestGetTargetURLUsesProdURL(t *testing.T) {
	t.Setenv("APP_ENV", EnvProd)
	t.Setenv("PROD_API_URL", "https://prod.ganzam.local")

	if GetTargetURL() != "https://prod.ganzam.local" {
		t.Fatalf("expected prod target url, got %q", GetTargetURL())
	}
}

func TestGetTargetURLUsesTestURL(t *testing.T) {
	t.Setenv("APP_ENV", EnvTest)
	t.Setenv("TEST_API_URL", "https://test.ganzam.local")

	if GetTargetURL() != "https://test.ganzam.local" {
		t.Fatalf("expected test target url, got %q", GetTargetURL())
	}
}
