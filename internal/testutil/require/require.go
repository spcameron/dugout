package require

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/spcameron/dugout/internal/testutil"
)

func Equal[T any](t *testing.T, got, want T) {
	t.Helper()

	if !testutil.IsEqual(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}

func NotEqual[T any](t *testing.T, got, want T) {
	t.Helper()

	if testutil.IsEqual(got, want) {
		t.Fatalf("got: %v; expected values to be different", got)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()

	if !got {
		t.Fatalf("got: false; want: true")
	}
}

func False(t *testing.T, got bool) {
	t.Helper()

	if got {
		t.Fatalf("got: true; want: false")
	}
}

func Nil(t *testing.T, got any) {
	t.Helper()

	if !testutil.IsNil(got) {
		t.Fatalf("got: %v; want: nil", got)
	}
}

func NotNil(t *testing.T, got any) {
	t.Helper()

	if testutil.IsNil(got) {
		t.Fatalf("got: nil; want: non-nil")
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ErrorIs(t *testing.T, got, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}

func ErrorAs(t *testing.T, got error, target any) {
	t.Helper()

	if got == nil {
		t.Fatalf("got: nil; want assignable to: %T", target)
		return
	}

	if !errors.As(got, target) {
		t.Fatalf("got: %v; want assignable to: %T", got, target)
	}
}

func Contains(t *testing.T, got, substr string) {
	t.Helper()

	if !strings.Contains(got, substr) {
		t.Fatalf("got: %q; expected to contain: %q", got, substr)
	}
}

func MatchesRegexp(t *testing.T, got, pattern string) {
	t.Helper()

	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid regexp %q: %v", pattern, err)
	}
	if !re.MatchString(got) {
		t.Fatalf("got: %q; want to match: %q", got, pattern)
	}
}
