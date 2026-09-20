package results

import (
	"fmt"
	"strings"
	"testing"
)

func TestServerContextBoundsAndCopiesAuthorizedDescriptors(t *testing.T) {
	ctx := visualContext("en", "uk", "pt", "fr", "de", "it", "es", "ja", "ko", "zh")
	ctx.Mode = ContextAssisted
	for i := 0; i < 100; i++ {
		ctx.Sources = append(ctx.Sources, Source{ID: fmt.Sprintf("%03d-", i) + strings.Repeat("x", 124), ContextExtent: true})
	}
	trusted, err := validateContext(ctx)
	if err != nil {
		t.Fatal("exact context bounds rejected", err)
	}
	firstID := ctx.Sources[0].ID
	ctx.Sources[0].ContextExtent = false
	ctx.Sources[0].ID = "mutated"
	ctx.Languages[0] = "ar"
	if !trusted.sources[firstID].ContextExtent || !trusted.languages["en"] {
		t.Fatal("trusted context aliases input")
	}
	for _, change := range []func(*Context){
		func(c *Context) { c.Languages = append(c.Languages, "ar") },
		func(c *Context) { c.Sources = append(c.Sources, Source{ID: "over-limit"}) },
		func(c *Context) { c.Sources[0].ID = strings.Repeat("x", 129) },
		func(c *Context) { c.Languages[0] = strings.Repeat("x", 129) },
		func(c *Context) { c.Sources[0].ID = string([]byte{0xff}) },
	} {
		copy := ctx
		copy.Languages = append([]string{}, ctx.Languages...)
		copy.Languages[0] = "en"
		copy.Sources = append([]Source{}, ctx.Sources...)
		change(&copy)
		if _, err := validateContext(copy); err == nil {
			t.Fatal("invalid bounded context accepted")
		}
	}
}
