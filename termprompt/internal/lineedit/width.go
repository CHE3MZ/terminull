package liner

import (
	"github.com/CHE3MZ/terminull/v2/termcore"
)

// countGlyphs returns the display width of s. Grapheme clusters such as
// combining sequences and emoji families count as one unit, matching how
// the terminal draws them.
func countGlyphs(s []rune) int {
	return termcore.Width(string(s))
}

func countMultiLineGlyphs(s []rune, columns int, start int) int {
	if columns <= 0 {
		columns = 1
	}

	n := start
	for text := string(s); len(text) > 0; {
		cluster, _ := termcore.FirstGrapheme(text)
		if cluster == "" {
			break
		}
		text = text[len(cluster):]

		switch width := termcore.Width(cluster); width {
		case 0:
		case 2:
			n += 2
			// No room for a wide glyph in the ending, so skip a column
			// and display it at the beginning.
			if n%columns == 1 {
				n++
			}
		default:
			n += width
		}
	}
	return n
}

func getPrefixGlyphs(s []rune, num int) []rune {
	text := string(s)
	p := 0
	for n := 0; n < num && p < len(text); {
		cluster, _ := termcore.FirstGrapheme(text[p:])
		if cluster == "" {
			break
		}
		p += len(cluster)
		if termcore.Width(cluster) > 0 {
			n++
		}
	}
	// Keep trailing zero-width clusters, as the old rune walk did.
	for p < len(text) {
		cluster, _ := termcore.FirstGrapheme(text[p:])
		if cluster == "" || termcore.Width(cluster) != 0 {
			break
		}
		p += len(cluster)
	}
	return []rune(text[:p])
}

func getSuffixGlyphs(s []rune, num int) []rune {
	text := string(s)
	starts := make([]int, 0, 8)
	ends := make([]int, 0, 8)
	for p := 0; p < len(text); {
		cluster, _ := termcore.FirstGrapheme(text[p:])
		if cluster == "" {
			break
		}
		starts = append(starts, p)
		p += len(cluster)
		ends = append(ends, p)
	}

	start := len(starts)
	for n := 0; start > 0 && n < num; {
		start--
		if termcore.Width(text[starts[start]:ends[start]]) > 0 {
			n++
		}
	}
	if start == len(starts) {
		return nil
	}
	return []rune(text[starts[start]:])
}
