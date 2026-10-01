package markdown

import (
	"strings"
	"testing"
)

func TestRenderAddsAnchorsAndLazyImages(t *testing.T) {
	out, err := Render("## Where it fits\n\n![diagram](/uploads/x.png)\n\n# Top\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<h2 id="where-it-fits">Where it fits<a href="#where-it-fits" class="anchor" title="Link to this section">#</a></h2>`,
		`loading="lazy"`,
		`decoding="async"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// h1 is the post title's level and must not get a self-link.
	if strings.Contains(out, `<h1 id="top">Top<a`) {
		t.Errorf("h1 should not get an anchor:\n%s", out)
	}
}
